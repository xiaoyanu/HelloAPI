package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestLocalDateTime(t *testing.T) {
	for _, input := range []string{"2026-11-19T08:15:02", "2026-11-19 08:15:02", "2026-11-19T08:15:02.123"} {
		var got LocalDateTime
		if err := json.Unmarshal([]byte(`"`+input+`"`), &got); err != nil {
			t.Fatal(err)
		}
		encoded, err := json.Marshal(got)
		if err != nil || string(encoded) != `"2026-11-19T08:15:02"` {
			t.Fatalf("编码 = %s，错误 = %v", encoded, err)
		}
		value, err := got.Value()
		if err != nil {
			t.Fatal(err)
		}
		var scanned LocalDateTime
		if err := scanned.Scan(value); err != nil || !scanned.Equal(got.Time) {
			t.Fatalf("数据库往返 = %v，错误 = %v", scanned, err)
		}
		if err := scanned.Scan([]byte("2026-11-19 08:15:02")); err != nil || scanned.Hour() != 8 {
			t.Fatalf("扫描字节 = %v，错误 = %v", scanned, err)
		}
	}
	var got LocalDateTime
	if err := json.Unmarshal([]byte(`"2026-11-19T08:15:02Z"`), &got); err == nil {
		t.Fatal("接受了带时区的输入")
	}
	if err := got.Scan(time.Date(2026, 11, 19, 8, 15, 2, 0, time.UTC)); err != nil || got.Hour() != 8 {
		t.Fatalf("扫描时间 = %v，错误 = %v", got, err)
	}
	var nullable *LocalDateTime
	if err := json.Unmarshal([]byte("null"), &nullable); err != nil || nullable != nil {
		t.Fatalf("空值 = %v，错误 = %v", nullable, err)
	}
	date, err := json.Marshal(Date{got.Time})
	if err != nil || string(date) != `"2026-11-19"` {
		t.Fatalf("日期 = %s，错误 = %v", date, err)
	}
}
