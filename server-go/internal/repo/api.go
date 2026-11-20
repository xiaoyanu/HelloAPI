package repo

import (
	"context"
	"github.com/zxz/helloapi/internal/model"
)

func (s *Store) GetAPIList(ctx context.Context, pageSize int, offset int) ([]model.APIApp, error) {
	const query = "select * from helloapi_api_apps where `view_status` = 0 AND `status` != 3 ORDER BY created DESC LIMIT ? OFFSET ?"
	return selectRows[model.APIApp](ctx, s.query, query, pageSize, offset)
}

func (s *Store) GetAPIListAllCount(ctx context.Context) (int64, error) {
	const query = "select count(*) from helloapi_api_apps where `view_status` = 0  AND `status` != 3"
	return scalar[int64](ctx, s.query, query)
}

func (s *Store) GetAPIParam(ctx context.Context, apiID int) ([]model.APIParam, error) {
	const query = "select * from helloapi_api_params where api_id = ?"
	return selectRows[model.APIParam](ctx, s.query, query, apiID)
}

func (s *Store) CreateAPIApp(ctx context.Context, apiApp *model.APIApp) (int64, error) {
	const query = "insert into helloapi_api_apps" +
		" (title,smallTitle,status,type,url,sendtype,returnType,returnContent,user_id,view_status) values (?, ?, ?," +
		" ?, ?, ?, ?, ?, ?,?)"
	result, err := s.query.ExecContext(ctx, query, apiApp.Title, apiApp.SmallTitle, apiApp.Status, apiApp.Type, apiApp.URL, apiApp.SendType, apiApp.ReturnType, apiApp.ReturnContent, apiApp.UserID, apiApp.ViewStatus)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	apiApp.ID = int(id)
	return result.RowsAffected()
}

func (s *Store) CreateAPIParam(ctx context.Context, apiParam *model.APIParam) (int64, error) {
	const query = "insert into helloapi_api_params (api_id, name, required, type, msg) values (?, ?, ?, ?, ?)"
	return affected(ctx, s.query, query, apiParam.APIID, apiParam.Name, apiParam.Required, apiParam.Type, apiParam.Msg)
}

func (s *Store) CheckAPIExist(ctx context.Context, apiID int) (int64, error) {
	const query = "select count(*) from helloapi_api_apps where id = ?"
	return scalar[int64](ctx, s.query, query, apiID)
}

func (s *Store) CheckAPICreator(ctx context.Context, apiID int, userID int) (int64, error) {
	const query = "select count(*) from helloapi_api_apps where id = ? and user_id = ?"
	return scalar[int64](ctx, s.query, query, apiID, userID)
}

func (s *Store) UpdateAPIApp(ctx context.Context, apiApp *model.APIApp) (int64, error) {
	const query = "update helloapi_api_apps set `view_status` = ?, title = ?, smallTitle = ?, status = ?, type = ?, url = ?," +
		" sendType = ?, returnType = ?, returnContent = ? where id = ?"
	return affected(ctx, s.query, query, apiApp.ViewStatus, apiApp.Title, apiApp.SmallTitle, apiApp.Status, apiApp.Type, apiApp.URL, apiApp.SendType, apiApp.ReturnType, apiApp.ReturnContent, apiApp.ID)
}

func (s *Store) UpdateAPIParam(ctx context.Context, apiParam *model.APIParam) (int64, error) {
	const query = "update helloapi_api_params set name = ?, required = ?, type = ?, msg = ? where api_id = ?"
	return affected(ctx, s.query, query, apiParam.Name, apiParam.Required, apiParam.Type, apiParam.Msg, apiParam.APIID)
}

func (s *Store) CheckAPIParamExistObj(ctx context.Context, apiParam *model.APIParam) (int64, error) {
	const query = "select count(*) from helloapi_api_params where api_id = ? and name = ?"
	return scalar[int64](ctx, s.query, query, apiParam.APIID, apiParam.Name)
}

func (s *Store) DeleteAPIParamAll(ctx context.Context, apiID int) error {
	const query = "delete from helloapi_api_params where api_id = ?"
	return execute(ctx, s.query, query, apiID)
}

func (s *Store) DeleteAPIApp(ctx context.Context, apiID int) error {
	const query = "delete from helloapi_api_apps where id = ?"
	return execute(ctx, s.query, query, apiID)
}

func (s *Store) SearchAPIList(ctx context.Context, keyword string, pageSize int, offset int) ([]model.APIApp, error) {
	const query = "select * from helloapi_api_apps where `view_status` = 0 AND `status` != 3 AND (`title` LIKE" +
		" CONCAT('%',?,'%') OR `smalltitle` LIKE CONCAT('%',?,'%')) ORDER BY created DESC LIMIT ? OFFSET ?"
	return selectRows[model.APIApp](ctx, s.query, query, keyword, keyword, pageSize, offset)
}

func (s *Store) SearchAPIListCount(ctx context.Context, keyword string) (int64, error) {
	const query = "select count(*) from helloapi_api_apps where `view_status` = 0 AND `status` != 3 AND (`title` LIKE" +
		" CONCAT('%',?,'%') OR `smalltitle` LIKE CONCAT('%',?,'%'))"
	return scalar[int64](ctx, s.query, query, keyword, keyword)
}

func (s *Store) GetAPIApp(ctx context.Context, apiID int) (*model.APIApp, error) {
	const query = "select * from helloapi_api_apps where id = ?"
	return getRow[model.APIApp](ctx, s.query, query, apiID)
}

func (s *Store) CreateAPIKey(ctx context.Context, apiID int, key string, kind int, started *model.LocalDateTime, expired *model.LocalDateTime, count int64, desc *string) error {
	const query = "insert into helloapi_api_keys (`api_id`, `key` , `type`, `started`,`expired`,`count`,`desc`) values (?, ?," +
		" ?, ?, ?, ?, ?)"
	return execute(ctx, s.query, query, apiID, key, kind, started, expired, count, desc)
}

func (s *Store) CheckAPIKeyExist(ctx context.Context, key string) (int64, error) {
	const query = "select count(*) from helloapi_api_keys where `key` = ?"
	return scalar[int64](ctx, s.query, query, key)
}

func (s *Store) ReduceAPIKeyCount(ctx context.Context, key string) (int64, error) {
	const query = "update helloapi_api_keys set count = count - 1 where `key` = ? AND count > 0"
	return affected(ctx, s.query, query, key)
}

func (s *Store) GetAPIKey(ctx context.Context, key string) (*model.APIKey, error) {
	const query = "select * from helloapi_api_keys where `key` = ?"
	return getRow[model.APIKey](ctx, s.query, query, key)
}

func (s *Store) GetAPIKeyList(ctx context.Context, userID int, pageSize int, offset int) ([]model.APIKey, error) {
	const query = "select * from helloapi_api_keys where api_id in (select id from helloapi_api_apps where user_id = ?) ORDER" +
		" BY created DESC LIMIT ? OFFSET ?"
	return selectRows[model.APIKey](ctx, s.query, query, userID, pageSize, offset)
}

func (s *Store) GetAPIKeyListAllCount(ctx context.Context, userID int) (int64, error) {
	const query = "select count(*) from helloapi_api_keys where api_id in (select id from helloapi_api_apps where user_id = ?)"
	return scalar[int64](ctx, s.query, query, userID)
}

func (s *Store) GetAPIIDByKey(ctx context.Context, key string) (int, error) {
	const query = "select api_id from helloapi_api_keys where `key` = ?"
	return scalar[int](ctx, s.query, query, key)
}

func (s *Store) UpdateAPIKey(ctx context.Context, key string, kind int, started *model.LocalDateTime, expired *model.LocalDateTime, count int64, desc *string) error {
	const query = "update helloapi_api_keys set `type` = ?, `started` = ?, `expired` = ?, `count` = ?, `desc` = ? where `key` = ?"
	return execute(ctx, s.query, query, kind, started, expired, count, desc, key)
}

func (s *Store) DeleteAPIKey(ctx context.Context, key string) error {
	const query = "delete from helloapi_api_keys where `key` = ?"
	return execute(ctx, s.query, query, key)
}

func (s *Store) DeleteAPIKeyAll(ctx context.Context, apiID int) error {
	const query = "delete from helloapi_api_keys where `api_id` = ?"
	return execute(ctx, s.query, query, apiID)
}

func (s *Store) ResetAPIKey(ctx context.Context, key string, newKey string) error {
	const query = "update helloapi_api_keys set `key` = ? where `key` = ?"
	return execute(ctx, s.query, query, newKey, key)
}

func (s *Store) APIKeyListSearch(ctx context.Context, userID int, keyword *string, kind int, status int, pageSize int, offset int) ([]model.APIKey, error) {
	const query = "SELECT * FROM helloapi_api_keys WHERE api_id IN (SELECT id FROM helloapi_api_apps WHERE user_id = ?) AND" +
		" (? = '' OR `key` = ? OR `desc` LIKE CONCAT('%',?,'%') OR `api_id` = ?) AND (? = -1 OR `type` = ?) AND (?" +
		" = -1      OR (? = 0 AND ((`type` = 0 AND (expired IS NULL OR expired > NOW())) OR (`type` = 1 AND `count`" +
		" > 0)))      OR (? = 1 AND ((`type` = 0 AND expired IS NOT NULL AND expired <= NOW()) OR (`type` = 1 AND" +
		" `count` <= 0)))) ORDER BY created DESC LIMIT ? OFFSET ?"
	return selectRows[model.APIKey](ctx, s.query, query, userID, keyword, keyword, keyword, keyword, kind, kind, status, status, status, pageSize, offset)
}

func (s *Store) APIKeyListSearchCount(ctx context.Context, userID int, keyword *string, kind int, status int) (int64, error) {
	const query = "SELECT count(*) FROM helloapi_api_keys WHERE api_id IN (SELECT id FROM helloapi_api_apps WHERE user_id =" +
		" ?) AND (? = '' OR `key` = ? OR `desc` LIKE CONCAT('%',?,'%') OR `api_id` = ?) AND (? = -1 OR `type` = ?)" +
		" AND (? = -1      OR (? = 0 AND ((`type` = 0 AND (expired IS NULL OR expired > NOW())) OR (`type` = 1 AND" +
		" `count` > 0)))      OR (? = 1 AND ((`type` = 0 AND expired IS NOT NULL AND expired <= NOW()) OR (`type` =" +
		" 1 AND `count` <= 0)))) ORDER BY created DESC"
	return scalar[int64](ctx, s.query, query, userID, keyword, keyword, keyword, keyword, kind, kind, status, status, status)
}

func (s *Store) CheckAPIViewStatus(ctx context.Context, apiID int) (int64, error) {
	const query = "select count(*) from helloapi_api_apps where id = ? and view_status = 0"
	return scalar[int64](ctx, s.query, query, apiID)
}

// DeleteAPIData 按 Java 单独删除 API 时的顺序，在同一事务内清理关联数据。
func (s *Store) DeleteAPIData(ctx context.Context, apiID int) error {
	return s.inTx(ctx, func(tx *Store) error {
		for _, remove := range []func(context.Context, int) error{
			tx.DeleteAPIApp, tx.DeleteAPIParamAll, tx.DeleteAPICount,
			tx.DeleteAPILog, tx.DeleteAPIDailyStats, tx.DeleteAPIKeyAll,
		} {
			if err := remove(ctx, apiID); err != nil {
				return err
			}
		}
		return nil
	})
}

// APIWriter 是创建/更新 API 时需要在同一事务内执行的写操作集合。
type APIWriter interface {
	CreateAPIApp(context.Context, *model.APIApp) (int64, error)
	CreateAPIParam(context.Context, *model.APIParam) (int64, error)
	UpdateAPIApp(context.Context, *model.APIApp) (int64, error)
	UpdateAPIParam(context.Context, *model.APIParam) (int64, error)
	CheckAPIParamExistObj(context.Context, *model.APIParam) (int64, error)
	DeleteAPIParamAll(context.Context, int) error
}

// WriteAPIInTx 在单个事务中执行 fn；fn 返回错误时整体回滚，避免留下主记录与参数不一致的数据。
func (s *Store) WriteAPIInTx(ctx context.Context, fn func(APIWriter) error) error {
	return s.inTx(ctx, func(tx *Store) error { return fn(tx) })
}
