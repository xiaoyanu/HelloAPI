package repo

import (
	"context"
	"github.com/zxz/helloapi/internal/model"
)

func (s *Store) GetSettingValue(ctx context.Context, key string) (*string, error) {
	const query = "SELECT `value` FROM `helloapi_settings` WHERE `key` = ?"
	return getRow[string](ctx, s.query, query, key)
}

func (s *Store) UpdateSettingValue(ctx context.Context, key string, value string) error {
	const query = "INSERT INTO `helloapi_settings` (`key`, `value`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `value` = ?"
	return execute(ctx, s.query, query, key, value, value)
}

func (s *Store) GetSettingValueAll(ctx context.Context) ([]model.Setting, error) {
	const query = "SELECT * FROM `helloapi_settings`"
	return selectRows[model.Setting](ctx, s.query, query)
}
