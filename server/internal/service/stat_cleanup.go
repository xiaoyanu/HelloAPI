package service

import (
	"context"
	"fmt"
)

type LogCleanupRepository interface {
	FindCheckpointNoLock(ctx context.Context) (*int64, error)
	FindPurgeUpperLogID(ctx context.Context, checkpoint int64, batchSize int) (int64, error)
	DeleteExpiredLogs(ctx context.Context, upper int64, retentionDays int) (int64, error)
}

type LogCleanup struct {
	repo          LogCleanupRepository
	retentionDays int
}

func NewLogCleanup(store LogCleanupRepository, retentionDays int) *LogCleanup {
	return &LogCleanup{repo: store, retentionDays: retentionDays}
}

func (s *LogCleanup) PurgeExpiredBatch(ctx context.Context) (int64, error) {
	const batchSize = 5000
	checkpoint, err := s.repo.FindCheckpointNoLock(ctx)
	if err != nil {
		return 0, fmt.Errorf("读取统计检查点: %w", err)
	}
	if checkpoint == nil {
		return 0, nil
	}
	// 只在已完成聚合的日志中选择清理窗口，使用较旧的检查点也安全。
	upper, err := s.repo.FindPurgeUpperLogID(ctx, *checkpoint, batchSize)
	if err != nil {
		return 0, fmt.Errorf("读取清理窗口上界: %w", err)
	}
	if upper == 0 {
		return 0, nil
	}
	deleted, err := s.repo.DeleteExpiredLogs(ctx, upper, s.retentionDays)
	if err != nil {
		return 0, fmt.Errorf("删除过期调用日志: %w", err)
	}
	return deleted, nil
}
