package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/model"
	"github.com/zxz/helloapi/internal/repo"
)

// passwordRepoFake 只实现修改密码用到的方法，其余方法调用会因嵌入的 nil 接口而 panic。
type passwordRepoFake struct {
	UserRepository
	updated    int64
	updateHits int
	setHits    int
}

func (f *passwordRepoFake) CheckUserIDExists(context.Context, int) (int64, error) { return 1, nil }

func (f *passwordRepoFake) UpdateUserPassword(context.Context, int, string, string) (int64, error) {
	f.updateHits++
	return f.updated, nil
}

func (f *passwordRepoFake) SetUserPassword(context.Context, int, string) error {
	f.setHits++
	return nil
}

func TestUpdateUserPasswordRejectsWrongOldPassword(t *testing.T) {
	ctx := httpx.WithUser(context.Background(), httpx.Identity{ID: 7})
	cases := []struct {
		name     string
		body     map[string]string
		updated  int64
		wantCode int
		wantMsg  string
		wantHits int
	}{
		{"旧密码错误", map[string]string{"oldPassword": "wrong-old", "newPassword": "new-pass"}, 0, 400, "旧密码错误", 1},
		{"修改成功", map[string]string{"oldPassword": "right-old", "newPassword": "new-pass"}, 1, 200, "成功", 1},
		{"旧密码为空", map[string]string{"newPassword": "new-pass"}, 1, 400, "旧密码不能为空", 0},
		{"新密码过短", map[string]string{"oldPassword": "right-old", "newPassword": "abc"}, 1, 400, "密码不能少于6个字符", 0},
		{"新密码含空格", map[string]string{"oldPassword": "right-old", "newPassword": "abc def"}, 1, 400, "密码不能包含空格", 0},
		{"新旧相同", map[string]string{"oldPassword": "same-pass", "newPassword": "same-pass"}, 1, 400, "新旧密码相同", 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fake := &passwordRepoFake{updated: tc.updated}
			response, err := NewUser(fake, nil).UpdateUserPassword(ctx, map[string]string{}, tc.body)
			if err != nil || response.Code != tc.wantCode || response.Msg != tc.wantMsg || fake.updateHits != tc.wantHits {
				t.Fatalf("响应 = %+v，错误 = %v，更新次数 = %d", response, err, fake.updateHits)
			}
		})
	}
}

func TestSetUserPasswordValidatesNewPassword(t *testing.T) {
	ctx := httpx.WithUser(context.Background(), httpx.Identity{ID: 1, Mode: httpx.Admin})
	fake := &passwordRepoFake{}
	service := NewUser(fake, nil)
	response, err := service.SetUserPassword(ctx, map[string]string{"id": "2"}, map[string]string{"password": "12345"})
	if err != nil || response.Code != 400 || fake.setHits != 0 {
		t.Fatalf("过短密码未被拒绝：%+v，%v，写入次数 %d", response, err, fake.setHits)
	}
	response, err = service.SetUserPassword(ctx, map[string]string{"id": "2"}, map[string]string{"password": "123456"})
	if err != nil || response.Code != 200 || fake.setHits != 1 {
		t.Fatalf("合法密码未写入：%+v，%v，写入次数 %d", response, err, fake.setHits)
	}
}

// apiWriteFake 模拟事务：记录写入，回调出错时视为回滚并丢弃本次写入。
type apiWriteFake struct {
	APIRepository
	paramAffected int64
	committed     []string
	pending       []string
	attempted     bool
}

func (f *apiWriteFake) GetSettingValue(context.Context, string) (*string, error) { return nil, nil }
func (f *apiWriteFake) CheckUserIDExists(context.Context, int) (int64, error)    { return 1, nil }
func (f *apiWriteFake) CheckAPIExist(context.Context, int) (int64, error)        { return 1, nil }
func (f *apiWriteFake) CheckAPICreator(context.Context, int, int) (int64, error) { return 1, nil }

func (f *apiWriteFake) WriteAPIInTx(_ context.Context, fn func(repo.APIWriter) error) error {
	f.pending = nil
	f.attempted = true
	if err := fn(f); err != nil {
		f.pending = nil
		return err
	}
	f.committed = append(f.committed, f.pending...)
	return nil
}

func (f *apiWriteFake) CreateAPIApp(_ context.Context, app *model.APIApp) (int64, error) {
	app.ID = 99
	f.pending = append(f.pending, "app")
	return 1, nil
}

func (f *apiWriteFake) UpdateAPIApp(context.Context, *model.APIApp) (int64, error) {
	f.pending = append(f.pending, "app")
	return 1, nil
}

func (f *apiWriteFake) DeleteAPIParamAll(context.Context, int) error {
	f.pending = append(f.pending, "delete-params")
	return nil
}

func (f *apiWriteFake) CheckAPIParamExistObj(context.Context, *model.APIParam) (int64, error) {
	return 0, nil
}

func (f *apiWriteFake) UpdateAPIParam(context.Context, *model.APIParam) (int64, error) {
	return 0, errors.New("不应调用")
}

func (f *apiWriteFake) CreateAPIParam(context.Context, *model.APIParam) (int64, error) {
	f.pending = append(f.pending, "param")
	return f.paramAffected, nil
}

// apiBody 通过 httpx.ObjectMap 解码，保证数字类型与真实请求一致（json.Number）。
func apiBody(t *testing.T) map[string]any {
	t.Helper()
	body, err := httpx.ObjectMap(strings.NewReader(`{"title":"测试","smallTitle":"描述","type":0,"status":0,
		"url":"https://example.com","sendType":0,"returnType":"json","returnContent":"{}","view_status":0,
		"params":[{"name":"q","required":1,"type":"string","msg":"关键词"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func TestCreateAndUpdateAPIRollbackWhenParamWriteFails(t *testing.T) {
	ctx := httpx.WithUser(context.Background(), httpx.Identity{ID: 1, Mode: httpx.Admin})
	for _, tc := range []struct {
		name string
		run  func(*API, map[string]any) (httpx.Response, error)
	}{
		{"创建", func(s *API, body map[string]any) (httpx.Response, error) { return s.CreateAPI(ctx, body) }},
		{"更新", func(s *API, body map[string]any) (httpx.Response, error) { return s.UpdateAPI(ctx, "99", body) }},
	} {
		t.Run(tc.name+"失败回滚", func(t *testing.T) {
			fake := &apiWriteFake{paramAffected: 0}
			response, err := tc.run(NewAPI(fake, nil), apiBody(t))
			// 必须确认事务回调确实执行过（attempted），否则可能是请求在校验阶段就被拒绝。
			if err != nil || response.Code != 400 || len(fake.committed) != 0 || !fake.attempted {
				t.Fatalf("响应 = %+v，错误 = %v，已提交 = %v", response, err, fake.committed)
			}
		})
		t.Run(tc.name+"成功提交", func(t *testing.T) {
			fake := &apiWriteFake{paramAffected: 1}
			response, err := tc.run(NewAPI(fake, nil), apiBody(t))
			if err != nil || response.Code != 200 || len(fake.committed) == 0 {
				t.Fatalf("响应 = %+v，错误 = %v，已提交 = %v", response, err, fake.committed)
			}
		})
	}
}
