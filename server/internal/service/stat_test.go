package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

type statLogFake struct {
	StatRepository
	saved        bool
	header, body *string
}

func (*statLogFake) CheckUserKeyExists(context.Context, *string) (int64, error) { return 1, nil }
func (*statLogFake) GetUserIDByUserKey(context.Context, *string) (int, error)   { return 1, nil }
func (*statLogFake) CheckAPIExist(context.Context, int) (int64, error)          { return 1, nil }
func (*statLogFake) CheckAPIViewStatus(context.Context, int) (int64, error)     { return 1, nil }
func (f *statLogFake) SaveAPILog(_ context.Context, _ int, _, header, body, _ *string, _ int) error {
	f.saved, f.header, f.body = true, header, body
	return nil
}

func TestLogAPIPayloadLimits(t *testing.T) {
	const small = `{"a":1}`
	const multibyte = `"你好"`
	large := `"` + strings.Repeat("a", 2048) + `"`
	for _, tc := range []struct {
		name         string
		limit        int
		header, body string
		wantHeader   string
		wantBody     string
	}{
		{"字段独立按字节限额", len(small), small, multibyte, small, `{"truncated":true,"size":8}`},
		{"默认大小的占位", 2048, large, small, `{"truncated":true,"size":2050}`, small},
		{"关闭保存", 0, small, multibyte, "", ""},
		{"不限大小", -1, large, small, large, small},
		{"非法 JSON", 1, "{invalid}", "", "", ""},
		{"不限大小仍校验 JSON", -1, "{invalid}", "", "", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &statLogFake{}
			result, err := NewStat(fake, tc.limit).LogAPI(t.Context(), map[string]string{
				"user_key": "key", "user_id": "1", "api_id": "1", "header": tc.header, "body": tc.body,
			})
			if err != nil || result.Code != 200 || !fake.saved {
				t.Fatalf("上报结果 = %+v，错误 = %v，保存 = %v", result, err, fake.saved)
			}
			for _, field := range []struct {
				name string
				got  *string
				want string
			}{{"header", fake.header, tc.wantHeader}, {"body", fake.body, tc.wantBody}} {
				if field.want == "" {
					if field.got != nil {
						t.Errorf("%s = %q，期望 NULL", field.name, *field.got)
					}
					continue
				}
				if field.got == nil || *field.got != field.want || !json.Valid([]byte(*field.got)) {
					t.Errorf("%s 保存值不符，期望合法 JSON：%s", field.name, field.want)
				}
			}
		})
	}
}
