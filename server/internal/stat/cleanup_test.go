package stat

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

type logPurgerFunc func(context.Context) (int64, error)

func (f logPurgerFunc) PurgeExpiredBatch(ctx context.Context) (int64, error) { return f(ctx) }

func TestRunCleanupStopsRoundOnZeroOrError(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
	}{
		{"无过期日志", nil},
		{"删除失败", errors.New("删除日志失败")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				start := time.Now()
				var calls []time.Duration
				purger := logPurgerFunc(func(context.Context) (int64, error) {
					calls = append(calls, time.Since(start))
					switch len(calls) {
					case 1, 2:
						return 5, nil
					case 3:
						return 0, tc.err
					default:
						cancel()
						return 0, nil
					}
				})
				var log bytes.Buffer
				logger := slog.New(slog.NewTextHandler(&log, nil))
				RunCleanup(ctx, purger, time.Second, 10*time.Second, 100*time.Millisecond, logger)
				want := []time.Duration{time.Second, 1100 * time.Millisecond, 1200 * time.Millisecond, 11200 * time.Millisecond}
				if !slices.Equal(calls, want) {
					t.Fatalf("批次开始时刻 = %v，期望 %v", calls, want)
				}
				if !strings.Contains(log.String(), "level=INFO") || !strings.Contains(log.String(), "删除行数=10") ||
					strings.Contains(log.String(), "level=ERROR") != (tc.err != nil) {
					t.Fatalf("清理日志 = %s", &log)
				}
			})
		})
	}
}

func TestRunCleanupCancelsDuringWait(t *testing.T) {
	for _, tc := range []struct {
		name         string
		initialDelay time.Duration
		deleted      int64
		wantCalls    int
	}{
		{"初始延迟", time.Hour, 0, 0},
		{"批次暂停", 0, 1, 1},
		{"轮次间隔", 0, 0, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				go func() {
					time.Sleep(time.Second)
					cancel()
				}()
				start := time.Now()
				calls := 0
				purger := logPurgerFunc(func(context.Context) (int64, error) {
					calls++
					return tc.deleted, nil
				})
				RunCleanup(ctx, purger, tc.initialDelay, time.Hour, time.Hour, slog.New(slog.DiscardHandler))
				if elapsed := time.Since(start); elapsed != time.Second || calls != tc.wantCalls {
					t.Fatalf("退出耗时 = %s，调用次数 = %d，期望 1s、%d 次", elapsed, calls, tc.wantCalls)
				}
			})
		})
	}
}
