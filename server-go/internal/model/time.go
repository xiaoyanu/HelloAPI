package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const dateTimeLayout = "2006-01-02T15:04:05"

// LocalDateTime 表示没有时区的数据库日期时间。
type LocalDateTime struct{ time.Time }

func ParseLocalDateTime(s string) (LocalDateTime, error) {
	// Java 也接受省略秒和带小数秒的 ISO 本地时间。
	if len(s) == 16 && s[10] == 'T' {
		s += ":00"
	}
	s = strings.Replace(s, " ", "T", 1)
	value, err := time.ParseInLocation(dateTimeLayout, s, time.Local)
	return LocalDateTime{value}, err
}

func (t LocalDateTime) MarshalJSON() ([]byte, error) {
	if t.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Format(dateTimeLayout))
}

func (t *LocalDateTime) UnmarshalJSON(raw []byte) error {
	if string(raw) == "null" {
		*t = LocalDateTime{}
		return nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return err
	}
	return t.Scan(value)
}

func (t *LocalDateTime) Scan(value any) error {
	switch v := value.(type) {
	case time.Time:
		*t = LocalDateTime{v}
	case string:
		parsed, err := ParseLocalDateTime(v)
		if err != nil {
			return err
		}
		*t = parsed
	case []byte:
		return t.Scan(string(v))
	case nil:
		*t = LocalDateTime{}
	default:
		return fmt.Errorf("无法将 %T 扫描为日期时间", value)
	}
	return nil
}

func (t LocalDateTime) Value() (driver.Value, error) {
	if t.IsZero() {
		return nil, nil
	}
	// 以墙上时间写入 DATETIME，避免驱动按连接时区转换。
	return t.Format("2006-01-02 15:04:05.999999"), nil
}

type Date struct{ time.Time }

func (d Date) MarshalJSON() ([]byte, error) { return json.Marshal(d.Format(time.DateOnly)) }
