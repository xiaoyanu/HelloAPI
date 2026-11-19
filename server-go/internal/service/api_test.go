package service

import (
	"encoding/json"
	"testing"
)

func TestRequestParamsUsesJavaNestedNumberConversions(t *testing.T) {
	rows := []any{
		map[string]any{"name": "first", "required": "1.0", "type": "string", "msg": "保留第一项"},
		map[string]any{"name": "first", "required": false, "type": "", "msg": ""},
		map[string]any{"name": json.Number("1e0"), "required": true, "type": "number", "msg": "数量"},
	}
	params, message, err := requestParams(rows)
	if err != nil || message != "" || len(params) != 2 {
		t.Fatalf("参数 = %v，消息 = %q，错误 = %v", params, message, err)
	}
	if params[0].Required != 1 || *params[0].Msg != "保留第一项" ||
		*params[1].Name != "1.0" || params[1].Required != 1 {
		t.Fatalf("嵌套转换不符：%+v", params)
	}

	rows[0].(map[string]any)["required"] = json.Number("4294967297")
	if _, _, err := requestParams(rows); err == nil {
		t.Fatal("接受了超过 Java 参数整数范围的值")
	}
}
