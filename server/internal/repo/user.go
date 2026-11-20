package repo

import (
	"context"
	"database/sql"
	"github.com/zxz/helloapi/internal/model"
)

func (s *Store) CheckUsernameExists(ctx context.Context, username string) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_users` WHERE `username` = ?"
	return scalar[int64](ctx, s.query, query, username)
}

func (s *Store) Register(ctx context.Context, username string, password string, nick string, mode int, mail string) (int64, error) {
	const query = "INSERT INTO `helloapi_users` (`username`, `password`,`nick`,`mode`,`mail`) VALUES (?, ?,?,?,?)"
	return affected(ctx, s.query, query, username, password, nick, mode, mail)
}

func (s *Store) Login(ctx context.Context, username string, password string) (*model.User, error) {
	const query = "SELECT * FROM `helloapi_users` WHERE `username` = ? AND `password` = ?"
	return unique[model.User](ctx, s.query, query, username, password)
}

func (s *Store) GetUserInfo(ctx context.Context, id int) (*model.User, error) {
	const query = "SELECT * FROM `helloapi_users` WHERE `id` = ?"
	return getRow[model.User](ctx, s.query, query, id)
}

func (s *Store) GetUserKey(ctx context.Context, userID int) (*model.UserKey, error) {
	const query = "SELECT * FROM `helloapi_user_keys` WHERE `user_id` = ?"
	return getRow[model.UserKey](ctx, s.query, query, userID)
}

func (s *Store) CheckUserKeyExists(ctx context.Context, key *string) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_user_keys` WHERE `key` = ?"
	return scalar[int64](ctx, s.query, query, key)
}

func (s *Store) DeleteUser(ctx context.Context, userID int) error {
	const query = "DELETE FROM `helloapi_users` WHERE `id` = ?"
	return execute(ctx, s.query, query, userID)
}

func (s *Store) DeleteUserKey(ctx context.Context, userID int) error {
	const query = "DELETE FROM `helloapi_user_keys` WHERE `user_id` = ?"
	return execute(ctx, s.query, query, userID)
}

func (s *Store) CheckUserIDExists(ctx context.Context, userID int) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_users` WHERE `id` = ?"
	return scalar[int64](ctx, s.query, query, userID)
}

func (s *Store) GetUserAPIList(ctx context.Context, userID int, pageSize int, offset int) ([]model.APIApp, error) {
	const query = "SELECT * FROM `helloapi_api_apps` WHERE `user_id` = ? ORDER BY created DESC LIMIT ? OFFSET ? "
	return selectRows[model.APIApp](ctx, s.query, query, userID, pageSize, offset)
}

func (s *Store) GetUserAPIListAll(ctx context.Context, userID int) ([]model.APIApp, error) {
	const query = "SELECT * FROM `helloapi_api_apps` WHERE `user_id` = ?"
	return selectRows[model.APIApp](ctx, s.query, query, userID)
}

func (s *Store) GetUserAPIListAllCount(ctx context.Context, userID int) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_api_apps` WHERE `user_id` = ?"
	return scalar[int64](ctx, s.query, query, userID)
}

func (s *Store) ResetUserKey(ctx context.Context, userID int, key string) error {
	const query = "INSERT INTO `helloapi_user_keys` (`user_id`, `key`) VALUES (?, ?) ON DUPLICATE KEY UPDATE `key` = ?"
	return execute(ctx, s.query, query, userID, key, key)
}

func (s *Store) GetUserCount(ctx context.Context) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_users`"
	return scalar[int64](ctx, s.query, query)
}

func (s *Store) GetUserAPIListSearch(ctx context.Context, userID int, keyword string, kind int, status int, viewStatus int, pageSize int, offset int) ([]model.APIApp, error) {
	const query = "SELECT * FROM `helloapi_api_apps` WHERE `user_id` = ? AND (? IS NULL OR ? = '' OR `url` LIKE" +
		" CONCAT('%',?,'%') OR `title` LIKE CONCAT('%', ?, '%') OR `id` = ? OR `smallTitle` LIKE CONCAT('%', ?," +
		" '%')) AND (? = -1 OR `type` = ?) AND (? = -1 OR `status` = ?) AND (? = -1 OR `view_status` = ?) ORDER BY" +
		" created DESC LIMIT ? OFFSET ?"
	return selectRows[model.APIApp](ctx, s.query, query, userID, keyword, keyword, keyword, keyword, keyword, keyword, kind, kind, status, status, viewStatus, viewStatus, pageSize, offset)
}

func (s *Store) GetUserAPIListSearchCount(ctx context.Context, userID int, keyword string, kind int, status int, viewStatus int) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_api_apps` WHERE `user_id` = ? AND (? IS NULL OR ? = '' OR `url` LIKE" +
		" CONCAT('%',?,'%') OR `title` LIKE CONCAT('%', ?, '%') OR `id` = ? OR `smallTitle` LIKE CONCAT('%', ?," +
		" '%')) AND (? = -1 OR `type` = ?) AND (? = -1 OR `status` = ?) AND (? = -1 OR `view_status` = ?)"
	return scalar[int64](ctx, s.query, query, userID, keyword, keyword, keyword, keyword, keyword, keyword, kind, kind, status, status, viewStatus, viewStatus)
}

func (s *Store) UpdateUserNick(ctx context.Context, userID int, nick string) error {
	const query = "UPDATE `helloapi_users` SET `nick` = ? WHERE `id` = ?"
	return execute(ctx, s.query, query, nick, userID)
}

func (s *Store) UpdateUserMail(ctx context.Context, userID int, mail string) error {
	const query = "UPDATE `helloapi_users` SET `mail` = ? WHERE `id` = ?"
	return execute(ctx, s.query, query, mail, userID)
}

func (s *Store) UpdateUserPassword(ctx context.Context, userID int, oldPassword string, newPassword string) (int64, error) {
	const query = "UPDATE `helloapi_users` SET `password` = ? WHERE `id` = ? AND `password`=?"
	return affected(ctx, s.query, query, newPassword, userID, oldPassword)
}

func (s *Store) SetUserPassword(ctx context.Context, userID int, password string) error {
	const query = "UPDATE `helloapi_users` SET `password` = ? WHERE `id` = ? "
	return execute(ctx, s.query, query, password, userID)
}

func (s *Store) GetUserList(ctx context.Context, pageSize int, offset int) ([]model.User, error) {
	const query = "SELECT id, username, nick, mode, mail, created, updated FROM `helloapi_users` ORDER BY mode DESC, created" +
		" LIMIT ? OFFSET ? "
	return selectRows[model.User](ctx, s.query, query, pageSize, offset)
}

func (s *Store) GetUserListSearch(ctx context.Context, keyword string, mode int, pageSize int, offset int) ([]model.User, error) {
	const query = "SELECT id, username, nick, mode, mail, created, updated FROM `helloapi_users` WHERE (? IS NULL OR ? = ''" +
		" OR `username` LIKE CONCAT('%', ?, '%') OR `nick` LIKE CONCAT('%', ?, '%') OR `id` = ?) AND (? = -1 OR" +
		" `mode` = ?) ORDER BY `mode` DESC, `created`  LIMIT ? OFFSET ?"
	return selectRows[model.User](ctx, s.query, query, keyword, keyword, keyword, keyword, keyword, mode, mode, pageSize, offset)
}

func (s *Store) GetUserListSearchCount(ctx context.Context, keyword string, mode int) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_users` WHERE (? IS NULL OR ? = '' OR `username` LIKE CONCAT('%', ?, '%') OR" +
		" `nick` LIKE CONCAT('%', ?, '%') OR `id` = ?) AND (? = -1 OR `mode` = ?)"
	return scalar[int64](ctx, s.query, query, keyword, keyword, keyword, keyword, keyword, mode, mode)
}

func (s *Store) SetUserMode(ctx context.Context, userID int, mode int) error {
	const query = "UPDATE `helloapi_users` SET `mode` = ? WHERE `id` = ?"
	return execute(ctx, s.query, query, mode, userID)
}

func (s *Store) CheckUserIsAdmin(ctx context.Context, userID int) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_users` WHERE `id` = ? AND `mode` = 1"
	return scalar[int64](ctx, s.query, query, userID)
}

func (s *Store) GetUserIDByUserKey(ctx context.Context, key *string) (int, error) {
	const query = "SELECT `user_id` FROM `helloapi_user_keys` WHERE `key` = ?"
	value, err := unique[int](ctx, s.query, query, key)
	if err != nil {
		return 0, err
	}
	if value == nil {
		return 0, sql.ErrNoRows
	}
	return *value, nil
}

func (s *Store) GetCheckAPIList(ctx context.Context, pageSize int, offset int) ([]model.APIApp, error) {
	const query = "SELECT * FROM `helloapi_api_apps` WHERE `view_status` = 2 ORDER BY created DESC LIMIT ? OFFSET ? "
	return selectRows[model.APIApp](ctx, s.query, query, pageSize, offset)
}

func (s *Store) GetCheckAPIListAllCount(ctx context.Context) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_api_apps` WHERE `view_status` = 2"
	return scalar[int64](ctx, s.query, query)
}

func (s *Store) GetCheckAPIListSearch(ctx context.Context, keyword string, kind int, status int, pageSize int, offset int) ([]model.APIApp, error) {
	const query = "SELECT * FROM `helloapi_api_apps` WHERE `view_status` = 2 AND (? IS NULL OR ? = '' OR `url` LIKE" +
		" CONCAT('%',?,'%') OR `title` LIKE CONCAT('%', ?, '%') OR `id` = ? OR `smallTitle` LIKE CONCAT('%', ?," +
		" '%')) AND (? = -1 OR `type` = ?) AND (? = -1 OR `status` = ?) ORDER BY created DESC LIMIT ? OFFSET ?"
	return selectRows[model.APIApp](ctx, s.query, query, keyword, keyword, keyword, keyword, keyword, keyword, kind, kind, status, status, pageSize, offset)
}

func (s *Store) GetCheckAPIListSearchCount(ctx context.Context, keyword string, kind int, status int) (int64, error) {
	const query = "SELECT COUNT(*) FROM `helloapi_api_apps` WHERE `view_status` = 2 AND (? IS NULL OR ? = '' OR `url` LIKE" +
		" CONCAT('%',?,'%') OR `title` LIKE CONCAT('%', ?, '%') OR `id` = ? OR `smallTitle` LIKE CONCAT('%', ?," +
		" '%')) AND (? = -1 OR `type` = ?) AND (? = -1 OR `status` = ?)"
	return scalar[int64](ctx, s.query, query, keyword, keyword, keyword, keyword, keyword, keyword, kind, kind, status, status)
}

func (s *Store) CheckAppChange(ctx context.Context, apiID int, viewStatus int) error {
	const query = "UPDATE `helloapi_api_apps` SET `view_status` = ? WHERE `id` = ?"
	return execute(ctx, s.query, query, viewStatus, apiID)
}

// DeleteUserData 按 Java 删除用户时的顺序，在同一事务内清理用户及其 API。
func (s *Store) DeleteUserData(ctx context.Context, userID int) error {
	return s.inTx(ctx, func(tx *Store) error {
		if err := tx.DeleteUser(ctx, userID); err != nil {
			return err
		}
		if err := tx.DeleteUserKey(ctx, userID); err != nil {
			return err
		}
		apps, err := tx.GetUserAPIListAll(ctx, userID)
		if err != nil {
			return err
		}
		for _, app := range apps {
			for _, remove := range []func(context.Context, int) error{
				tx.DeleteAPIKeyAll, tx.DeleteAPIParamAll, tx.DeleteAPILog,
				tx.DeleteAPICount, tx.DeleteAPIDailyStats, tx.DeleteAPIApp,
			} {
				if err := remove(ctx, app.ID); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
