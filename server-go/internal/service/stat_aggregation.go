package service

import (
	"context"
	"fmt"
	"math"

	"github.com/zxz/helloapi/internal/repo"
)

type AggregationRepository interface {
	AggregateInTx(context.Context, func(repo.AggregationTx) error) error
}

type Aggregation struct{ repo AggregationRepository }

func NewAggregation(store AggregationRepository) *Aggregation { return &Aggregation{repo: store} }

func (s *Aggregation) AggregateNextBatch(ctx context.Context) (int, error) {
	const batchSize = 5000
	var processed int
	err := s.repo.AggregateInTx(ctx, func(tx repo.AggregationTx) error {
		from, err := tx.FindCheckpoint(ctx)
		if err != nil {
			return fmt.Errorf("锁定统计检查点: %w", err)
		}
		upper, err := tx.FindBatchUpperLogID(ctx, from, batchSize)
		if err != nil {
			return fmt.Errorf("读取批次日志上界: %w", err)
		}
		if upper <= from {
			return nil
		}
		if err := tx.AggregateDaily(ctx, from, upper); err != nil {
			return fmt.Errorf("写入每日调用量: %w", err)
		}
		if err := tx.AggregateTotals(ctx, from, upper); err != nil {
			return fmt.Errorf("写入累计调用量: %w", err)
		}
		if err := tx.AdvanceCheckpoint(ctx, upper); err != nil {
			return fmt.Errorf("更新统计检查点: %w", err)
		}
		if upper-from > math.MaxInt32 {
			return fmt.Errorf("批次日志 ID 跨度超过 32 位整数范围")
		}
		processed = int(upper - from)
		return nil
	})
	if err != nil {
		return 0, err
	}
	return processed, nil
}
