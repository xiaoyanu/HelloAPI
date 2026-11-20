package handler

import (
	"net/http"

	"github.com/zxz/helloapi/internal/httpx"
)

func (s *Server) GetAPIList(r *http.Request) (httpx.Response, error) {
	return s.api.GetAPIList(r.Context(), httpx.Query(r))
}

func (s *Server) CreateAPI(r *http.Request) (httpx.Response, error) {
	body, err := httpx.ObjectMap(r.Body)
	if err != nil {
		return httpx.Reply(400, httpx.MsgParam), nil
	}
	return s.api.CreateAPI(r.Context(), body)
}

func (s *Server) UpdateAPI(r *http.Request) (httpx.Response, error) {
	body, err := httpx.ObjectMap(r.Body)
	if err != nil {
		return httpx.Reply(400, httpx.MsgParam), nil
	}
	return s.api.UpdateAPI(r.Context(), parameter(r, "apiId"), body)
}

func (s *Server) DeleteAPI(r *http.Request) (httpx.Response, error) {
	return s.api.DeleteAPI(r.Context(), parameter(r, "apiId"))
}

func (s *Server) SearchAPIList(r *http.Request) (httpx.Response, error) {
	return s.api.SearchAPIList(r.Context(), httpx.Query(r))
}

func (s *Server) GetAPIApp(r *http.Request) (httpx.Response, error) {
	return s.api.GetAPIApp(r.Context(), parameter(r, "apiId"), r.Header.Get("Authorization"))
}

func (s *Server) CreateAPIKey(r *http.Request) (httpx.Response, error) {
	body, err := httpx.ObjectMap(r.Body)
	if err != nil {
		return httpx.Reply(400, httpx.MsgParam), nil
	}
	return s.api.CreateAPIKey(r.Context(), parameter(r, "api_id"), body)
}

func (s *Server) GetUserAPIKeyList(r *http.Request) (httpx.Response, error) {
	return s.api.GetUserAPIKeyList(r.Context(), parameter(r, "user_id"), httpx.Query(r))
}

func (s *Server) UserAPIKeyListSearch(r *http.Request) (httpx.Response, error) {
	return s.api.UserAPIKeyListSearch(r.Context(), parameter(r, "user_id"), httpx.Query(r))
}

func (s *Server) UpdateAPIKey(r *http.Request) (httpx.Response, error) {
	body, err := httpx.ObjectMap(r.Body)
	if err != nil {
		return httpx.Reply(400, httpx.MsgParam), nil
	}
	return s.api.UpdateAPIKey(r.Context(), parameter(r, "key"), body)
}

func (s *Server) GetAPIKey(r *http.Request) (httpx.Response, error) {
	return s.api.GetAPIKey(r.Context(), parameter(r, "key"))
}

func (s *Server) DeleteAPIKey(r *http.Request) (httpx.Response, error) {
	return s.api.DeleteAPIKey(r.Context(), parameter(r, "key"))
}

func (s *Server) ResetAPIKey(r *http.Request) (httpx.Response, error) {
	return s.api.ResetAPIKey(r.Context(), parameter(r, "key"))
}
