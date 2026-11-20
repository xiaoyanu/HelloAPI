package service

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type logCleanupFake struct {
	checkpoint                         *int64
	upper, deleted                     int64
	checkpointErr, upperErr, deleteErr error
	calls                              []string
	gotCheckpoint, gotUpper            int64
	gotBatchSize, gotRetentionDays     int
}

func (f *logCleanupFake) FindCheckpointNoLock(context.Context) (*int64, error) {
	f.calls = append(f.calls, "checkpoint")
	return f.checkpoint, f.checkpointErr
}

func (f *logCleanupFake) FindPurgeUpperLogID(_ context.Context, checkpoint int64, batchSize int) (int64, error) {
	f.calls = append(f.calls, "upper")
	f.gotCheckpoint, f.gotBatchSize = checkpoint, batchSize
	return f.upper, f.upperErr
}

func (f *logCleanupFake) DeleteExpiredLogs(_ context.Context, upper int64, retentionDays int) (int64, error) {
	f.calls = append(f.calls, "delete")
	f.gotUpper, f.gotRetentionDays = upper, retentionDays
	return f.deleted, f.deleteErr
}

func TestPurgeExpiredBatch(t *testing.T) {
	const retentionDays = 37
	checkpoint := int64(25000)
	for _, tc := range []struct {
		name  string
		fake  logCleanupFake
		calls []string
	}{
		{"无检查点", logCleanupFake{}, []string{"checkpoint"}},
		{"无已聚合日志", logCleanupFake{checkpoint: &checkpoint}, []string{"checkpoint", "upper"}},
		{"删除过期日志", logCleanupFake{checkpoint: &checkpoint, upper: 7000, deleted: 4000}, []string{"checkpoint", "upper", "delete"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fake := &tc.fake
			got, err := NewLogCleanup(fake, retentionDays).PurgeExpiredBatch(t.Context())
			if err != nil || got != fake.deleted || !slices.Equal(fake.calls, tc.calls) {
				t.Fatalf("删除数 = %d，错误 = %v，调用 = %v", got, err, fake.calls)
			}
			if len(tc.calls) >= 2 && (fake.gotCheckpoint != checkpoint || fake.gotBatchSize != 5000) {
				t.Fatalf("窗口查询收到检查点 %d，批大小 %d", fake.gotCheckpoint, fake.gotBatchSize)
			}
			if len(tc.calls) == 3 && (fake.gotUpper != 7000 || fake.gotRetentionDays != retentionDays) {
				t.Fatalf("删除收到窗口上界 %d，保留天数 %d", fake.gotUpper, fake.gotRetentionDays)
			}
		})
	}
}

func TestPurgeExpiredBatchPropagatesErrors(t *testing.T) {
	wantErr := errors.New("读写失败")
	checkpoint := int64(2)
	for _, tc := range []struct {
		name  string
		fake  logCleanupFake
		calls []string
	}{
		{"检查点", logCleanupFake{checkpointErr: wantErr}, []string{"checkpoint"}},
		{"窗口", logCleanupFake{checkpoint: &checkpoint, upperErr: wantErr}, []string{"checkpoint", "upper"}},
		{"删除", logCleanupFake{checkpoint: &checkpoint, upper: 1, deleteErr: wantErr}, []string{"checkpoint", "upper", "delete"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewLogCleanup(&tc.fake, 90).PurgeExpiredBatch(t.Context())
			if got != 0 || !errors.Is(err, wantErr) || !slices.Equal(tc.fake.calls, tc.calls) {
				t.Fatalf("删除数 = %d，错误 = %v，调用 = %v", got, err, tc.fake.calls)
			}
		})
	}
}
