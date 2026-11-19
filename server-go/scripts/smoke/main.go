package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/png"
	"io"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type response struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

type smoke struct {
	client *http.Client
	base   string
	dir    string
	input  *bufio.Scanner
	token  string
}

func main() {
	base := flag.String("base", "http://127.0.0.1:8080", "后端 HTTP 地址")
	dir := flag.String("captcha-dir", ".", "验证码图片所在目录")
	adminUser := flag.String("admin-user", "", "已有管理员账号，可用于复跑")
	adminPassword := flag.String("admin-password", "", "已有管理员密码，也可用 HELLOAPI_SMOKE_ADMIN_PASSWORD")
	flag.Parse()
	if *adminPassword == "" {
		*adminPassword = os.Getenv("HELLOAPI_SMOKE_ADMIN_PASSWORD")
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		fail(err)
	}
	run := smoke{
		client: &http.Client{Jar: jar, Timeout: 10 * time.Second},
		base:   strings.TrimRight(*base, "/"), dir: *dir, input: bufio.NewScanner(os.Stdin),
	}
	if err := run.verify(*adminUser, *adminPassword); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "冒烟失败", err)
	os.Exit(1)
}

func (s *smoke) call(method, path string, body any, token string) (response, error) {
	var result response
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return result, err
		}
		reader = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, s.base+"/api/v1"+path, reader)
	if err != nil {
		return result, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return result, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return result, fmt.Errorf("%s %s 返回 HTTP %d", method, path, res.StatusCode)
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("%s %s 解码响应: %w", method, path, err)
	}
	return result, nil
}

func (s *smoke) ok(method, path string, body, out any) error {
	result, err := s.call(method, path, body, s.token)
	if err != nil {
		return err
	}
	if result.Code != 200 {
		return fmt.Errorf("%s %s 返回业务码 %d，%s", method, path, result.Code, result.Msg)
	}
	if out != nil {
		if err := json.Unmarshal(result.Data, out); err != nil {
			return fmt.Errorf("%s %s 解码 data: %w", method, path, err)
		}
	}
	return nil
}

func (s *smoke) captcha(step string) (string, error) {
	res, err := s.client.Get(s.base + "/api/v1/user/Captcha?" + strconv.FormatInt(time.Now().UnixMilli(), 10))
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return "", err
	}
	media := res.Header.Get("Content-Type")
	if res.StatusCode != http.StatusOK || (media != "image/png" && media != "image/gif") || res.Header.Get("Cache-Control") == "" {
		return "", fmt.Errorf("验证码响应 HTTP %d，媒体类型 %q，缓存控制 %q", res.StatusCode, media, res.Header.Get("Cache-Control"))
	}
	config, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	if config.Width != 120 || config.Height != 32 {
		return "", fmt.Errorf("验证码尺寸 %d×%d", config.Width, config.Height)
	}

	file, err := os.CreateTemp(s.dir, "helloapi-"+step+"-*."+strings.TrimPrefix(media, "image/"))
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())
	_, writeErr := file.Write(raw)
	if err := errors.Join(writeErr, file.Close()); err != nil {
		return "", err
	}
	path, err := filepath.Abs(file.Name())
	if err != nil {
		return "", err
	}
	fmt.Printf("已获取%s验证码（120×32），请打开 %s 并输入四位字符\n", step, path)
	if !s.input.Scan() {
		if err := s.input.Err(); err != nil {
			return "", err
		}
		return "", errors.New("未输入验证码")
	}
	return strings.TrimSpace(s.input.Text()), nil
}

type user struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Mode     int    `json:"mode"`
}

func (s *smoke) login(username, password string) (user, error) {
	code, err := s.captcha("登录")
	if err != nil {
		return user{}, err
	}
	result, err := s.call("POST", "/user/Login", map[string]string{"username": username, "password": password, "code": code}, "")
	if err != nil {
		return user{}, err
	}
	if result.Code != 200 {
		return user{}, fmt.Errorf("登录返回业务码 %d，%s", result.Code, result.Msg)
	}
	var data struct {
		Token string `json:"token"`
		User  user   `json:"user"`
	}
	if err := json.Unmarshal(result.Data, &data); err != nil {
		return user{}, err
	}
	if data.Token == "" || data.User.ID < 1 {
		return user{}, errors.New("登录响应缺少用户或令牌")
	}
	s.token = data.Token
	return data.User, nil
}

func (s *smoke) verify(adminUser, adminPassword string) error {
	suffix := strconv.FormatInt(time.Now().UnixNano(), 36)
	username := "smoke_" + suffix
	password := "HelloapiSmoke-" + suffix
	code, err := s.captcha("注册")
	if err != nil {
		return err
	}
	if err := s.ok("POST", "/user/Register", map[string]string{
		"username": username, "password": password, "code": code, "mail": username + "@example.com",
	}, nil); err != nil {
		return err
	}
	fmt.Println("注册通过")
	account, err := s.login(username, password)
	if err != nil {
		return err
	}
	var info user
	if err := s.ok("GET", "/user/", nil, &info); err != nil {
		return err
	}
	if info != account || info.Username != username {
		return fmt.Errorf("用户信息与登录结果不一致：%+v", info)
	}
	fmt.Printf("登录与用户信息通过，用户 ID %d，权限 %d\n", info.ID, info.Mode)

	var userKey string
	if err := s.ok("PUT", "/user/key/"+strconv.Itoa(info.ID), nil, &userKey); err != nil {
		return err
	}
	title := "冒烟接口-" + suffix
	appBody := map[string]any{
		"title": title, "smallTitle": "验证 API 生命周期", "type": 0, "status": 0,
		"url": "https://example.com/test", "sendType": 0, "returnType": "JSON", "returnContent": "{}",
		"view_status": 2, "params": []map[string]any{{"name": "page", "required": 0, "type": "number", "msg": "页码"}},
	}
	if err := s.ok("POST", "/app/", appBody, nil); err != nil {
		return err
	}
	var apps struct {
		List []struct {
			ID         int    `json:"id"`
			Title      string `json:"title"`
			ViewStatus int    `json:"view_status"`
		} `json:"list"`
	}
	if err := s.ok("GET", "/user/AppList", nil, &apps); err != nil {
		return err
	}
	var apiID int
	for _, app := range apps.List {
		if app.Title == title && app.ViewStatus == 2 {
			apiID = app.ID
			break
		}
	}
	if apiID == 0 {
		return errors.New("用户 API 列表没有新建的待审核 API")
	}
	ownerToken := s.token
	if info.Mode != 1 {
		if adminUser == "" || adminPassword == "" {
			return errors.New("当前用户是普通用户，请使用空库或提供已有管理员凭证")
		}
		admin, err := s.login(adminUser, adminPassword)
		if err != nil {
			return err
		}
		if admin.Mode != 1 {
			return errors.New("提供的账号没有管理员权限")
		}
	}
	if err := s.ok("PUT", "/user/CheckApp", map[string]int{"api_id": apiID, "view_status": 0}, nil); err != nil {
		return err
	}
	s.token = ownerToken
	apiPath := "/app/" + strconv.Itoa(apiID)
	var detail struct {
		ID         int `json:"id"`
		ViewStatus int `json:"view_status"`
		UserID     int `json:"user_id"`
		Params     []struct {
			Name string `json:"name"`
		} `json:"params"`
	}
	if err := s.ok("GET", apiPath, nil, &detail); err != nil {
		return err
	}
	if detail.ID != apiID || detail.ViewStatus != 0 || detail.UserID != info.ID || len(detail.Params) != 1 || detail.Params[0].Name != "page" {
		return fmt.Errorf("API 详情与创建及审核结果不一致：%+v", detail)
	}
	fmt.Printf("创建、审核和查询 API 通过，API ID %d\n", apiID)

	const available = 3
	var apiKey string
	if err := s.ok("POST", "/app/key/"+strconv.Itoa(apiID), map[string]int{"type": 1, "count": available}, &apiKey); err != nil {
		return err
	}
	var keyInfo struct {
		APIID   int     `json:"api_id"`
		Count   int64   `json:"count"`
		Started *string `json:"started"`
		Expired *string `json:"expired"`
	}
	if err := s.ok("GET", "/app/key/"+apiKey, nil, &keyInfo); err != nil {
		return err
	}
	if keyInfo.APIID != apiID || keyInfo.Count != available || keyInfo.Started != nil || keyInfo.Expired != nil {
		return fmt.Errorf("API 密钥详情不符：%+v", keyInfo)
	}
	if err := s.checkTimedKey(apiID); err != nil {
		return err
	}

	logBody := map[string]any{
		"api_id": apiID, "api_key": apiKey, "user_id": info.ID, "user_key": userKey,
		"ip": "127.0.0.1", "header": `{"Accept":"application/json"}`, "body": "不是 JSON",
	}
	if err := s.ok("POST", "/stat/log", logBody, nil); err != nil {
		return err
	}
	if err := s.ok("POST", "/stat/log/", logBody, nil); err != nil {
		return err
	}
	if err := s.checkConcurrentLogs(logBody); err != nil {
		return err
	}
	if err := s.ok("GET", "/app/key/"+apiKey, nil, &keyInfo); err != nil {
		return err
	}
	if keyInfo.Count != 0 {
		return fmt.Errorf("密钥剩余次数应为零，实际为 %d", keyInfo.Count)
	}
	fmt.Println("创建 API Key 与日志上报通过，最后一次配额并发请求 1 个成功、7 个耗尽")

	if err := s.waitForStats(title, available); err != nil {
		return err
	}
	if err := s.ok("DELETE", apiPath, nil, nil); err != nil {
		return err
	}
	deleted, err := s.call("GET", apiPath, nil, "")
	if err != nil {
		return err
	}
	if deleted.Code != 400 || deleted.Msg != "API不存在" {
		return fmt.Errorf("删除后获取 API 返回 %+v", deleted)
	}
	fmt.Println("删除 API 通过，详情返回 API不存在")
	return nil
}

func (s *smoke) checkTimedKey(apiID int) error {
	start := time.Now().Add(-time.Hour)
	end := start.Add(2 * time.Hour)
	const isoLayout = "2006-01-02T15:04:05"
	var key string
	if err := s.ok("POST", "/app/key/"+strconv.Itoa(apiID), map[string]any{
		"type": 0, "count": 0, "started": start.Format(isoLayout), "expired": end.Format(time.DateTime),
	}, &key); err != nil {
		return err
	}
	var data struct {
		Type    int    `json:"type"`
		Started string `json:"started"`
		Expired string `json:"expired"`
	}
	if err := s.ok("GET", "/app/key/"+key, nil, &data); err != nil {
		return err
	}
	if data.Type != 0 || data.Started != start.Format(isoLayout) || data.Expired != end.Format(isoLayout) {
		return fmt.Errorf("期限密钥时间往返不符：%+v", data)
	}
	fmt.Println("期限型 API Key 两种时间输入与秒精度输出通过")
	return nil
}

func (s *smoke) checkConcurrentLogs(body map[string]any) error {
	const requests = 8
	replies := make(chan error, requests)
	start := make(chan struct{})
	for range requests {
		go func() {
			<-start
			result, err := s.call("POST", "/stat/log", body, "")
			if err == nil {
				switch {
				case result.Code == 200:
				case result.Code == 400 && result.Msg == "API密钥 / APIKey耗尽":
					err = errExhausted
				default:
					err = fmt.Errorf("并发上报收到意外响应：%+v", result)
				}
			}
			replies <- err
		}()
	}
	close(start)
	var succeeded, exhausted int
	for range requests {
		err := <-replies
		switch {
		case err == nil:
			succeeded++
		case errors.Is(err, errExhausted):
			exhausted++
		default:
			return err
		}
	}
	if succeeded != 1 || exhausted != requests-1 {
		return fmt.Errorf("并发上报成功 %d，耗尽 %d", succeeded, exhausted)
	}
	return nil
}

var errExhausted = errors.New("密钥次数已用完")

func (s *smoke) waitForStats(title string, expected int64) error {
	type statValue struct {
		Count int64 `json:"count"`
	}
	type dashboard struct {
		All   statValue `json:"apiAllCount"`
		Today []struct {
			Name  string `json:"name"`
			Count int64  `json:"count"`
		} `json:"apiTodayCountArray"`
		Week []struct {
			Date string `json:"date"`
		} `json:"apiWeekCountArray"`
	}
	deadline := time.Now().Add(50 * time.Second)
	for {
		var values dashboard
		if err := s.ok("GET", "/stat/dashboard", nil, &values); err != nil {
			return err
		}
		var today int64
		for _, value := range values.Today {
			if value.Name == title {
				today = value.Count
				break
			}
		}
		if today == expected && values.All.Count >= expected && len(values.Today) == 7 && len(values.Week) == 7 {
			var public statValue
			if err := s.ok("GET", "/stat/count", nil, &public); err != nil {
				return err
			}
			if public.Count >= expected {
				fmt.Printf("统计通过，当前接口今日 %d 次，Dashboard 累计 %d 次，公开累计 %d 次\n", today, values.All.Count, public.Count)
				return nil
			}
		}
		if !time.Now().Before(deadline) {
			return fmt.Errorf("等待聚合超时，当前接口今日 %d 次，累计 %d 次", today, values.All.Count)
		}
		time.Sleep(time.Second)
	}
}
