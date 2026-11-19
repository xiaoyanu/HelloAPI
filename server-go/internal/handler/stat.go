package handler

import (
	"net/http"

	"github.com/zxz/helloapi/internal/httpx"
)

func (s *Server) LogAPI(r *http.Request) (httpx.Response, error) {
	return s.stat.LogAPI(r.Context(), httpx.StringMap(r.Body))
}

func (s *Server) GetDashboard(r *http.Request) (httpx.Response, error) {
	value, err := s.dashboard.GetDashboard(r.Context())
	return httpx.Success(value), err
}

func (s *Server) GetAPIAllCount(r *http.Request) (httpx.Response, error) {
	value, err := s.dashboard.GetAPIAllCount(r.Context())
	return httpx.Success(map[string]int64{"count": value}), err
}
