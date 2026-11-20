package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/zxz/helloapi/internal/auth"
	"github.com/zxz/helloapi/internal/httpx"
)

type adminLookup func(context.Context, int) (int64, error)

func (a adminLookup) CheckUserIsAdmin(ctx context.Context, id int) (int64, error) { return a(ctx, id) }

func TestRouterRegistersExactAliases(t *testing.T) {
	router := New(nil, nil, nil, nil, nil, nil, nil, auth.New("test"), nil).(chi.Routes)
	var total int
	if err := chi.Walk(router, func(method, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		total++
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if total != 57 {
		t.Errorf("路由数 = %d，期望控制器 56 条加静态资源 1 条", total)
	}
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/v1/user/Captcha"},
		{"GET", "/api/v1/user/AppList/Search"},
		{"GET", "/api/v1/app/"},
		{"GET", "/api/v1/app/key/list/"},
		{"GET", "/api/v1/app/key/list/{user_id}/Search"},
		{"PUT", "/api/v1/user/key/"},
		{"PUT", "/api/v1/user/key/{userId}"},
		{"PUT", "/api/v1/user/key"},
		{"DELETE", "/api/v1/user/"},
		{"POST", "/api/v1/stat/log"},
		{"POST", "/api/v1/stat/log/"},
		{"GET", "/Logo.psd"},
	} {
		if !router.Match(chi.NewRouteContext(), route.method, route.path) {
			t.Errorf("缺少 %s %s", route.method, route.path)
		}
	}
	for _, path := range []string{"/api/v1/user", "/api/v1/user/captcha", "/api/v1/user/list/Search/"} {
		if router.Match(chi.NewRouteContext(), http.MethodGet, path) {
			t.Errorf("错误匹配 GET %s", path)
		}
	}
}

func TestAuthRequiredBodyAndPreflight(t *testing.T) {
	jwt := auth.New("test")
	var checks int
	admins := adminLookup(func(_ context.Context, _ int) (int64, error) {
		checks++
		return 0, nil
	})
	router := New(nil, nil, nil, nil, nil, nil, nil, jwt, admins)
	token, err := jwt.Issue(4, "test", 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		method, path, body, token, want string
	}{
		{"GET", "/api/v1/app/key/list/", "", "", `{"code":401,"msg":"Token过期/无效"}`},
		{"POST", "/api/v1/app/", "{", token, `{"code":400,"msg":"参数错误"}`},
		{"PUT", "/api/v1/app/key/xyz", "null", token, `{"code":400,"msg":"参数错误"}`},
		{"GET", "/api/v1/user/list", "", token, `{"code":403,"msg":"非管理员用户，没有操作权限"}`},
	} {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		if tc.token != "" {
			req.Header.Set("Authorization", "Bearer "+tc.token)
		}
		res := httptest.NewRecorder()
		router.ServeHTTP(res, req)
		if res.Code != 200 || res.Body.String() != tc.want {
			t.Errorf("%s %s 返回 HTTP %d，%s", tc.method, tc.path, res.Code, res.Body.String())
		}
	}
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/user/list", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Authorization, Content-Type")
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	for key, value := range map[string]string{
		"Access-Control-Allow-Origin":      "https://example.com",
		"Access-Control-Allow-Credentials": "true",
		"Access-Control-Allow-Methods":     "GET",
		"Access-Control-Allow-Headers":     "Authorization, Content-Type",
		"Access-Control-Max-Age":           "3600",
	} {
		if res.Header().Get(key) != value {
			t.Errorf("预检 %s = %q", key, res.Header().Get(key))
		}
	}
	if res.Code != 204 || checks != 1 {
		t.Errorf("预检 HTTP %d，管理员查询 %d 次", res.Code, checks)
	}
}

func TestAdminUsesLivePermissionAndPromotesOrdinaryToken(t *testing.T) {
	jwt := auth.New("test")
	token, err := jwt.Issue(4, "test", 0)
	if err != nil {
		t.Fatal(err)
	}
	handler := httpx.RequireAdmin(jwt, adminLookup(func(_ context.Context, id int) (int64, error) {
		if id != 4 {
			t.Errorf("查库用户 ID = %d", id)
		}
		return 1, nil
	}))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if identity := httpx.User(r.Context()); identity.ID != 4 || identity.Mode != 1 {
			t.Errorf("请求上下文身份 = %+v", identity)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != 204 {
		t.Errorf("管理员请求 HTTP %d", res.Code)
	}
}

func TestLogoSupportsHEAD(t *testing.T) {
	route := New(nil, nil, nil, nil, nil, nil, nil, auth.New("test"), nil)
	req := httptest.NewRequest(http.MethodHead, "/Logo.psd", nil)
	res := httptest.NewRecorder()
	route.ServeHTTP(res, req)
	if res.Code != 200 || res.Header().Get("Content-Type") != "image/vnd.adobe.photoshop" ||
		res.Header().Get("Content-Length") != "458331" || res.Body.Len() != 0 {
		t.Fatalf("静态资源 HEAD 返回 HTTP %d，响应头 %v，响应体 %d 字节", res.Code, res.Header(), res.Body.Len())
	}
}
