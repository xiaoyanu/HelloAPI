package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestRecoverReturnsJSONWithCORSHeaders(t *testing.T) {
	handler := CORS(Recover(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("模拟请求异常")
	})))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != 500 || res.Body.String() != `{"code":500,"msg":"服务器内部错误"}` || res.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("异常响应 HTTP %d，响应体 %s，响应头 %v", res.Code, res.Body.String(), res.Header())
	}
}

func TestStringMap(t *testing.T) {
	got := StringMap(strings.NewReader(`{"number":1,"decimal":1.25,"large":9007199254740993,"flag":true,"none":null,"empty":"","text":"原文"}`))
	want := map[string]string{"number": "1", "decimal": "1.25", "large": "9007199254740993", "flag": "true", "empty": "", "text": "原文"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("结果 = %#v", got)
	}
	for _, raw := range []string{"", "{", "[]", "null", "{} {}", `{"a":[],"b":1}`} {
		if got := StringMap(strings.NewReader(raw)); len(got) != 0 {
			t.Fatalf("无效请求 %q 解码为 %v", raw, got)
		}
	}
	req := httptest.NewRequest("GET", "/?page=1&page=2", nil)
	if got := Query(req)["page"]; got != "1" {
		t.Fatalf("首个值 = %q", got)
	}
}

func TestResponseEmptySlice(t *testing.T) {
	for _, tc := range []struct {
		data any
		want string
	}{
		{nil, `{"code":200,"msg":"成功"}`},
		{([]string)(nil), `{"code":200,"msg":"成功","data":[]}`},
		{[]string{}, `{"code":200,"msg":"成功","data":[]}`},
		{(*int)(nil), `{"code":200,"msg":"成功"}`},
	} {
		got, err := json.Marshal(Success(tc.data))
		if err != nil || string(got) != tc.want {
			t.Fatalf("响应 = %s，错误 = %v", got, err)
		}
	}
}
