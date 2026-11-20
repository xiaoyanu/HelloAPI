package stat

import (
	"context"
	"log/slog"
	"time"
)

type Aggregator interface {
	AggregateNextBatch(context.Context) (int, error)
}

func Run(ctx context.Context, aggregator Aggregator, initialDelay, fixedDelay time.Duration, logger *slog.Logger) {
	initial := time.NewTimer(initialDelay)
	defer initial.Stop()
	select {
	case <-ctx.Done():
		return
	case <-initial.C:
	}

	ticker := time.NewTicker(fixedDelay)
	ticker.Stop()
	defer ticker.Stop()
	for ctx.Err() == nil {
		for ctx.Err() == nil {
			count, err := aggregator.AggregateNextBatch(ctx)
			if err != nil && ctx.Err() == nil {
				logger.ErrorContext(ctx, "聚合调用日志失败", "错误", err)
			}
			if err != nil || count <= 0 {
				break
			}
		}

		// 一轮积压处理结束后才开始等待下一次调度。
		ticker.Reset(fixedDelay)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			ticker.Stop()
		}
	}
}
