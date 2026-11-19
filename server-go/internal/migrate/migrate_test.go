package migrate

import (
	"testing"
)

func TestVersionOrder(t *testing.T) {
	for _, pair := range [][2]string{{"1", "1.1"}, {"1.1", "1.2"}, {"1.2", "1.10"}, {"1.10", "2"}} {
		if compareVersion(pair[0], pair[1]) >= 0 {
			t.Fatalf("版本顺序错误：%v", pair)
		}
	}
	if compareVersion("1.0", "1") != 0 {
		t.Fatal("尾部零版本应等价")
	}
}

func TestFlywayChecksum(t *testing.T) {
	// IEEE CRC32 的标准样例 123456789 对应有符号值 -873187034。
	for _, input := range []string{"123456789", "123\n456\r\n789\n", "\ufeff123\r456\r789"} {
		if got := checksum([]byte(input)); got != -873187034 {
			t.Fatalf("校验和 = %d", got)
		}
	}
	scripts, err := embeddedScripts()
	if err != nil || len(scripts) < 3 || scripts[0].version != "1" {
		t.Fatalf("内嵌迁移 = %d，错误 = %v", len(scripts), err)
	}
	// 对照 Java Flyway 在 MySQL 5.7 实际写入的历史记录。
	known := map[string]int32{"1": -1305210990, "1.1": -1832970016, "1.2": -844713052}
	for _, script := range scripts {
		if expected, ok := known[script.version]; ok && script.checksum != expected {
			t.Errorf("版本 %s 的校验和 = %d，Flyway 记录 = %d", script.version, script.checksum, expected)
		}
	}
}
