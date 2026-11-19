package handler

import (
	"net/http"

	"github.com/zxz/helloapi/internal/httpx"
)

func (s *Server) GetSettingList(r *http.Request) (httpx.Response, error) {
	return s.setting.GetSettingList(r.Context())
}

func (s *Server) GetSetting(r *http.Request) (httpx.Response, error) {
	return s.setting.GetSetting(r.Context(), parameter(r, "settingKey"))
}

func (s *Server) UpdateSettingValue(r *http.Request) (httpx.Response, error) {
	return s.setting.UpdateSettingValue(r.Context(), httpx.StringMap(r.Body))
}
