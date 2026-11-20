package handler

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/zxz/helloapi/internal/auth"
	"github.com/zxz/helloapi/internal/captcha"
	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/service"
)

type Server struct {
	user      *service.User
	api       *service.API
	link      *service.Link
	setting   *service.Setting
	stat      *service.Stat
	dashboard *service.Dashboard
	captcha   *captcha.Sessions
}

func New(user *service.User, api *service.API, link *service.Link, setting *service.Setting,
	stat *service.Stat, dashboard *service.Dashboard, sessions *captcha.Sessions,
	jwt *auth.JWT, admins httpx.AdminChecker) http.Handler {
	s := &Server{user: user, api: api, link: link, setting: setting, stat: stat, dashboard: dashboard, captcha: sessions}
	r := chi.NewRouter()
	r.Use(httpx.CORS, httpx.Recover, middleware.GetHead)
	login := r.With(httpx.RequireLogin(jwt))
	admin := r.With(httpx.RequireAdmin(jwt, admins))

	r.Get(httpx.UserURL+"/Captcha", s.GetCaptcha)
	r.Post(httpx.UserURL+"/Register", s.Register)
	r.Post(httpx.UserURL+"/Login", s.Login)
	login.Get(httpx.UserURL+"/", adapt(s.GetUserInfo))
	login.Put(httpx.UserURL+"/nick", adapt(s.UpdateUserNick))
	login.Put(httpx.UserURL+"/mail", adapt(s.UpdateUserMail))
	login.Put(httpx.UserURL+"/password", adapt(s.UpdateUserPassword))
	admin.Put(httpx.UserURL+"/setPassword", adapt(s.SetUserPassword))
	admin.Put(httpx.UserURL+"/setMode", adapt(s.SetUserMode))
	login.Get(httpx.UserURL+"/key", adapt(s.GetUserKey))
	admin.Delete(httpx.UserURL+"/{userId}", adapt(s.DeleteUser))
	admin.Delete(httpx.UserURL+"/", adapt(s.DeleteUser))
	login.Get(httpx.UserURL+"/AppList", adapt(s.GetUserAPIList))
	login.Get(httpx.UserURL+"/AppList/Search", adapt(s.UserAPIListSearch))
	admin.Get(httpx.UserURL+"/list", adapt(s.GetUserList))
	admin.Get(httpx.UserURL+"/list/Search", adapt(s.UserListSearch))
	login.Put(httpx.UserURL+"/key/", adapt(s.ResetUserKey))
	login.Put(httpx.UserURL+"/key/{userId}", adapt(s.ResetUserKey))
	login.Put(httpx.UserURL+"/key", adapt(s.ResetUserKey))
	admin.Get(httpx.UserURL+"/CheckAppList", adapt(s.GetCheckAPIList))
	admin.Get(httpx.UserURL+"/CheckAppList/Search", adapt(s.CheckAPIListSearch))
	admin.Put(httpx.UserURL+"/CheckApp", adapt(s.CheckAppChange))

	r.Get(httpx.APIURL+"/list", adapt(s.GetAPIList))
	login.Post(httpx.APIURL+"/", adapt(s.CreateAPI))
	login.Put(httpx.APIURL+"/{apiId}", adapt(s.UpdateAPI))
	login.Put(httpx.APIURL+"/", adapt(s.UpdateAPI))
	login.Delete(httpx.APIURL+"/{apiId}", adapt(s.DeleteAPI))
	login.Delete(httpx.APIURL+"/", adapt(s.DeleteAPI))
	r.Get(httpx.APIURL+"/search", adapt(s.SearchAPIList))
	r.Get(httpx.APIURL+"/{apiId}", adapt(s.GetAPIApp))
	r.Get(httpx.APIURL+"/", adapt(s.GetAPIApp))
	login.Post(httpx.APIURL+"/key/{api_id}", adapt(s.CreateAPIKey))
	login.Post(httpx.APIURL+"/key/", adapt(s.CreateAPIKey))
	login.Get(httpx.APIURL+"/key/list/", adapt(s.GetUserAPIKeyList))
	login.Get(httpx.APIURL+"/key/list/{user_id}", adapt(s.GetUserAPIKeyList))
	login.Get(httpx.APIURL+"/key/list/{user_id}/Search", adapt(s.UserAPIKeyListSearch))
	login.Put(httpx.APIURL+"/key/{key}", adapt(s.UpdateAPIKey))
	login.Put(httpx.APIURL+"/key/", adapt(s.UpdateAPIKey))
	login.Get(httpx.APIURL+"/key/{key}", adapt(s.GetAPIKey))
	login.Get(httpx.APIURL+"/key/", adapt(s.GetAPIKey))
	login.Delete(httpx.APIURL+"/key/{key}", adapt(s.DeleteAPIKey))
	login.Delete(httpx.APIURL+"/key/", adapt(s.DeleteAPIKey))
	login.Put(httpx.APIURL+"/key/reset/{key}", adapt(s.ResetAPIKey))
	login.Put(httpx.APIURL+"/key/reset/", adapt(s.ResetAPIKey))

	r.Get(httpx.LinkURL+"/", adapt(s.GetLinkList))
	admin.Post(httpx.LinkURL+"/", adapt(s.CreateLink))
	admin.Put(httpx.LinkURL+"/{link_id}", adapt(s.UpdateLink))
	admin.Delete(httpx.LinkURL+"/{link_id}", adapt(s.DeleteLink))
	admin.Get(httpx.LinkURL+"/Search", adapt(s.GetLinkListSearch))

	admin.Get(httpx.SettingURL+"/list", adapt(s.GetSettingList))
	r.Get(httpx.SettingURL+"/{settingKey}", adapt(s.GetSetting))
	admin.Put(httpx.SettingURL+"/", adapt(s.UpdateSettingValue))

	r.Post(httpx.StatURL+"/log", adapt(s.LogAPI))
	r.Post(httpx.StatURL+"/log/", adapt(s.LogAPI))
	login.Get(httpx.StatURL+"/dashboard", adapt(s.GetDashboard))
	r.Get(httpx.StatURL+"/count", adapt(s.GetAPIAllCount))

	r.Get("/Logo.psd", serveLogo)
	return r
}

func parameter(r *http.Request, name string) string {
	value := chi.URLParam(r, name)
	if value == "" {
		return "0"
	}
	if r.URL.RawPath != "" {
		if decoded, err := url.PathUnescape(value); err == nil {
			return decoded
		}
	}
	return value
}

func adapt(action func(*http.Request) (httpx.Response, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := action(r)
		writeResult(w, r, result, err)
	}
}

func writeResult(w http.ResponseWriter, r *http.Request, result httpx.Response, err error) {
	if err != nil {
		writeFailure(w, r, err)
		return
	}
	httpx.Write(w, http.StatusOK, result)
}

func writeFailure(w http.ResponseWriter, r *http.Request, err error) {
	slog.ErrorContext(r.Context(), "请求处理失败", "方法", r.Method, "路由", chi.RouteContext(r.Context()).RoutePattern(), "错误", err)
	httpx.Write(w, http.StatusInternalServerError, httpx.Reply(500, httpx.MsgInternal))
}
