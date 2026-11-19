package service

import (
	"context"

	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/model"
)

type SettingRepository interface {
	GetSettingValueAll(ctx context.Context) ([]model.Setting, error)
	UpdateSettingValue(ctx context.Context, key string, value string) error
	settingReader
}

type Setting struct{ repo SettingRepository }

func NewSetting(store SettingRepository) *Setting { return &Setting{repo: store} }

func (s *Setting) GetSettingList(ctx context.Context) (httpx.Response, error) {
	settings, err := s.repo.GetSettingValueAll(ctx)
	if err != nil {
		return httpx.Response{}, err
	}
	type item struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	list := make([]item, 0, len(settings))
	for _, setting := range settings {
		list = append(list, item{Key: setting.Key, Value: setting.Value})
	}
	return httpx.Success(list), nil
}

func (s *Setting) GetSetting(ctx context.Context, key string) (httpx.Response, error) {
	switch key {
	case "link_name", "link_desc", "link_url", "link_icon", "link_email":
	default:
		return paramError(), nil
	}
	value, err := s.repo.GetSettingValue(ctx, key)
	return httpx.Success(value), err
}

func (s *Setting) UpdateSettingValue(ctx context.Context, body map[string]string) (httpx.Response, error) {
	key, value := body["key"], body["value"]
	if key == "" || value == "" {
		return paramError(), nil
	}
	return httpx.Success(), s.repo.UpdateSettingValue(ctx, key, value)
}
