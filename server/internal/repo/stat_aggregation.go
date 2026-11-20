package repo

import (
	"context"
)

func (s *Store) FindCheckpoint(ctx context.Context) (int64, error) {
	const query = "SELECT `last_log_id` FROM `helloapi_stat_checkpoint` WHERE `job_name` = 'api-daily-stats' FOR UPDATE"
	return scalar[int64](ctx, s.query, query)
}

func (s *Store) FindBatchUpperLogID(ctx context.Context, lastLogID int64, batchSize int) (int64, error) {
	const query = "SELECT COALESCE(MAX(`log_id`), ?) FROM (SELECT `log_id` FROM `helloapi_api_request_logs` WHERE `log_id` >" +
		" ? ORDER BY `log_id` LIMIT ?) AS batch"
	return scalar[int64](ctx, s.query, query, lastLogID, lastLogID, batchSize)
}

func (s *Store) AggregateDaily(ctx context.Context, fromLogID int64, toLogID int64) error {
	const query = "INSERT INTO `helloapi_api_daily_stats` (`stat_date`, `api_id`, `call_count`) SELECT DATE(`time`)," +
		" `api_id`, COUNT(*) FROM `helloapi_api_request_logs` WHERE `log_id` > ? AND `log_id` <= ? GROUP BY" +
		" DATE(`time`), `api_id` ON DUPLICATE KEY UPDATE `call_count` = `call_count` + VALUES(`call_count`)"
	return execute(ctx, s.query, query, fromLogID, toLogID)
}

func (s *Store) AggregateTotals(ctx context.Context, fromLogID int64, toLogID int64) error {
	const query = "INSERT INTO `helloapi_api_views` (`api_id`, `count`) SELECT `api_id`, COUNT(*) FROM" +
		" `helloapi_api_request_logs` WHERE `log_id` > ? AND `log_id` <= ? GROUP BY `api_id` ON DUPLICATE KEY" +
		" UPDATE `count` = `count` + VALUES(`count`)"
	return execute(ctx, s.query, query, fromLogID, toLogID)
}

func (s *Store) AdvanceCheckpoint(ctx context.Context, lastLogID int64) error {
	const query = "UPDATE `helloapi_stat_checkpoint` SET `last_log_id` = ? WHERE `job_name` = 'api-daily-stats'"
	return execute(ctx, s.query, query, lastLogID)
}

type AggregationTx interface {
	FindCheckpoint(context.Context) (int64, error)
	FindBatchUpperLogID(context.Context, int64, int) (int64, error)
	AggregateDaily(context.Context, int64, int64) error
	AggregateTotals(context.Context, int64, int64) error
	AdvanceCheckpoint(context.Context, int64) error
}

func (s *Store) AggregateInTx(ctx context.Context, fn func(AggregationTx) error) error {
	return s.inTx(ctx, func(tx *Store) error { return fn(tx) })
}
