package repo

import (
	"context"
	"time"

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

type dailyCountRow struct {
	Date  string `db:"date"`
	Count int64  `db:"count"`
}

func (s *Store) GetAPIDailyCountArray(ctx context.Context, days int) ([]model.APIWeekArray, error) {
	// 日期按文本读取，避免驱动把 MySQL 的日历日期解释为 Go 本地时区的时刻。
	date, err := scalar[string](ctx, s.query, "SELECT CAST(CURDATE() AS CHAR)")
	if err != nil {
		return nil, err
	}
	today, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return nil, err
	}
	first := today.AddDate(0, 0, 1-days)
	const query = "SELECT CAST(stat_date AS CHAR) AS date, SUM(call_count) AS count FROM helloapi_api_daily_stats" +
		" WHERE stat_date >= ? AND stat_date <= ? GROUP BY stat_date"
	rows, err := selectRows[dailyCountRow](ctx, s.query, query, first.Format(time.DateOnly), date)
	if err != nil {
		return nil, err
	}
	return fillDailyCounts(today, days, rows), nil
}

func fillDailyCounts(today time.Time, days int, rows []dailyCountRow) []model.APIWeekArray {
	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.Date] = row.Count
	}
	result := make([]model.APIWeekArray, days)
	for i := range result {
		date := today.AddDate(0, 0, i+1-days)
		result[i] = model.APIWeekArray{Date: model.LocalDateTime{Time: date}, Count: counts[date.Format(time.DateOnly)]}
	}
	return result
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
