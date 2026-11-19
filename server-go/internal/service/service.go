package service

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/model"
)

type Page[T any] struct {
	Total int64 `json:"total"`
	List  []T   `json:"list"`
}

func pageResult[T any](total int64, list []T) Page[T] {
	if list == nil {
		list = []T{}
	}
	return Page[T]{Total: total, List: list}
}

func pagination(query map[string]string) (size, offset int) {
	size = httpx.StrToInt(query["pageSize"])
	page := httpx.StrToInt(query["page"])
	if size < 1 {
		size = 30
	}
	if page < 1 {
		page = 1
	}
	// Java 的分页偏移使用 32 位整数运算。
	return size, int(int32(page-1) * int32(size))
}

func md5Text(value string) string {
	digest := md5.Sum([]byte(value))
	return hex.EncodeToString(digest[:])
}

func keyFor(id int) string {
	return md5Text(strconv.Itoa(id) + strconv.FormatInt(time.Now().UnixMilli(), 10))
}

func trim(value string) string {
	return strings.TrimFunc(value, func(r rune) bool { return r <= 0x20 })
}

func textLength(value string) int {
	length := 0
	for _, r := range value {
		length += utf16.RuneLen(r)
	}
	return length
}

func optional(values map[string]string, key string) *string {
	value, ok := values[key]
	if !ok {
		return nil
	}
	return &value
}

func paramError() httpx.Response { return httpx.Reply(400, httpx.MsgParam) }
func noUser() httpx.Response     { return httpx.Reply(401, httpx.MsgUserNotFound) }
func notAdmin() httpx.Response   { return httpx.Reply(403, httpx.MsgNotAdmin) }
func notCreator() httpx.Response { return httpx.Reply(400, httpx.MsgNotCreator) }

func paramOnError(response *httpx.Response, err *error) {
	if *err != nil {
		*response, *err = paramError(), nil
	}
}

type settingReader interface {
	GetSettingValue(context.Context, string) (*string, error)
}

func disabled(value *string) bool { return value != nil && *value == "false" }

type appListItem struct {
	ID         int                 `json:"id"`
	Title      *string             `json:"title"`
	Type       int                 `json:"type"`
	Status     int                 `json:"status"`
	Created    model.LocalDateTime `json:"created"`
	Updated    model.LocalDateTime `json:"updated"`
	UserID     int                 `json:"user_id"`
	ViewStatus int                 `json:"view_status"`
}

func appList(apps []model.APIApp) []appListItem {
	items := make([]appListItem, 0, len(apps))
	for _, app := range apps {
		items = append(items, appListItem{
			ID: app.ID, Title: app.Title, Type: app.Type, Status: app.Status, Created: app.Created,
			Updated: app.Updated, UserID: app.UserID, ViewStatus: app.ViewStatus,
		})
	}
	return items
}
