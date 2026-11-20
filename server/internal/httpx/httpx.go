package httpx

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

const (
	UserURL    = "/api/v1/user"
	APIURL     = "/api/v1/app"
	StatURL    = "/api/v1/stat"
	LinkURL    = "/api/v1/link"
	SettingURL = "/api/v1/setting"

	MsgTokenInvalid = "Token过期/无效"
	MsgNotAdmin     = "非管理员用户，没有操作权限"
	MsgParam        = "参数错误"
	MsgUserNotFound = "用户不存在"
	MsgNotCreator   = "您不是本人/管理员，没有操作权限"
	MsgAPINotFound  = "API不存在"
	MsgKeyNotFound  = "密钥不存在"
	MsgKeyExpired   = "密钥已过期"
	MsgKeyExhausted = "密钥次数已用完"
	MsgInternal     = "服务器内部错误"
	Admin           = 1
)

type Response struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func Reply(code int, msg string, data ...any) Response {
	response := Response{Code: code, Msg: msg}
	if len(data) == 0 || data[0] == nil {
		return response
	}
	value := reflect.ValueOf(data[0])
	switch value.Kind() {
	case reflect.Slice:
		if value.IsNil() {
			response.Data = reflect.MakeSlice(value.Type(), 0, 0).Interface()
			return response
		}
	case reflect.Pointer, reflect.Map:
		if value.IsNil() {
			return response
		}
	}
	response.Data = data[0]
	return response
}

func Success(data ...any) Response { return Reply(200, "成功", data...) }

func Write(w http.ResponseWriter, status int, response Response) {
	raw, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json;charset=UTF-8")
	w.WriteHeader(status)
	_, _ = w.Write(raw)
}

func ObjectMap(body io.Reader) (map[string]any, error) {
	if body == nil {
		return nil, io.EOF
	}
	decoder := json.NewDecoder(body)
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil {
		return nil, err
	}
	if object == nil {
		return nil, errors.New("请求体必须为 JSON 对象")
	}
	if err := decoder.Decode(new(any)); !errors.Is(err, io.EOF) {
		return nil, errors.New("JSON 对象之后存在多余数据")
	}
	return object, nil
}

func StringMap(body io.Reader) map[string]string {
	result := make(map[string]string)
	object, err := ObjectMap(body)
	if err != nil {
		return result
	}
	for key, value := range object {
		switch typed := value.(type) {
		case string:
			result[key] = typed
		case json.Number:
			result[key] = typed.String()
		case bool:
			result[key] = strconv.FormatBool(typed)
		case nil:
		default:
			return make(map[string]string)
		}
	}
	return result
}

func Query(r *http.Request) map[string]string {
	values := r.URL.Query()
	result := make(map[string]string, len(values))
	for key, list := range values {
		if len(list) > 0 {
			result[key] = list[0]
		}
	}
	return result
}

func StrToInt(value string) int {
	number, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return -1
	}
	return int(number)
}
