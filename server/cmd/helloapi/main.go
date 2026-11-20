package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/zxz/helloapi/internal/auth"
	"github.com/zxz/helloapi/internal/captcha"
	"github.com/zxz/helloapi/internal/config"
	"github.com/zxz/helloapi/internal/handler"
	"github.com/zxz/helloapi/internal/migrate"
	"github.com/zxz/helloapi/internal/repo"
	"github.com/zxz/helloapi/internal/service"
	"github.com/zxz/helloapi/internal/stat"
)

func main() {
	configPath := flag.String("config", "", "配置文件路径")
	flag.Parse()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, *configPath, logger); err != nil {
		logger.Error("服务退出", "错误", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, configPath string, logger *slog.Logger) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}
	mysqlConfig := cfg.Database.MySQL()
	connector, err := mysql.NewConnector(mysqlConfig)
	if err != nil {
		return err
	}
	db := sqlx.NewDb(sql.OpenDB(connector), "mysql")
	defer db.Close()

	pingCtx, cancelPing := context.WithTimeout(ctx, 10*time.Second)
	err = db.PingContext(pingCtx)
	cancelPing()
	if err != nil {
		return fmt.Errorf("连接 MySQL: %w", err)
	}
	if err := migrate.Run(ctx, mysqlConfig, logger); err != nil {
		return err
	}

	store := repo.New(db)
	jwt := auth.New(cfg.JWT.Secret)
	sessions := captcha.New()
	router := handler.New(service.NewUser(store, jwt), service.NewAPI(store, jwt), service.NewLink(store),
		service.NewSetting(store), service.NewStat(store), service.NewDashboard(store), sessions, jwt, store)
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.Server.Port), Handler: router,
		ReadHeaderTimeout: 5 * time.Second, IdleTimeout: 60 * time.Second,
	}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("监听 HTTP 端口: %w", err)
	}

	workCtx, stopWorkers := context.WithCancel(ctx)
	var workers sync.WaitGroup
	workers.Go(func() { sessions.Cleanup(workCtx) })
	workers.Go(func() {
		stat.Run(workCtx, service.NewAggregation(store),
			time.Duration(cfg.Stat.InitialDelayMS)*time.Millisecond, time.Duration(cfg.Stat.FixedDelayMS)*time.Millisecond, logger)
	})
	defer func() {
		stopWorkers()
		workers.Wait()
	}()

	listening := make(chan error, 1)
	go func() { listening <- server.Serve(listener) }()
	logger.Info("HTTP 服务启动", "监听地址", server.Addr)
	select {
	case err := <-listening:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("启动 HTTP 服务: %w", err)
	case <-ctx.Done():
		stopWorkers()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return errors.Join(err, server.Close())
		}
		if err := <-listening; !errors.Is(err, http.ErrServerClosed) {
			return err
		}
		logger.Info("HTTP 服务已停止")
		return nil
	}
}
