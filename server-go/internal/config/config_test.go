package config

import (
	"strings"
	"testing"
)

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
