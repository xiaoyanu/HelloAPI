package repo

import (
	"context"
	"github.com/zxz/helloapi/internal/model"
)

func (s *Store) DeleteAPICount(ctx context.Context, apiID int) error {
	const query = "delete from helloapi_api_views where api_id = ?"
	return execute(ctx, s.query, query, apiID)
}

func (s *Store) DeleteAPILog(ctx context.Context, apiID int) error {
	const query = "delete from helloapi_api_request_logs where api_id = ?"
	return execute(ctx, s.query, query, apiID)
}

func (s *Store) DeleteAPIDailyStats(ctx context.Context, apiID int) error {
	const query = "delete from helloapi_api_daily_stats where api_id = ?"
	return execute(ctx, s.query, query, apiID)
}

func (s *Store) InsertAPILog(ctx context.Context, apiID int, ip *string, header *string, body *string, apiKey *string, userID int) error {
	const query = "insert into helloapi_api_request_logs (`api_id`,`ip`, `header`, `body`,`api_key`,`user_id`) values (?, ?, ?, ?,?,?)"
	return execute(ctx, s.query, query, apiID, ip, header, body, apiKey, userID)
}

func (s *Store) GetAPIAllCount(ctx context.Context) (int64, error) {
	const query = "SELECT COALESCE(SUM(`count`), 0) FROM `helloapi_api_views`"
	return scalar[int64](ctx, s.query, query)
}

func (s *Store) GetAPIUsageSummary(ctx context.Context) (*model.APIUsageSummary, error) {
	const query = "SELECT COALESCE(SUM(CASE WHEN `stat_date` = CURDATE() THEN `call_count` ELSE 0 END), 0) AS todayCount," +
		" COALESCE(SUM(CASE WHEN `stat_date` = CURDATE() THEN `call_count` WHEN `stat_date` = CURDATE() - INTERVAL" +
		" 1 DAY THEN -`call_count` ELSE 0 END), 0) AS todayChange, COALESCE(SUM(CASE WHEN `stat_date` >= CURDATE()" +
		" - INTERVAL 6 DAY THEN `call_count` ELSE 0 END), 0) AS weekCount, COALESCE(SUM(CASE WHEN `stat_date` >=" +
		" CURDATE() - INTERVAL 6 DAY THEN `call_count` WHEN `stat_date` >= CURDATE() - INTERVAL 13 DAY THEN" +
		" -`call_count` ELSE 0 END), 0) AS weekChange, COALESCE(SUM(CASE WHEN `stat_date` >= CURDATE() - INTERVAL" +
		" 29 DAY THEN `call_count` ELSE 0 END), 0) AS monthCount, COALESCE(SUM(CASE WHEN `stat_date` >= CURDATE() -" +
		" INTERVAL 29 DAY THEN `call_count` WHEN `stat_date` >= CURDATE() - INTERVAL 59 DAY THEN -`call_count` ELSE" +
		" 0 END), 0) AS monthChange FROM `helloapi_api_daily_stats` WHERE `stat_date` >= CURDATE() - INTERVAL 59" +
		" DAY"
	return getRow[model.APIUsageSummary](ctx, s.query, query)
}

func (s *Store) GetAPIWeekCountArrayFromDaily(ctx context.Context) ([]model.APIWeekArray, error) {
	const query = "SELECT CAST(Days.Date AS DATETIME) AS date, COALESCE(SUM(s.call_count), 0) AS count FROM (SELECT CURDATE()" +
		" - INTERVAL 6 DAY AS Date UNION ALL SELECT CURDATE() - INTERVAL 5 DAY UNION ALL SELECT CURDATE() -" +
		" INTERVAL 4 DAY UNION ALL SELECT CURDATE() - INTERVAL 3 DAY UNION ALL SELECT CURDATE() - INTERVAL 2 DAY" +
		" UNION ALL SELECT CURDATE() - INTERVAL 1 DAY UNION ALL SELECT CURDATE()) AS Days LEFT JOIN" +
		" `helloapi_api_daily_stats` s ON s.stat_date = Days.Date GROUP BY Days.Date ORDER BY Days.Date"
	return selectRows[model.APIWeekArray](ctx, s.query, query)
}

func (s *Store) GetAPITodayCountArrayFromDaily(ctx context.Context) ([]model.APITodayArray, error) {
	const query = "SELECT a.title AS name, COALESCE(s.call_count, 0) AS count FROM `helloapi_api_apps` a LEFT JOIN" +
		" `helloapi_api_daily_stats` s ON s.api_id = a.id AND s.stat_date = CURDATE() ORDER BY count DESC LIMIT 7"
	return selectRows[model.APITodayArray](ctx, s.query, query)
}

func (s *Store) GetDashboardEntitySummary(ctx context.Context) (*model.DashboardEntitySummary, error) {
	const query = "SELECT (SELECT COUNT(*) FROM `helloapi_users`) AS userCount, (SELECT COUNT(*) FROM `helloapi_users` WHERE" +
		" `created` >= CURDATE() - INTERVAL 29 DAY AND `created` < CURDATE() + INTERVAL 1 DAY) AS" +
		" userMonthRegisterCount, (SELECT COUNT(*) FROM `helloapi_api_apps`) AS apiAppCount, (SELECT COUNT(*) FROM" +
		" `helloapi_api_apps` WHERE `created` >= CURDATE() - INTERVAL 29 DAY AND `created` < CURDATE() + INTERVAL 1" +
		" DAY) AS apiAppMonthCount, (SELECT COALESCE(SUM(`count`), 0) FROM `helloapi_api_views`) AS apiAllCount"
	return getRow[model.DashboardEntitySummary](ctx, s.query, query)
}

func (s *Store) SaveAPILog(ctx context.Context, apiID int, ip, header, body, apiKey *string, userID int) error {
	return s.inTx(ctx, func(tx *Store) error {
		return tx.InsertAPILog(ctx, apiID, ip, header, body, apiKey, userID)
	})
}
