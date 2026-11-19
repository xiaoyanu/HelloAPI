package httpx

import (
	"context"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/zxz/helloapi/internal/auth"
)

type identityKey struct{}

type Identity struct{ ID, Mode int }

func User(ctx context.Context) Identity {
	identity, _ := ctx.Value(identityKey{}).(Identity)
	return identity
}

func RequireLogin(jwt *auth.JWT) func(http.Handler) http.Handler {
	return require(jwt, nil)
}

type AdminChecker interface {
	CheckUserIsAdmin(context.Context, int) (int64, error)
}

func RequireAdmin(jwt *auth.JWT, users AdminChecker) func(http.Handler) http.Handler {
	return require(jwt, users)
}

func require(jwt *auth.JWT, users AdminChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := jwt.Bearer(r.Header.Get("Authorization"))
			if !ok || claims.ID < 1 {
				Write(w, http.StatusOK, Reply(401, MsgTokenInvalid))
				return
			}
			if users != nil {
				count, err := users.CheckUserIsAdmin(r.Context(), claims.ID)
				if err != nil {
					slog.ErrorContext(r.Context(), "查询管理员权限失败", "错误", err)
					Write(w, http.StatusInternalServerError, Reply(500, MsgInternal))
					return
				}
				if count < 1 {
					Write(w, http.StatusOK, Reply(403, MsgNotAdmin))
					return
				}
				if claims.Mode == 0 {
					claims.Mode = Admin
				}
			}
			ctx := context.WithValue(r.Context(), identityKey{}, Identity{ID: claims.ID, Mode: claims.Mode})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Add("Vary", "Origin")
		h.Add("Vary", "Access-Control-Request-Method")
		h.Add("Vary", "Access-Control-Request-Headers")
		if origin := r.Header.Get("Origin"); origin != "" {
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			if method := r.Header.Get("Access-Control-Request-Method"); r.Method == http.MethodOptions && method != "" {
				h.Set("Access-Control-Allow-Methods", method)
				if headers := strings.Join(r.Header.Values("Access-Control-Request-Headers"), ", "); headers != "" {
					h.Set("Access-Control-Allow-Headers", headers)
				}
				h.Set("Access-Control-Max-Age", "3600")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				slog.ErrorContext(r.Context(), "处理请求时发生异常", "错误", recovered, "堆栈", string(debug.Stack()))
				Write(w, http.StatusInternalServerError, Reply(500, MsgInternal))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
