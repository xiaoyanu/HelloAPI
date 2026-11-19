package service

import (
	"context"

	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/model"
)

type LinkRepository interface {
	GetLinkList(ctx context.Context, pageSize int, offset int, desc bool) ([]model.Link, error)
	GetLinkListCount(ctx context.Context) (int64, error)
	GetLinkListSearch(ctx context.Context, keyword string, pageSize int, offset int) ([]model.Link, error)
	GetLinkListSearchCount(ctx context.Context, keyword string) (int64, error)
	InsertLink(ctx context.Context, link *model.Link) (int64, error)
	UpdateLink(ctx context.Context, link *model.Link) (int64, error)
	DeleteLink(ctx context.Context, linkID int) (int64, error)
	GetLinkByID(ctx context.Context, linkID int) (*model.Link, error)
}

type Link struct{ repo LinkRepository }

func NewLink(store LinkRepository) *Link { return &Link{repo: store} }

func (s *Link) GetLinkList(ctx context.Context, query map[string]string) (httpx.Response, error) {
	size, offset := pagination(query)
	total, err := s.repo.GetLinkListCount(ctx)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.GetLinkList(ctx, size, offset, httpx.StrToInt(query["desc"]) >= 1)
	return httpx.Reply(200, "获取成功", pageResult(total, list)), err
}

func (s *Link) GetLinkListSearch(ctx context.Context, query map[string]string) (httpx.Response, error) {
	keyword := query["keywords"]
	size, offset := pagination(query)
	total, err := s.repo.GetLinkListSearchCount(ctx, keyword)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.GetLinkListSearch(ctx, keyword, size, offset)
	return httpx.Success(pageResult(total, list)), err
}

func (s *Link) CreateLink(ctx context.Context, body map[string]string) (httpx.Response, error) {
	name, url := body["name"], body["url"]
	if name == "" {
		return httpx.Reply(400, "名称不能为空"), nil
	}
	if url == "" {
		return httpx.Reply(400, "链接不能为空"), nil
	}

	link := model.Link{Name: name, URL: url, Desc: optional(body, "desc"), Avatar: optional(body, "avatar")}
	result, err := s.repo.InsertLink(ctx, &link)
	if err != nil {
		return httpx.Response{}, err
	}
	if result < 1 {
		return httpx.Reply(500, "添加失败"), nil
	}
	return httpx.Reply(200, "添加成功", link), nil
}

func (s *Link) UpdateLink(ctx context.Context, id string, body map[string]string) (httpx.Response, error) {
	linkID := httpx.StrToInt(id)
	if linkID <= 0 {
		return httpx.Reply(400, "无效的友链ID"), nil
	}
	link, err := s.repo.GetLinkByID(ctx, linkID)
	if err != nil {
		return httpx.Response{}, err
	}
	if link == nil {
		return httpx.Reply(404, "友链不存在"), nil
	}
	if value, ok := body["name"]; ok {
		link.Name = value
	}
	if value, ok := body["url"]; ok {
		link.URL = value
	}
	if value := optional(body, "desc"); value != nil {
		link.Desc = value
	}
	if value := optional(body, "avatar"); value != nil {
		link.Avatar = value
	}
	result, err := s.repo.UpdateLink(ctx, link)
	if err != nil {
		return httpx.Response{}, err
	}
	if result < 1 {
		return httpx.Reply(500, "修改失败"), nil
	}
	return httpx.Reply(200, "修改成功", link), nil
}

func (s *Link) DeleteLink(ctx context.Context, id string) (httpx.Response, error) {
	linkID := httpx.StrToInt(id)
	if linkID <= 0 {
		return httpx.Reply(400, "无效的友链ID"), nil
	}
	result, err := s.repo.DeleteLink(ctx, linkID)
	if err != nil {
		return httpx.Response{}, err
	}
	if result < 1 {
		return httpx.Reply(500, "删除失败或友链不存在"), nil
	}
	return httpx.Reply(200, "删除成功"), nil
}
