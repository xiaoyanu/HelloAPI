package handler

import (
	"net/http"

	"github.com/zxz/helloapi/internal/httpx"
)

func (s *Server) GetCaptcha(w http.ResponseWriter, r *http.Request) {
	if err := s.captcha.Issue(w, r); err != nil {
		writeFailure(w, r, err)
	}
}

func (s *Server) verifyCaptcha(w http.ResponseWriter, r *http.Request, clearOnFailure bool) func(string) bool {
	return func(code string) bool { return s.captcha.Verify(w, r, code, clearOnFailure) }
}

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	result, err := s.user.Register(r.Context(), httpx.StringMap(r.Body), s.verifyCaptcha(w, r, false))
	writeResult(w, r, result, err)
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	result, err := s.user.Login(r.Context(), httpx.StringMap(r.Body), s.verifyCaptcha(w, r, true))
	writeResult(w, r, result, err)
}

func (s *Server) GetUserInfo(r *http.Request) (httpx.Response, error) {
	return s.user.GetUserInfo(r.Context(), httpx.Query(r))
}

func (s *Server) UpdateUserNick(r *http.Request) (httpx.Response, error) {
	return s.user.UpdateUserNick(r.Context(), httpx.Query(r), httpx.StringMap(r.Body))
}

func (s *Server) UpdateUserMail(r *http.Request) (httpx.Response, error) {
	return s.user.UpdateUserMail(r.Context(), httpx.Query(r), httpx.StringMap(r.Body))
}

func (s *Server) UpdateUserPassword(r *http.Request) (httpx.Response, error) {
	return s.user.UpdateUserPassword(r.Context(), httpx.Query(r), httpx.StringMap(r.Body))
}

func (s *Server) SetUserPassword(r *http.Request) (httpx.Response, error) {
	return s.user.SetUserPassword(r.Context(), httpx.Query(r), httpx.StringMap(r.Body))
}

func (s *Server) SetUserMode(r *http.Request) (httpx.Response, error) {
	return s.user.SetUserMode(r.Context(), httpx.Query(r), httpx.StringMap(r.Body))
}

func (s *Server) GetUserKey(r *http.Request) (httpx.Response, error) {
	return s.user.GetUserKey(r.Context())
}

func (s *Server) DeleteUser(r *http.Request) (httpx.Response, error) {
	return s.user.DeleteUser(r.Context(), parameter(r, "userId"))
}

func (s *Server) GetUserAPIList(r *http.Request) (httpx.Response, error) {
	return s.user.GetUserAPIList(r.Context(), httpx.Query(r))
}

func (s *Server) UserAPIListSearch(r *http.Request) (httpx.Response, error) {
	return s.user.UserAPIListSearch(r.Context(), httpx.Query(r))
}

func (s *Server) GetUserList(r *http.Request) (httpx.Response, error) {
	return s.user.GetUserList(r.Context(), httpx.Query(r))
}

func (s *Server) UserListSearch(r *http.Request) (httpx.Response, error) {
	return s.user.UserListSearch(r.Context(), httpx.Query(r))
}

func (s *Server) ResetUserKey(r *http.Request) (httpx.Response, error) {
	return s.user.ResetUserKey(r.Context(), parameter(r, "userId"))
}

func (s *Server) GetCheckAPIList(r *http.Request) (httpx.Response, error) {
	return s.user.GetCheckAPIList(r.Context(), httpx.Query(r))
}

func (s *Server) CheckAPIListSearch(r *http.Request) (httpx.Response, error) {
	return s.user.CheckAPIListSearch(r.Context(), httpx.Query(r))
}

func (s *Server) CheckAppChange(r *http.Request) (httpx.Response, error) {
	return s.user.CheckAppChange(r.Context(), httpx.StringMap(r.Body))
}
