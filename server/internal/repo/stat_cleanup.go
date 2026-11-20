package repo

import "context"

func (s *Store) FindCheckpointNoLock(ctx context.Context) (*int64, error) {
	const query = "SELECT `last_log_id` FROM `helloapi_stat_checkpoint` WHERE `job_name` = 'api-daily-stats'"
	return getRow[int64](ctx, s.query, query)
}

func (s *Store) FindPurgeUpperLogID(ctx context.Context, checkpoint int64, batchSize int) (int64, error) {
	const query = "SELECT COALESCE(MAX(`log_id`), 0) FROM (SELECT `log_id` FROM `helloapi_api_request_logs`" +
		" WHERE `log_id` <= ? ORDER BY `log_id` LIMIT ?) AS batch"
	return scalar[int64](ctx, s.query, query, checkpoint, batchSize)
}

func (s *Store) DeleteExpiredLogs(ctx context.Context, upper int64, retentionDays int) (int64, error) {
	const query = "DELETE FROM `helloapi_api_request_logs` WHERE `log_id` <= ? AND `time` < CURDATE() - INTERVAL ? DAY"
	return affected(ctx, s.query, query, upper, retentionDays)
}
