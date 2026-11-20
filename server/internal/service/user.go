package service

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/model"
)

type UserRepository interface {
	CheckUsernameExists(ctx context.Context, username string) (int64, error)
	Register(ctx context.Context, username string, password string, nick string, mode int, mail string) (int64, error)
	GetUserCount(ctx context.Context) (int64, error)
	Login(ctx context.Context, username string, password string) (*model.User, error)
	CheckUserIDExists(ctx context.Context, userID int) (int64, error)
	GetUserInfo(ctx context.Context, id int) (*model.User, error)
	GetUserKey(ctx context.Context, userID int) (*model.UserKey, error)
	GetUserAPIList(ctx context.Context, userID int, pageSize int, offset int) ([]model.APIApp, error)
	GetUserAPIListAllCount(ctx context.Context, userID int) (int64, error)
	ResetUserKey(ctx context.Context, userID int, key string) error
	GetUserAPIListSearch(ctx context.Context, userID int, keyword string, kind int, status int, viewStatus int, pageSize int, offset int) ([]model.APIApp, error)
	GetUserAPIListSearchCount(ctx context.Context, userID int, keyword string, kind int, status int, viewStatus int) (int64, error)
	UpdateUserNick(ctx context.Context, userID int, nick string) error
	UpdateUserMail(ctx context.Context, userID int, mail string) error
	UpdateUserPassword(ctx context.Context, userID int, oldPassword string, newPassword string) (int64, error)
	SetUserPassword(ctx context.Context, userID int, password string) error
	GetUserList(ctx context.Context, pageSize int, offset int) ([]model.User, error)
	GetUserListSearch(ctx context.Context, keyword string, mode int, pageSize int, offset int) ([]model.User, error)
	GetUserListSearchCount(ctx context.Context, keyword string, mode int) (int64, error)
	SetUserMode(ctx context.Context, userID int, mode int) error
	GetCheckAPIList(ctx context.Context, pageSize int, offset int) ([]model.APIApp, error)
	GetCheckAPIListAllCount(ctx context.Context) (int64, error)
	GetCheckAPIListSearch(ctx context.Context, keyword string, kind int, status int, pageSize int, offset int) ([]model.APIApp, error)
	GetCheckAPIListSearchCount(ctx context.Context, keyword string, kind int, status int) (int64, error)
	CheckAppChange(ctx context.Context, apiID int, viewStatus int) error
	CheckAPIExist(ctx context.Context, apiID int) (int64, error)
	settingReader
	DeleteUserData(context.Context, int) error
}

type LoginTokens interface {
	Issue(id int, name string, mode int) (string, error)
}

type User struct {
	repo   UserRepository
	tokens LoginTokens
}

func NewUser(store UserRepository, tokens LoginTokens) *User {
	return &User{repo: store, tokens: tokens}
}

var mailPattern = regexp.MustCompile(`^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9]{2,6}$`)

func (s *User) Register(ctx context.Context, body map[string]string, verify func(string) bool) (httpx.Response, error) {
	if len(body) == 0 {
		return httpx.Reply(400, "请求参数不能为空"), nil
	}
	username, password, code, mail := body["username"], body["password"], body["code"], body["mail"]
	switch {
	case trim(username) == "":
		return httpx.Reply(400, "账号不能为空"), nil
	case username == password:
		return httpx.Reply(400, "账号和密码不能相同"), nil
	case textLength(username) < 4:
		return httpx.Reply(400, "账号不能少于4个字符"), nil
	case trim(password) == "":
		return httpx.Reply(400, "密码不能为空"), nil
	case textLength(password) < 6:
		return httpx.Reply(400, "密码不能少于6个字符"), nil
	case strings.Contains(password, " "):
		return httpx.Reply(400, "密码不能包含空格"), nil
	case strings.Contains(username, " "):
		return httpx.Reply(400, "账号不能包含空格"), nil
	case textLength(username) > 32:
		return httpx.Reply(400, "账号不能超过32个字符"), nil
	case textLength(password) > 64:
		return httpx.Reply(400, "密码不能超过64个字符"), nil
	case code == "":
		return httpx.Reply(400, "验证码不能为空"), nil
	case textLength(code) != 4:
		return httpx.Reply(400, "验证码长度必须为4个字符"), nil
	case trim(mail) == "":
		return httpx.Reply(400, "邮箱不能为空"), nil
	case !mailPattern.MatchString(mail):
		return httpx.Reply(400, "邮箱格式错误"), nil
	case !verify(code):
		return httpx.Reply(400, "验证码错误/过期，请重新输入"), nil
	}

	setting, err := s.repo.GetSettingValue(ctx, "register")
	if err != nil {
		return httpx.Response{}, err
	}
	if disabled(setting) {
		return httpx.Reply(400, "注册已关闭"), nil
	}
	exists, err := s.repo.CheckUsernameExists(ctx, username)
	if err != nil {
		return httpx.Reply(500, "注册失败"), nil
	}
	if exists > 0 {
		return httpx.Reply(400, "账号已存在"), nil
	}
	count, err := s.repo.GetUserCount(ctx)
	if err != nil {
		return httpx.Reply(500, "注册失败"), nil
	}
	mode := 0
	if count < 1 {
		mode = httpx.Admin
	}
	affected, err := s.repo.Register(ctx, username, md5Text(password), username, mode, mail)
	if err != nil {
		return httpx.Reply(500, "注册失败"), nil
	}
	if affected < 1 {
		return httpx.Reply(500, "注册失败，数据库插入失败"), nil
	}
	return httpx.Reply(200, "注册成功"), nil
}

func (s *User) Login(ctx context.Context, body map[string]string, verify func(string) bool) (httpx.Response, error) {
	username, password, code := body["username"], body["password"], body["code"]
	switch {
	case trim(username) == "":
		return httpx.Reply(400, "账号不能为空"), nil
	case trim(password) == "":
		return httpx.Reply(400, "密码不能为空"), nil
	case strings.Contains(username, " "):
		return httpx.Reply(400, "账号不能包含空格"), nil
	case strings.Contains(password, " "):
		return httpx.Reply(400, "密码不能包含空格"), nil
	case textLength(username) > 32:
		return httpx.Reply(400, "账号不能超过32个字符"), nil
	case textLength(username) < 4:
		return httpx.Reply(400, "账号不能少于4个字符"), nil
	case textLength(password) < 6:
		return httpx.Reply(400, "密码不能少于6个字符"), nil
	case textLength(password) > 64:
		return httpx.Reply(400, "密码不能超过64个字符"), nil
	case code == "":
		return httpx.Reply(400, "验证码不能为空"), nil
	case textLength(code) != 4:
		return httpx.Reply(400, "验证码长度必须为4个字符"), nil
	case !verify(code):
		return httpx.Reply(400, "验证码错误/过期，请重新输入"), nil
	}
	exists, err := s.repo.CheckUsernameExists(ctx, username)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return httpx.Reply(400, "账号不存在"), nil
	}
	user, err := s.repo.Login(ctx, username, md5Text(password))
	if err != nil {
		return httpx.Response{}, err
	}
	if user == nil {
		return httpx.Reply(400, "密码错误"), nil
	}
	token, err := s.tokens.Issue(user.ID, user.Username, user.Mode)
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Reply(200, "登陆成功", struct {
		Token string         `json:"token"`
		User  model.UserInfo `json:"user"`
	}{Token: token, User: user.UserInfo}), nil
}

func (s *User) GetUserInfo(ctx context.Context, query map[string]string) (httpx.Response, error) {
	identity := httpx.User(ctx)
	exists, err := s.repo.CheckUserIDExists(ctx, identity.ID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	userID := identity.ID
	if target := query["id"]; trim(target) != "" {
		if identity.Mode != httpx.Admin {
			return notAdmin(), nil
		}
		userID = httpx.StrToInt(target)
	}
	user, err := s.repo.GetUserInfo(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	if user == nil {
		return noUser(), nil
	}
	return httpx.Reply(200, "获取成功", user.UserInfo), nil
}

func (s *User) GetUserKey(ctx context.Context) (httpx.Response, error) {
	userID := httpx.User(ctx).ID
	exists, err := s.repo.CheckUserIDExists(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	key, err := s.repo.GetUserKey(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	value := ""
	if key != nil {
		value = key.Key
	}
	return httpx.Reply(200, "获取成功", struct {
		ID  int    `json:"id"`
		Key string `json:"key"`
	}{ID: userID, Key: value}), nil
}

func (s *User) DeleteUser(ctx context.Context, id string) (response httpx.Response, err error) {
	defer paramOnError(&response, &err)
	userID := httpx.StrToInt(id)
	if userID == 0 {
		return paramError(), nil
	}
	if userID == httpx.User(ctx).ID {
		return httpx.Reply(403, "你不能删除自己"), nil
	}
	exists, err := s.repo.CheckUserIDExists(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	return httpx.Success(), s.repo.DeleteUserData(ctx, userID)
}

func (s *User) ResetUserKey(ctx context.Context, id string) (httpx.Response, error) {
	identity := httpx.User(ctx)
	setting, err := s.repo.GetSettingValue(ctx, "user_key")
	if err != nil {
		return httpx.Response{}, err
	}
	if identity.Mode != httpx.Admin && disabled(setting) {
		return httpx.Reply(400, "重置密钥已关闭"), nil
	}
	userID := httpx.StrToInt(id)
	if userID != identity.ID && identity.Mode != httpx.Admin {
		return notAdmin(), nil
	}
	exists, err := s.repo.CheckUserIDExists(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	key := keyFor(userID)
	return httpx.Success(key), s.repo.ResetUserKey(ctx, userID, key)
}

// resolveUser 按各入口的规则校验当前用户及可选的管理员指定用户。
func (s *User) resolveUser(ctx context.Context, query map[string]string, checkCurrent bool) (int, httpx.Response, error) {
	identity := httpx.User(ctx)
	if checkCurrent {
		exists, err := s.repo.CheckUserIDExists(ctx, identity.ID)
		if err != nil {
			return 0, httpx.Response{}, err
		}
		if exists < 1 {
			return 0, noUser(), nil
		}
	}
	target := query["id"]
	if trim(target) == "" {
		return identity.ID, httpx.Response{}, nil
	}
	if identity.Mode != httpx.Admin {
		return 0, notAdmin(), nil
	}
	userID := httpx.StrToInt(target)
	exists, err := s.repo.CheckUserIDExists(ctx, userID)
	if err != nil {
		return 0, httpx.Response{}, err
	}
	if exists < 1 {
		return 0, noUser(), nil
	}
	return userID, httpx.Response{}, nil
}

func (s *User) GetUserAPIList(ctx context.Context, query map[string]string) (httpx.Response, error) {
	userID, rejection, err := s.resolveUser(ctx, query, true)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	size, offset := pagination(query)
	total, err := s.repo.GetUserAPIListAllCount(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	apps, err := s.repo.GetUserAPIList(ctx, userID, size, offset)
	return httpx.Reply(200, "获取成功", pageResult(total, appList(apps))), err
}

func (s *User) UserAPIListSearch(ctx context.Context, query map[string]string) (httpx.Response, error) {
	userID, rejection, err := s.resolveUser(ctx, query, false)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	size, offset := pagination(query)
	keyword := query["keywords"]
	kind, status, view := httpx.StrToInt(query["type"]), httpx.StrToInt(query["status"]), httpx.StrToInt(query["view_status"])
	total, err := s.repo.GetUserAPIListSearchCount(ctx, userID, keyword, kind, status, view)
	if err != nil {
		return httpx.Response{}, err
	}
	apps, err := s.repo.GetUserAPIListSearch(ctx, userID, keyword, kind, status, view, size, offset)
	return httpx.Reply(200, "获取成功", pageResult(total, appList(apps))), err
}

func (s *User) UpdateUserNick(ctx context.Context, query, body map[string]string) (httpx.Response, error) {
	userID, rejection, err := s.resolveUser(ctx, query, true)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	nick := body["nick"]
	if trim(nick) == "" {
		return httpx.Reply(400, "昵称不能为空"), nil
	}
	if textLength(nick) > 32 {
		return httpx.Reply(400, "昵称不能超过32个字符"), nil
	}
	if err := s.repo.UpdateUserNick(ctx, userID, nick); err != nil {
		return httpx.Reply(500, "修改昵称失败"), nil
	}
	return httpx.Success(), nil
}

func (s *User) UpdateUserMail(ctx context.Context, query, body map[string]string) (httpx.Response, error) {
	userID, rejection, err := s.resolveUser(ctx, query, true)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	mail := body["mail"]
	if trim(mail) == "" {
		return httpx.Reply(400, "邮箱不能为空"), nil
	}
	if !mailPattern.MatchString(mail) {
		return httpx.Reply(400, "邮箱格式错误"), nil
	}
	if err := s.repo.UpdateUserMail(ctx, userID, mail); err != nil {
		return httpx.Reply(500, "修改邮箱失败"), nil
	}
	return httpx.Success(), nil
}

func (s *User) UpdateUserPassword(ctx context.Context, query, body map[string]string) (httpx.Response, error) {
	userID, rejection, err := s.resolveUser(ctx, query, true)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	if trim(body["oldPassword"]) == "" {
		return httpx.Reply(400, "旧密码不能为空"), nil
	}
	if rejection := validateNewPassword(body["newPassword"]); rejection.Code != 0 {
		return rejection, nil
	}
	oldPassword, newPassword := md5Text(body["oldPassword"]), md5Text(body["newPassword"])
	if oldPassword == newPassword {
		return httpx.Reply(400, "新旧密码相同"), nil
	}
	// 旧密码作为 WHERE 条件；未更新任何行说明旧密码不匹配（新旧密码已确保不同，不会因值未变而返回 0）。
	updated, err := s.repo.UpdateUserPassword(ctx, userID, oldPassword, newPassword)
	if err != nil {
		return httpx.Response{}, err
	}
	if updated == 0 {
		return httpx.Reply(400, "旧密码错误"), nil
	}
	return httpx.Success(), nil
}

func (s *User) SetUserPassword(ctx context.Context, query, body map[string]string) (httpx.Response, error) {
	if httpx.User(ctx).Mode != httpx.Admin {
		return notAdmin(), nil
	}
	if trim(query["id"]) == "" {
		return paramError(), nil
	}
	userID, rejection, err := s.resolveUser(ctx, query, false)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	password, present := body["password"]
	if !present {
		return httpx.Response{}, errors.New("缺少设置密码字段")
	}
	if rejection := validateNewPassword(password); rejection.Code != 0 {
		return rejection, nil
	}
	return httpx.Success(), s.repo.SetUserPassword(ctx, userID, md5Text(password))
}

// validateNewPassword 校验即将写入的新密码，规则与登录一致，避免设置后无法登录。
// 通过时返回 Code 为 0 的空响应。
func validateNewPassword(password string) httpx.Response {
	switch {
	case trim(password) == "":
		return httpx.Reply(400, "密码不能为空")
	case strings.Contains(password, " "):
		return httpx.Reply(400, "密码不能包含空格")
	case textLength(password) < 6:
		return httpx.Reply(400, "密码不能少于6个字符")
	case textLength(password) > 64:
		return httpx.Reply(400, "密码不能超过64个字符")
	}
	return httpx.Response{}
}

func (s *User) verifyAdmin(ctx context.Context) (httpx.Response, error) {
	identity := httpx.User(ctx)
	exists, err := s.repo.CheckUserIDExists(ctx, identity.ID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	if identity.Mode != httpx.Admin {
		return notAdmin(), nil
	}
	return httpx.Response{}, nil
}

func (s *User) SetUserMode(ctx context.Context, query, body map[string]string) (httpx.Response, error) {
	if rejection, err := s.verifyAdmin(ctx); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	if trim(query["id"]) == "" {
		return paramError(), nil
	}
	userID, rejection, err := s.resolveUser(ctx, query, false)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	mode := httpx.StrToInt(body["mode"])
	if mode < 0 || mode > 1 {
		return httpx.Reply(400, "权限错误"), nil
	}
	return httpx.Success(), s.repo.SetUserMode(ctx, userID, mode)
}

func (s *User) GetUserList(ctx context.Context, query map[string]string) (httpx.Response, error) {
	if rejection, err := s.verifyAdmin(ctx); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	size, offset := pagination(query)
	users, err := s.repo.GetUserList(ctx, size, offset)
	if err != nil {
		return httpx.Response{}, err
	}
	total, err := s.repo.GetUserCount(ctx)
	return httpx.Reply(200, "获取成功", pageResult(total, users)), err
}

func (s *User) UserListSearch(ctx context.Context, query map[string]string) (httpx.Response, error) {
	if rejection, err := s.verifyAdmin(ctx); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	size, offset := pagination(query)
	keyword, mode := query["keywords"], httpx.StrToInt(query["mode"])
	total, err := s.repo.GetUserListSearchCount(ctx, keyword, mode)
	if err != nil {
		return httpx.Response{}, err
	}
	users, err := s.repo.GetUserListSearch(ctx, keyword, mode, size, offset)
	return httpx.Success(pageResult(total, users)), err
}

func (s *User) GetCheckAPIList(ctx context.Context, query map[string]string) (httpx.Response, error) {
	size, offset := pagination(query)
	total, err := s.repo.GetCheckAPIListAllCount(ctx)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.GetCheckAPIList(ctx, size, offset)
	return httpx.Reply(200, "获取成功", pageResult(total, appList(list))), err
}

func (s *User) CheckAPIListSearch(ctx context.Context, query map[string]string) (httpx.Response, error) {
	size, offset := pagination(query)
	keyword := query["keywords"]
	kind, status := httpx.StrToInt(query["type"]), httpx.StrToInt(query["status"])
	total, err := s.repo.GetCheckAPIListSearchCount(ctx, keyword, kind, status)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.GetCheckAPIListSearch(ctx, keyword, kind, status, size, offset)
	return httpx.Reply(200, "获取成功", pageResult(total, appList(list))), err
}

func (s *User) CheckAppChange(ctx context.Context, body map[string]string) (httpx.Response, error) {
	apiID, view := httpx.StrToInt(body["api_id"]), httpx.StrToInt(body["view_status"])
	exists, err := s.repo.CheckAPIExist(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return httpx.Reply(400, httpx.MsgAPINotFound), nil
	}
	if view < 0 || view > 2 {
		view = 2
	}
	return httpx.Success(), s.repo.CheckAppChange(ctx, apiID, view)
}
