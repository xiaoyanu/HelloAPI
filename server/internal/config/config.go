package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-sql-driver/mysql"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int `yaml:"port"`
	} `yaml:"server"`
	Database Database `yaml:"database"`
	JWT      struct {
		Secret string `yaml:"secret"`
	} `yaml:"jwt"`
	Stat struct {
		InitialDelayMS int `yaml:"initial-delay-ms"`
		FixedDelayMS   int `yaml:"fixed-delay-ms"`
	} `yaml:"stat"`
}

type Database struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (d Database) MySQL() *mysql.Config {
	cfg := mysql.NewConfig()
	cfg.User = d.Username
	cfg.Passwd = d.Password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(d.Host, strconv.Itoa(d.Port))
	cfg.DBName = d.Name
	// Charset 让驱动在握手后发送 SET NAMES；此选项不会返回错误。
	_ = cfg.Apply(mysql.Charset("utf8mb4", ""))
	cfg.ParseTime = true
	cfg.Loc = time.Local
	// MySQL Connector/J 默认返回匹配行数，包括值未变化的 UPDATE。
	cfg.ClientFoundRows = true
	return cfg
}

func Load(explicitPath string) (Config, error) {
	var cfg Config
	cfg.Server.Port = 8080
	cfg.Database = Database{Host: "localhost", Port: 3306, Name: "api", Username: "root", Password: "123456"}
	cfg.Stat.InitialDelayMS = 10000
	cfg.Stat.FixedDelayMS = 10000

	var paths []string
	if explicitPath != "" {
		paths = append(paths, explicitPath)
	}
	if executable, err := os.Executable(); err == nil {
		paths = append(paths, filepath.Join(filepath.Dir(executable), "config.yaml"))
	}
	paths = append(paths, "config.yaml")
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return cfg, fmt.Errorf("读取配置 %s: %w", path, err)
		}
		if err := yaml.Unmarshal(raw, &cfg); err != nil {
			return cfg, fmt.Errorf("解析配置 %s: %w", path, err)
		}
		break
	}

	for name, target := range map[string]*string{
		"HELLOAPI_DB_HOST":     &cfg.Database.Host,
		"HELLOAPI_DB_NAME":     &cfg.Database.Name,
		"HELLOAPI_DB_USERNAME": &cfg.Database.Username,
		"HELLOAPI_DB_PASSWORD": &cfg.Database.Password,
		"HELLOAPI_JWT_SECRET":  &cfg.JWT.Secret,
	} {
		if value, ok := os.LookupEnv(name); ok {
			*target = value
		}
	}
	for name, target := range map[string]*int{
		"HELLOAPI_PORT":                  &cfg.Server.Port,
		"HELLOAPI_DB_PORT":               &cfg.Database.Port,
		"HELLOAPI_STAT_INITIAL_DELAY_MS": &cfg.Stat.InitialDelayMS,
		"HELLOAPI_STAT_FIXED_DELAY_MS":   &cfg.Stat.FixedDelayMS,
	} {
		if raw, ok := os.LookupEnv(name); ok {
			value, err := strconv.Atoi(raw)
			if err != nil {
				return cfg, fmt.Errorf("环境变量 %s 必须为整数", name)
			}
			*target = value
		}
	}
	if cfg.Server.Port < 1 || cfg.Server.Port > 65535 || cfg.Database.Port < 1 || cfg.Database.Port > 65535 {
		return cfg, errors.New("HTTP 和数据库端口必须处于 1 到 65535")
	}
	const maxDelayMS = int64((1<<63 - 1) / time.Millisecond)
	if cfg.Stat.InitialDelayMS < 0 || int64(cfg.Stat.InitialDelayMS) > maxDelayMS || cfg.Stat.FixedDelayMS < 1 || int64(cfg.Stat.FixedDelayMS) > maxDelayMS {
		return cfg, errors.New("统计调度初始延迟必须非负，固定间隔必须为正，且两者须在可表示范围内")
	}
	return cfg, nil
}
