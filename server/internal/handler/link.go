package handler

import (
	"net/http"

	"github.com/zxz/helloapi/internal/httpx"
)

func (s *Server) GetLinkList(r *http.Request) (httpx.Response, error) {
	return s.link.GetLinkList(r.Context(), httpx.Query(r))
}

func (s *Server) CreateLink(r *http.Request) (httpx.Response, error) {
	return s.link.CreateLink(r.Context(), httpx.StringMap(r.Body))
}

func (s *Server) UpdateLink(r *http.Request) (httpx.Response, error) {
	return s.link.UpdateLink(r.Context(), parameter(r, "link_id"), httpx.StringMap(r.Body))
}

func (s *Server) DeleteLink(r *http.Request) (httpx.Response, error) {
	return s.link.DeleteLink(r.Context(), parameter(r, "link_id"))
}

func (s *Server) GetLinkListSearch(r *http.Request) (httpx.Response, error) {
	return s.link.GetLinkListSearch(r.Context(), httpx.Query(r))
}
