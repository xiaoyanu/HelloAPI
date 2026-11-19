package service

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/zxz/helloapi/internal/repo"
)

type aggregationFake struct {
	upper      int64
	fail       error
	calls      []string
	saved      int64
	rolledBack bool
}

func (f *aggregationFake) AggregateInTx(ctx context.Context, run func(repo.AggregationTx) error) error {
	err := run(f)
	f.rolledBack = err != nil
	return err
}

func (f *aggregationFake) FindCheckpoint(context.Context) (int64, error) { return 100, nil }
func (f *aggregationFake) FindBatchUpperLogID(_ context.Context, checkpoint int64, size int) (int64, error) {
	if checkpoint != 100 || size != 5000 {
		return 0, errors.New("检查点或批大小错误")
	}
	return f.upper, nil
}
func (f *aggregationFake) AggregateDaily(_ context.Context, from, to int64) error {
	f.calls = append(f.calls, "每日")
	if from != 100 || to != 125 {
		return errors.New("每日聚合范围错误")
	}
	return f.fail
}
func (f *aggregationFake) AggregateTotals(_ context.Context, from, to int64) error {
	f.calls = append(f.calls, "累计")
	if from != 100 || to != 125 {
		return errors.New("累计聚合范围错误")
	}
	return nil
}
func (f *aggregationFake) AdvanceCheckpoint(_ context.Context, value int64) error {
	f.calls = append(f.calls, "检查点")
	f.saved = value
	return nil
}

func TestAggregateReturnsZeroWithoutNewLogs(t *testing.T) {
	fake := &aggregationFake{upper: 100}
	count, err := NewAggregation(fake).AggregateNextBatch(t.Context())
	if err != nil || count != 0 || len(fake.calls) != 0 {
		t.Fatalf("处理量 %d，写入 %v，错误 %v", count, fake.calls, err)
	}
}

func TestAggregateWritesBothTablesBeforeCheckpoint(t *testing.T) {
	fake := &aggregationFake{upper: 125}
	count, err := NewAggregation(fake).AggregateNextBatch(t.Context())
	if err != nil || count != 25 || fake.saved != 125 || !reflect.DeepEqual(fake.calls, []string{"每日", "累计", "检查点"}) {
		t.Fatalf("处理量 %d，写入 %v，检查点 %d，错误 %v", count, fake.calls, fake.saved, err)
	}
}

func TestAggregateLeavesCheckpointOnFailure(t *testing.T) {
	failure := errors.New("模拟数据库失败")
	fake := &aggregationFake{upper: 125, fail: failure}
	_, err := NewAggregation(fake).AggregateNextBatch(t.Context())
	if !errors.Is(err, failure) || !fake.rolledBack || fake.saved != 0 || !reflect.DeepEqual(fake.calls, []string{"每日"}) {
		t.Fatalf("写入 %v，检查点 %d，回滚 %t，错误 %v", fake.calls, fake.saved, fake.rolledBack, err)
	}
}
