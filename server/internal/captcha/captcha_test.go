package captcha

import (
	"bytes"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCaptchaSessionAndConsumption(t *testing.T) {
	sessions := New()
	req := httptest.NewRequest(http.MethodGet, "/Captcha", nil)
	res := httptest.NewRecorder()
	if err := sessions.Issue(res, req); err != nil {
		t.Fatal(err)
	}
	cookies := res.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("会话 Cookie = %v", cookies)
	}
	cookie := cookies[0]
	if cookie.Name != "JSESSIONID" || cookie.Path != "/" || !cookie.HttpOnly || cookie.SameSite != 0 {
		t.Fatalf("会话 Cookie 属性 = %+v", cookie)
	}
	if res.Header().Get("Content-Type") != "image/png" || !strings.Contains(res.Header().Get("Cache-Control"), "no-cache") {
		t.Fatalf("验证码响应头 = %v", res.Header())
	}
	image, err := png.DecodeConfig(bytes.NewReader(res.Body.Bytes()))
	if err != nil || image.Width != 120 || image.Height != 32 {
		t.Fatalf("验证码尺寸 = %+v，错误 = %v", image, err)
	}

	request := httptest.NewRequest(http.MethodPost, "/Register", nil)
	request.AddCookie(cookie)
	code := strings.ToLower(sessions.entries[cookie.Value].code)
	attempt := httptest.NewRecorder()
	if sessions.Verify(attempt, request, "!!!!", false) || !sessions.Verify(attempt, request, code, false) || sessions.Verify(attempt, request, code, false) {
		t.Fatal("注册验证码应允许输错后重试，成功后失效")
	}

	if err := sessions.Issue(httptest.NewRecorder(), request); err != nil {
		t.Fatal(err)
	}
	code = sessions.entries[cookie.Value].code
	if sessions.Verify(attempt, request, "!!!!", true) || sessions.Verify(attempt, request, code, true) {
		t.Fatal("登录验证码应在输错后立即失效")
	}

	sessions.entries[cookie.Value] = entry{code: "AAAA", expires: time.Now().Add(-time.Minute)}
	if sessions.Verify(attempt, request, "AAAA", true) {
		t.Fatal("过期验证码仍有效")
	}
	newCookies := attempt.Result().Cookies()
	if len(newCookies) != 1 || newCookies[0].Value == cookie.Value {
		t.Fatalf("过期后验证应创建新会话，Cookie = %v", newCookies)
	}
}
