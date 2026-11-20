package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zxz/helloapi/internal/httpx"
	"github.com/zxz/helloapi/internal/model"
)

type StatRepository interface {
	CheckUserKeyExists(ctx context.Context, key *string) (int64, error)
	GetUserIDByUserKey(ctx context.Context, key *string) (int, error)
	CheckAPIExist(ctx context.Context, apiID int) (int64, error)
	CheckAPIViewStatus(ctx context.Context, apiID int) (int64, error)
	GetAPIKey(ctx context.Context, key string) (*model.APIKey, error)
	ReduceAPIKeyCount(ctx context.Context, key string) (int64, error)
	SaveAPILog(context.Context, int, *string, *string, *string, *string, int) error
}

type Stat struct {
	repo            StatRepository
	maxPayloadBytes int
}

func NewStat(store StatRepository, maxPayloadBytes int) *Stat {
	return &Stat{repo: store, maxPayloadBytes: maxPayloadBytes}
}

func (s *Stat) LogAPI(ctx context.Context, body map[string]string) (response httpx.Response, failure error) {
	defer func() {
		if failure != nil {
			response, failure = httpx.Reply(400, "错误", httpx.MsgParam), nil
		}
	}()
	userKey := optional(body, "user_key")
	exists, err := s.repo.CheckUserKeyExists(ctx, userKey)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return httpx.Reply(400, "调用失败，请联系站长/开发者 代码：ERROR_1", "user_key用户密钥不存在"), nil
	}
	userID := httpx.StrToInt(body["user_id"])
	verifiedID, err := s.repo.GetUserIDByUserKey(ctx, userKey)
	if err != nil {
		return httpx.Response{}, err
	}
	if verifiedID != userID {
		return httpx.Reply(400, "调用失败，请联系站长/开发者 代码：ERROR_2", "user_id与user_key不匹配"), nil
	}
	apiID := httpx.StrToInt(body["api_id"])
	exists, err = s.repo.CheckAPIExist(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	if exists < 1 {
		return httpx.Reply(400, "调用失败，请联系站长/开发者 代码：ERROR_3", "api_id不存在"), nil
	}
	approved, err := s.repo.CheckAPIViewStatus(ctx, apiID)
	if err != nil {
		return httpx.Response{}, err
	}
	if approved < 1 {
		return httpx.Reply(400, "调用失败，请联系站长/开发者 代码：ERROR_4", "api_id未过审"), nil
	}

	apiKey := optional(body, "api_key")
	if apiKey != nil && *apiKey != "" {
		key, err := s.repo.GetAPIKey(ctx, *apiKey)
		if err != nil {
			return httpx.Response{}, err
		}
		if key == nil {
			return httpx.Reply(400, "API密钥 / APIKey不存在", "api_key不存在"), nil
		}
		if key.APIID != apiID {
			return httpx.Reply(400, "API密钥 / APIKey错误", "api_id与api_key不匹配"), nil
		}
		if key.Type == 0 {
			// Java 对存量期限密钥的空到期时间走异常分支。
			if key.Expired == nil {
				return httpx.Reply(400, "错误", httpx.MsgParam), nil
			}
			if time.Now().After(key.Expired.Time) {
				return httpx.Reply(400, "API密钥 / APIKey过期", httpx.MsgKeyExpired), nil
			}
		} else {
			affected, err := s.repo.ReduceAPIKeyCount(ctx, *apiKey)
			if err != nil {
				return httpx.Response{}, err
			}
			if affected == 0 {
				return httpx.Reply(400, "API密钥 / APIKey耗尽", httpx.MsgKeyExhausted), nil
			}
		}
	}

	header, payload := s.logPayload(body["header"]), s.logPayload(body["body"])
	return httpx.Success(), s.repo.SaveAPILog(ctx, apiID, optional(body, "ip"), header, payload, apiKey, userID)
}

func (s *Stat) logPayload(value string) *string {
	if s.maxPayloadBytes == 0 || !json.Valid([]byte(value)) {
		return nil
	}
	if s.maxPayloadBytes > 0 && len(value) > s.maxPayloadBytes {
		value = fmt.Sprintf(`{"truncated":true,"size":%d}`, len(value))
	}
	return &value
}
