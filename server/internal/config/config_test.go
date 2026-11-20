package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatLogConfigDefaults(t *testing.T) {
	cfg, err := Load(writeTestConfig(t, "{}"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Stat.LogRetentionDays != 90 || cfg.Stat.LogCleanupIntervalMS != 3600000 || cfg.Stat.LogPayloadMaxBytes != 2048 {
		t.Fatalf("日志默认配置 = %+v", cfg.Stat)
	}
}

func TestStatLogConfigYAMLAndEnvironment(t *testing.T) {
	path := writeTestConfig(t, "stat:\n  log-retention-days: 0\n  log-cleanup-interval-ms: 3000\n  log-payload-max-bytes: 0\n")
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Stat.LogRetentionDays != 0 || cfg.Stat.LogCleanupIntervalMS != 3000 || cfg.Stat.LogPayloadMaxBytes != 0 {
		t.Fatalf("YAML 日志配置 = %+v", cfg.Stat)
	}
	t.Setenv("HELLOAPI_STAT_LOG_RETENTION_DAYS", "30")
	t.Setenv("HELLOAPI_STAT_LOG_CLEANUP_INTERVAL_MS", "60000")
	t.Setenv("HELLOAPI_STAT_LOG_PAYLOAD_MAX_BYTES", "-1")
	cfg, err = Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Stat.LogRetentionDays != 30 || cfg.Stat.LogCleanupIntervalMS != 60000 || cfg.Stat.LogPayloadMaxBytes != -1 {
		t.Fatalf("环境变量覆盖后的日志配置 = %+v", cfg.Stat)
	}
}

func TestStatLogConfigRejectsInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		field string
		value string
	}{
		{"log-retention-days", "-1"},
		{"log-cleanup-interval-ms", "0"},
		{"log-cleanup-interval-ms", "-1"},
		{"log-payload-max-bytes", "-2"},
	} {
		t.Run(tc.field+"="+tc.value, func(t *testing.T) {
			_, err := Load(writeTestConfig(t, "stat:\n  "+tc.field+": "+tc.value+"\n"))
			if err == nil || !strings.Contains(err.Error(), tc.field) {
				t.Fatalf("配置错误 = %v，期望包含 %s", err, tc.field)
			}
		})
	}
}

func writeTestConfig(t *testing.T, raw string) string {
	t.Helper()
	// 清除继承的配置覆盖项，避免测试结果依赖启动服务时使用的环境变量。
	for _, name := range []string{
		"HELLOAPI_PORT", "HELLOAPI_DB_PORT", "HELLOAPI_STAT_INITIAL_DELAY_MS", "HELLOAPI_STAT_FIXED_DELAY_MS",
		"HELLOAPI_STAT_LOG_RETENTION_DAYS", "HELLOAPI_STAT_LOG_CLEANUP_INTERVAL_MS", "HELLOAPI_STAT_LOG_PAYLOAD_MAX_BYTES",
	} {
		t.Setenv(name, "")
		if err := os.Unsetenv(name); err != nil {
			t.Fatal(err)
		}
	}
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(raw), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestMySQLConnectionOptions(t *testing.T) {
	db := Database{Host: "127.0.0.1", Port: 3307, Name: "test", Username: "tester"}
	cfg := db.MySQL()
	if _, ok := cfg.Params["charset"]; ok {
		t.Fatal("不能把 charset 作为 MySQL 系统变量发送")
	}
	dsn := cfg.FormatDSN()
	for _, option := range []string{"charset=utf8mb4", "parseTime=true", "loc=Local", "clientFoundRows=true"} {
		if !strings.Contains(dsn, option) {
			t.Errorf("DSN 缺少 %s", option)
		}
	}
	if cfg.MultiStatements {
		t.Fatal("业务连接开启了多语句")
	}
}
