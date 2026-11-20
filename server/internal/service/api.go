package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"

	"github.com/zxz/helloapi/internal/auth"
	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/model"
	"github.com/zxz/helloapi/internal/repo"
)

type APIRepository interface {
	GetAPIList(ctx context.Context, pageSize int, offset int) ([]model.APIApp, error)
	GetAPIListAllCount(ctx context.Context) (int64, error)
	SearchAPIList(ctx context.Context, keyword string, pageSize int, offset int) ([]model.APIApp, error)
	SearchAPIListCount(ctx context.Context, keyword string) (int64, error)
	GetAPIParam(ctx context.Context, apiID int) ([]model.APIParam, error)
	CheckAPIExist(ctx context.Context, apiID int) (int64, error)
	CheckAPICreator(ctx context.Context, apiID int, userID int) (int64, error)
	GetAPIApp(ctx context.Context, apiID int) (*model.APIApp, error)
	CreateAPIKey(ctx context.Context, apiID int, key string, kind int, started *model.LocalDateTime, expired *model.LocalDateTime, count int64, desc *string) error
	CheckAPIKeyExist(ctx context.Context, key string) (int64, error)
	GetAPIKey(ctx context.Context, key string) (*model.APIKey, error)
	GetAPIKeyList(ctx context.Context, userID int, pageSize int, offset int) ([]model.APIKey, error)
	GetAPIKeyListAllCount(ctx context.Context, userID int) (int64, error)
	GetAPIIDByKey(ctx context.Context, key string) (int, error)
	UpdateAPIKey(ctx context.Context, key string, kind int, started *model.LocalDateTime, expired *model.LocalDateTime, count int64, desc *string) error
	DeleteAPIKey(ctx context.Context, key string) error
	ResetAPIKey(ctx context.Context, key string, newKey string) error
	APIKeyListSearch(ctx context.Context, userID int, keyword *string, kind int, status int, pageSize int, offset int) ([]model.APIKey, error)
	APIKeyListSearchCount(ctx context.Context, userID int, keyword *string, kind int, status int) (int64, error)
	CheckUserIDExists(ctx context.Context, userID int) (int64, error)
	settingReader
	DeleteAPIData(context.Context, int) error
	WriteAPIInTx(context.Context, func(repo.APIWriter) error) error
}

// errAPIParamRejected 表示某条参数未写入成功，用于触发事务回滚。
var errAPIParamRejected = errors.New("API 参数写入失败")

type API struct {
	repo APIRepository
	jwt  *auth.JWT
}

func NewAPI(store APIRepository, jwt *auth.JWT) *API { return &API{repo: store, jwt: jwt} }

type homeItem struct {
	ID         int     `json:"id"`
	Title      *string `json:"title"`
	SmallTitle *string `json:"smallTitle"`
	Status     int     `json:"status"`
	Type       int     `json:"type"`
}

func homeList(apps []model.APIApp) []homeItem {
	items := make([]homeItem, 0, len(apps))
	for _, app := range apps {
		items = append(items, homeItem{ID: app.ID, Title: app.Title, SmallTitle: app.SmallTitle, Status: app.Status, Type: app.Type})
	}
	return items
}

func (s *API) GetAPIList(ctx context.Context, query map[string]string) (httpx.Response, error) {
	size, offset := pagination(query)
	total, err := s.repo.GetAPIListAllCount(ctx)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.GetAPIList(ctx, size, offset)
	return httpx.Reply(200, "获取成功", pageResult(total, homeList(list))), err
}

func (s *API) SearchAPIList(ctx context.Context, query map[string]string) (httpx.Response, error) {
	keyword := query["keywords"]
	if keyword == "" {
		return paramError(), nil
	}
	size, offset := pagination(query)
	total, err := s.repo.SearchAPIListCount(ctx, keyword)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.SearchAPIList(ctx, keyword, size, offset)
	return httpx.Success(pageResult(total, homeList(list))), err
}

func (s *API) CreateAPI(ctx context.Context, body map[string]any) (response httpx.Response, err error) {
	defer paramOnError(&response, &err)
	identity := httpx.User(ctx)
	setting, err := s.repo.GetSettingValue(ctx, "api")
	if err != nil {
		return httpx.Response{}, err
	}
	if identity.Mode != httpx.Admin && disabled(setting) {
		return httpx.Reply(400, "发布API已关闭"), nil
	}

	app, err := requestApp(body, identity.ID)
	if err != nil {
		return httpx.Response{}, err
	}
	exists, err := s.repo.CheckUserIDExists(ctx, identity.ID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	if app.ViewStatus != 2 && identity.Mode != httpx.Admin {
		return httpx.Reply(400, httpx.MsgNotAdmin), nil
	}
	params, msg, err := requestParams(body["params"])
	if err != nil {
		return httpx.Response{}, err
	}
	if msg != "" {
		return httpx.Reply(400, msg), nil
	}

	// 主记录与参数在同一事务中写入，任一步失败整体回滚（由 paramOnError 转为参数错误响应）。
	err = s.repo.WriteAPIInTx(ctx, func(tx repo.APIWriter) error {
		affected, err := tx.CreateAPIApp(ctx, &app)
		if err != nil || affected < 1 {
			return err
		}
		for _, param := range params {
			param.APIID = app.ID
			affected, err := tx.CreateAPIParam(ctx, &param)
			if err != nil {
				return err
			}
			if affected < 1 {
				return errAPIParamRejected
			}
		}
		return nil
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Success(), nil
}

func (s *API) validateAPIWriter(ctx context.Context, id int) (httpx.Response, error) {
	exists, err := s.repo.CheckAPIExist(ctx, id)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return httpx.Reply(400, httpx.MsgAPINotFound), nil
	}
	identity := httpx.User(ctx)
	userExists, err := s.repo.CheckUserIDExists(ctx, identity.ID)
	if err != nil {
		return httpx.Response{}, err
	}
	if userExists < 1 {
		return noUser(), nil
	}
	return s.requireCreator(ctx, id)
}

func (s *API) requireCreator(ctx context.Context, apiID int) (httpx.Response, error) {
	identity := httpx.User(ctx)
	count, err := s.repo.CheckAPICreator(ctx, apiID, identity.ID)
	if err != nil {
		return httpx.Response{}, err
	}
	if count < 1 && identity.Mode != httpx.Admin {
		return notCreator(), nil
	}
	return httpx.Response{}, nil
}

func (s *API) UpdateAPI(ctx context.Context, id string, body map[string]any) (response httpx.Response, err error) {
	defer paramOnError(&response, &err)
	apiID := httpx.StrToInt(id)
	if rejection, err := s.validateAPIWriter(ctx, apiID); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	identity := httpx.User(ctx)
	app, err := requestApp(body, identity.ID)
	if err != nil {
		return httpx.Response{}, err
	}
	if app.ViewStatus != 2 && identity.Mode != httpx.Admin {
		return httpx.Reply(400, httpx.MsgNotAdmin), nil
	}
	app.ID = apiID

	params, msg, err := requestParams(body["params"])
	if err != nil {
		return httpx.Response{}, err
	}
	if msg != "" {
		return httpx.Reply(400, msg), nil
	}
	// 删除旧参数、更新主记录、写入新参数在同一事务中完成，失败时旧数据保持不变。
	err = s.repo.WriteAPIInTx(ctx, func(tx repo.APIWriter) error {
		if err := tx.DeleteAPIParamAll(ctx, apiID); err != nil {
			return err
		}
		affected, err := tx.UpdateAPIApp(ctx, &app)
		if err != nil || affected < 1 {
			return err
		}
		for _, param := range params {
			param.APIID = apiID
			count, err := tx.CheckAPIParamExistObj(ctx, &param)
			if err != nil {
				return err
			}
			if count > 0 {
				affected, err = tx.UpdateAPIParam(ctx, &param)
			} else {
				affected, err = tx.CreateAPIParam(ctx, &param)
			}
			if err != nil {
				return err
			}
			if affected < 1 {
				return errAPIParamRejected
			}
		}
		return nil
	})
	if err != nil {
		return httpx.Response{}, err
	}
	return httpx.Success(), nil
}

func (s *API) DeleteAPI(ctx context.Context, id string) (response httpx.Response, err error) {
	defer paramOnError(&response, &err)
	apiID := httpx.StrToInt(id)
	if rejection, err := s.validateAPIWriter(ctx, apiID); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	return httpx.Success(), s.repo.DeleteAPIData(ctx, apiID)
}

type appVisibility struct {
	Status     int `json:"status"`
	ViewStatus int `json:"view_status"`
}

type detailParam struct {
	Name     *string `json:"name"`
	Required int     `json:"required"`
	Type     *string `json:"type"`
	Msg      *string `json:"msg"`
}

func (s *API) GetAPIApp(ctx context.Context, id, authorization string) (httpx.Response, error) {
	apiID := httpx.StrToInt(id)
	app, err := s.repo.GetAPIApp(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	if app == nil {
		return httpx.Reply(400, httpx.MsgAPINotFound), nil
	}
	if app.Status == 3 || app.ViewStatus != 0 {
		claims, _ := s.jwt.Bearer(authorization)
		if (claims.ID < 1 || app.UserID != claims.ID) && claims.Mode != httpx.Admin {
			return httpx.Success(appVisibility{Status: app.Status, ViewStatus: app.ViewStatus}), nil
		}
	}

	params, err := s.repo.GetAPIParam(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	list := make([]detailParam, 0, len(params))
	for _, param := range params {
		list = append(list, detailParam{Name: param.Name, Required: param.Required, Type: param.Type, Msg: param.Msg})
	}
	return httpx.Success(struct {
		ID            int           `json:"id"`
		Title         *string       `json:"title"`
		SmallTitle    *string       `json:"smallTitle"`
		Status        int           `json:"status"`
		Type          int           `json:"type"`
		URL           *string       `json:"url"`
		SendType      int           `json:"sendType"`
		ReturnType    *string       `json:"returnType"`
		ReturnContent *string       `json:"returnContent"`
		ViewStatus    int           `json:"view_status"`
		Params        []detailParam `json:"params"`
		UserID        int           `json:"user_id"`
	}{
		ID: app.ID, Title: app.Title, SmallTitle: app.SmallTitle, Status: app.Status, Type: app.Type, URL: app.URL,
		SendType: app.SendType, ReturnType: app.ReturnType, ReturnContent: app.ReturnContent,
		ViewStatus: app.ViewStatus, Params: list, UserID: app.UserID,
	}), nil
}

func (s *API) CreateAPIKey(ctx context.Context, id string, body map[string]any) (httpx.Response, error) {
	identity := httpx.User(ctx)
	setting, err := s.repo.GetSettingValue(ctx, "api_key")
	if err != nil {
		return httpx.Response{}, err
	}
	if identity.Mode != httpx.Admin && disabled(setting) {
		return httpx.Reply(400, "创建APIKey已关闭"), nil
	}
	apiID := httpx.StrToInt(id)
	exists, err := s.repo.CheckAPIExist(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return httpx.Reply(400, httpx.MsgAPINotFound), nil
	}
	if rejection, err := s.requireCreator(ctx, apiID); err != nil || rejection.Code != 0 {
		return rejection, err
	}

	keyData, valid, err := requestKey(body)
	if err != nil {
		return httpx.Response{}, err
	}
	if !valid {
		return paramError(), nil
	}
	key, err := s.availableKey(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	desc, err := jsonString(body, "desc")
	if err != nil {
		return httpx.Response{}, err
	}
	if desc != nil && textLength(*desc) > 245 {
		return httpx.Reply(400, "备注不能超过245个字符"), nil
	}
	key, err = s.writeUniqueKey(ctx, apiID, key, func(candidate string) error {
		return s.repo.CreateAPIKey(ctx, apiID, candidate, keyData.Type, keyData.Started, keyData.Expired, keyData.Count, desc)
	})
	return httpx.Success(key), err
}

// availableKey 延续 Java 的 ID 加毫秒时间戳后取 MD5 的规则。
func (s *API) availableKey(ctx context.Context, apiID int) (string, error) {
	for {
		key := keyFor(apiID)
		exists, err := s.repo.CheckAPIKeyExist(ctx, key)
		if err != nil {
			return "", err
		}
		if exists == 0 {
			return key, nil
		}
		timer := time.NewTimer(time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return "", ctx.Err()
		case <-timer.C:
		}
	}
}

func (s *API) writeUniqueKey(ctx context.Context, apiID int, key string, write func(string) error) (string, error) {
	for {
		err := write(key)
		if err == nil {
			return key, nil
		}
		var duplicate *mysql.MySQLError
		if !errors.As(err, &duplicate) || duplicate.Number != 1062 {
			return "", err
		}
		// 并发写入可能在查重成功之后抢占同一个密钥，由唯一索引判定并重试。
		key, err = s.availableKey(ctx, apiID)
		if err != nil {
			return "", err
		}
	}
}

func (s *API) verifyKeyOwner(ctx context.Context, key string) (int, httpx.Response, error) {
	exists, err := s.repo.CheckAPIKeyExist(ctx, key)
	if err != nil {
		return 0, httpx.Response{}, err
	}
	if exists < 1 {
		return 0, httpx.Reply(400, httpx.MsgKeyNotFound), nil
	}
	apiID, err := s.repo.GetAPIIDByKey(ctx, key)
	if err != nil {
		return 0, httpx.Response{}, err
	}
	rejection, err := s.requireCreator(ctx, apiID)
	return apiID, rejection, err
}

func (s *API) UpdateAPIKey(ctx context.Context, key string, body map[string]any) (httpx.Response, error) {
	if _, rejection, err := s.verifyKeyOwner(ctx, key); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	keyData, valid, err := requestKey(body)
	if err != nil {
		return httpx.Response{}, err
	}
	if !valid {
		return paramError(), nil
	}
	desc, err := jsonString(body, "desc")
	if err != nil {
		return httpx.Response{}, err
	}
	if desc != nil && textLength(*desc) > 245 {
		return httpx.Reply(400, "备注不能超过245个字符"), nil
	}
	return httpx.Success(), s.repo.UpdateAPIKey(ctx, key, keyData.Type, keyData.Started, keyData.Expired, keyData.Count, desc)
}

func (s *API) GetAPIKey(ctx context.Context, key string) (httpx.Response, error) {
	if _, rejection, err := s.verifyKeyOwner(ctx, key); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	data, err := s.repo.GetAPIKey(ctx, key)
	return httpx.Success(data), err
}

func (s *API) DeleteAPIKey(ctx context.Context, key string) (httpx.Response, error) {
	if _, rejection, err := s.verifyKeyOwner(ctx, key); err != nil || rejection.Code != 0 {
		return rejection, err
	}
	return httpx.Success(), s.repo.DeleteAPIKey(ctx, key)
}

func (s *API) ResetAPIKey(ctx context.Context, key string) (httpx.Response, error) {
	apiID, rejection, err := s.verifyKeyOwner(ctx, key)
	if err != nil || rejection.Code != 0 {
		return rejection, err
	}
	next, err := s.availableKey(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	next, err = s.writeUniqueKey(ctx, apiID, next, func(candidate string) error { return s.repo.ResetAPIKey(ctx, key, candidate) })
	return httpx.Success(next), err
}

func (s *API) GetUserAPIKeyList(ctx context.Context, id string, query map[string]string) (httpx.Response, error) {
	userID := httpx.StrToInt(id)
	identity := httpx.User(ctx)
	if userID != identity.ID && identity.Mode != httpx.Admin {
		return notCreator(), nil
	}
	exists, err := s.repo.CheckUserIDExists(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	size, offset := pagination(query)
	total, err := s.repo.GetAPIKeyListAllCount(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.GetAPIKeyList(ctx, userID, size, offset)
	return httpx.Success(pageResult(total, list)), err
}

func (s *API) UserAPIKeyListSearch(ctx context.Context, id string, query map[string]string) (httpx.Response, error) {
	userID := httpx.StrToInt(id)
	// Java 读取了拦截器没有设置的 tokenUserId 属性，转整数后固定为 -1。
	if userID != -1 && httpx.User(ctx).Mode != httpx.Admin {
		return notCreator(), nil
	}
	exists, err := s.repo.CheckUserIDExists(ctx, userID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return noUser(), nil
	}
	size, offset := pagination(query)
	keyword := optional(query, "keywords")
	kind, status := httpx.StrToInt(query["type"]), httpx.StrToInt(query["status"])
	total, err := s.repo.APIKeyListSearchCount(ctx, userID, keyword, kind, status)
	if err != nil {
		return httpx.Response{}, err
	}
	list, err := s.repo.APIKeyListSearch(ctx, userID, keyword, kind, status, size, offset)
	return httpx.Success(pageResult(total, list)), err
}

func jsonString(body map[string]any, key string) (*string, error) {
	value := body[key]
	if value == nil {
		return nil, nil
	}
	text, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("字段 %s 必须为字符串", key)
	}
	return &text, nil
}

func jsonInt(body map[string]any, key string) (int, error) {
	value, ok := body[key].(json.Number)
	if !ok {
		return 0, fmt.Errorf("字段 %s 必须为整数", key)
	}
	number, err := strconv.ParseInt(value.String(), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("字段 %s 必须为 32 位整数", key)
	}
	return int(number), nil
}

func requestApp(body map[string]any, userID int) (model.APIApp, error) {
	app := model.APIApp{UserID: userID}
	// 字段转换顺序与 Java 的 getRequestApiApp 一致。
	var err error
	if app.Title, err = jsonString(body, "title"); err != nil {
		return app, err
	}
	if app.SmallTitle, err = jsonString(body, "smallTitle"); err != nil {
		return app, err
	}
	if app.Type, err = jsonInt(body, "type"); err != nil {
		return app, err
	}
	if app.Status, err = jsonInt(body, "status"); err != nil {
		return app, err
	}
	if app.URL, err = jsonString(body, "url"); err != nil {
		return app, err
	}
	if app.SendType, err = jsonInt(body, "sendType"); err != nil {
		return app, err
	}
	if app.ReturnType, err = jsonString(body, "returnType"); err != nil {
		return app, err
	}
	if app.ReturnContent, err = jsonString(body, "returnContent"); err != nil {
		return app, err
	}
	app.ViewStatus, err = jsonInt(body, "view_status")
	return app, err
}

func requestKey(body map[string]any) (model.APIKey, bool, error) {
	var data model.APIKey
	kind, err := jsonInt(body, "type")
	if err != nil {
		return data, false, err
	}
	data.Type = kind

	startText, err := jsonString(body, "started")
	if err != nil {
		return data, false, err
	}
	if startText != nil && *startText != "" {
		endText, err := jsonString(body, "expired")
		if err != nil {
			return data, false, err
		}
		if endText != nil && *endText != "" {
			started, err := model.ParseLocalDateTime(*startText)
			if err != nil {
				return data, false, nil
			}
			expired, err := model.ParseLocalDateTime(*endText)
			if err != nil || !expired.After(started.Time) {
				return data, false, nil
			}
			data.Started, data.Expired = &started, &expired
		}
	}

	count, err := jsonInt(body, "count")
	if err != nil {
		return data, false, err
	}
	data.Count = int64(count)
	return data, !(kind == 0 && data.Started == nil || kind == 1 && count < 0), nil
}

func requestParams(input any) ([]model.APIParam, string, error) {
	if input == nil {
		return nil, "", nil
	}
	rows, ok := input.([]any)
	if !ok {
		return nil, "", errors.New("params 必须为数组")
	}
	params := make([]model.APIParam, 0, len(rows))
	type paramName struct {
		value   string
		present bool
	}
	names := make(map[paramName]struct{}, len(rows))
	for _, row := range rows {
		fields, ok := row.(map[string]any)
		if !ok {
			return nil, "", errors.New("参数项必须为对象")
		}
		required, err := paramRequired(fields["required"])
		if err != nil {
			return nil, "", err
		}
		param := model.APIParam{
			Name: paramText(fields["name"]), Required: required, Type: paramText(fields["type"]), Msg: paramText(fields["msg"]),
		}
		var key paramName
		if param.Name != nil {
			key.value, key.present = *param.Name, true
		}
		if _, exists := names[key]; exists {
			continue
		}
		names[key] = struct{}{}
		params = append(params, param)
	}
	for i, param := range params {
		var msg string
		switch {
		case param.Name == nil || trim(*param.Name) == "":
			msg = "参数名不能为空"
		case param.Required != 0 && param.Required != 1:
			msg = "必填项不能为空"
		case param.Type == nil || trim(*param.Type) == "":
			msg = "类型不能为空"
		case param.Msg == nil || trim(*param.Msg) == "":
			msg = "描述不能为空"
		case textLength(*param.Msg) > 245:
			msg = "描述不能超过245个字符"
		}
		if msg != "" {
			return nil, fmt.Sprintf("第 %d 行参数的%s", i+1, msg), nil
		}
	}
	return params, "", nil
}

func paramText(value any) *string {
	switch v := value.(type) {
	case nil:
		return nil
	case string:
		return &v
	case json.Number:
		return paramNumberText(v)
	default:
		raw, _ := json.Marshal(value)
		text := string(raw)
		return &text
	}
}

var trailingDecimalZeros = regexp.MustCompile(`\.0*$`)

func paramRequired(value any) (int, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case json.Number:
		if !strings.ContainsAny(v.String(), ".eE") {
			number, err := strconv.ParseInt(v.String(), 10, 32)
			return int(number), err
		}
		text := paramNumberText(v)
		if text == nil {
			return 0, nil
		}
		// Fastjson 将小数作为 BigDecimal，截取整数部分并保留低 32 位。
		decimal, _ := new(big.Rat).SetString(*text)
		integer := new(big.Int).Quo(decimal.Num(), decimal.Denom())
		lowBits := integer.And(integer, big.NewInt(math.MaxUint32))
		return int(int32(lowBits.Int64())), nil
	case string:
		if v == "" || v == "null" || v == "NULL" {
			return 0, nil
		}
		text := strings.ReplaceAll(v, ",", "")
		text = trailingDecimalZeros.ReplaceAllString(text, "")
		number, err := strconv.ParseInt(text, 10, 32)
		if err == nil {
			return int(number), nil
		}
	}
	return 0, errors.New("参数项的 required 无法转换为整数")
}

// paramNumberText 保留嵌套参数经 Jackson 和 Fastjson 转换成字符串的常见数字形式。
func paramNumberText(number json.Number) *string {
	text := number.String()
	if !strings.ContainsAny(text, ".eE") {
		if text == "-0" {
			text = "0"
		}
		return &text
	}
	value, err := number.Float64()
	if err != nil {
		return nil
	}
	if value == 0 {
		text = "0.0"
		return &text
	}
	abs := math.Abs(value)
	if abs >= 1e7 || abs < 1e-3 {
		mantissa, rawExponent, _ := strings.Cut(strconv.FormatFloat(value, 'E', -1, 64), "E")
		exponent, _ := strconv.Atoi(rawExponent)
		if !strings.Contains(mantissa, ".") {
			mantissa += ".0"
		}
		if exponent >= -6 && exponent < 0 {
			digits := strings.ReplaceAll(strings.TrimPrefix(mantissa, "-"), ".", "")
			text = "0." + strings.Repeat("0", -exponent-1) + digits
			if value < 0 {
				text = "-" + text
			}
		} else {
			text = fmt.Sprintf("%sE%+d", mantissa, exponent)
		}
	} else {
		text = strconv.FormatFloat(value, 'f', -1, 64)
		if !strings.Contains(text, ".") {
			text += ".0"
		}
	}
	return &text
}
