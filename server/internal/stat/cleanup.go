package stat

import (
	"context"
	"log/slog"
	"time"
)

type LogPurger interface {
	PurgeExpiredBatch(context.Context) (int64, error)
}

func RunCleanup(ctx context.Context, purger LogPurger, initialDelay, interval, pause time.Duration, logger *slog.Logger) {
	timer := time.NewTimer(initialDelay)
	defer timer.Stop()
	wait := func() bool {
		select {
		case <-ctx.Done():
			return false
		case <-timer.C:
			return ctx.Err() == nil
		}
	}

	for wait() {
		start := time.Now()
		var total int64
		for {
			deleted, err := purger.PurgeExpiredBatch(ctx)
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				logger.ErrorContext(ctx, "清理过期调用日志失败", "错误", err)
				break
			}
			if deleted == 0 {
				break
			}
			total += deleted
			timer.Reset(pause)
			if !wait() {
				return
			}
		}
		logger.InfoContext(ctx, "本轮过期调用日志清理结束", "删除行数", total, "耗时", time.Since(start))
		timer.Reset(interval)
	}
}
