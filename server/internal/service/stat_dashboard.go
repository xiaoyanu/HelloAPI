package service

import (
	"context"
	"slices"
	"sync"
	"time"

	"github.com/zxz/helloapi/internal/model"
)

type DashboardRepository interface {
	GetAPIUsageSummary(ctx context.Context) (*model.APIUsageSummary, error)
	GetDashboardEntitySummary(ctx context.Context) (*model.DashboardEntitySummary, error)
	GetAPITodayCountArrayFromDaily(ctx context.Context) ([]model.APITodayArray, error)
	GetAPIDailyCountArray(ctx context.Context, days int) ([]model.APIWeekArray, error)
	GetAPIAllCount(ctx context.Context) (int64, error)
}

const (
	cacheTTL       = 30 * time.Second
	dailyTrendDays = 90
)

type cached[T any] struct {
	mu      sync.Mutex
	value   T
	expires time.Time
}

func (c *cached[T]) get(load func() (T, error)) (T, error) {
	c.mu.Lock()
	if time.Now().Before(c.expires) {
		value := c.value
		c.mu.Unlock()
		return value, nil
	}
	c.mu.Unlock()
	value, err := load()
	if err != nil {
		return value, err
	}
	c.mu.Lock()
	c.value = value
	c.expires = time.Now().Add(cacheTTL)
	c.mu.Unlock()
	return value, nil
}

type Dashboard struct {
	repo          DashboardRepository
	statDashboard cached[model.StatDashboard]
	apiAllCount   cached[int64]
}

func NewDashboard(store DashboardRepository) *Dashboard { return &Dashboard{repo: store} }

func (s *Dashboard) GetDashboard(ctx context.Context) (model.StatDashboard, error) {
	return s.statDashboard.get(func() (model.StatDashboard, error) { return s.loadDashboard(ctx) })
}

func (s *Dashboard) loadDashboard(ctx context.Context) (model.StatDashboard, error) {
	dashboard := model.StatDashboard{
		APIWeekCountArray:  []model.APIWeekArray{},
		APIDailyCountArray: []model.APIWeekArray{},
	}
	usage, err := s.repo.GetAPIUsageSummary(ctx)
	if err != nil {
		return dashboard, err
	}
	if usage == nil {
		usage = new(model.APIUsageSummary)
	}
	entities, err := s.repo.GetDashboardEntitySummary(ctx)
	if err != nil {
		return dashboard, err
	}
	if entities == nil {
		entities = new(model.DashboardEntitySummary)
	}
	top, err := s.repo.GetAPITodayCountArrayFromDaily(ctx)
	if err != nil {
		return dashboard, err
	}
	rankings := append(make([]model.APITodayArray, 0, max(7, len(top))), top...)
	for len(rankings) < 7 {
		rankings = append(rankings, model.APITodayArray{})
	}
	daily, err := s.repo.GetAPIDailyCountArray(ctx, dailyTrendDays)
	if err != nil {
		return dashboard, err
	}
	if daily == nil {
		daily = dashboard.APIDailyCountArray
	}

	return model.StatDashboard{
		UserCount:              model.StatValue{Count: entities.UserCount},
		UserMonthRegisterCount: model.StatValue{Count: entities.UserMonthRegisterCount},
		APIAppCount:            model.StatValue{Count: entities.APIAppCount},
		APIAppMonthCount:       model.StatValue{Count: entities.APIAppMonthCount},
		APIAllCount:            model.StatValue{Count: entities.APIAllCount},
		APITodayCount:          model.StatValue{Count: usage.TodayCount, Change: usage.TodayChange},
		APIWeekCount:           model.StatValue{Count: usage.WeekCount, Change: usage.WeekChange},
		APIMonthCount:          model.StatValue{Count: usage.MonthCount, Change: usage.MonthChange},
		APIWeekCountArray:      slices.Clone(daily[max(0, len(daily)-7):]),
		APIDailyCountArray:     daily,
		APITodayCountArray:     rankings,
	}, nil
}

func (s *Dashboard) GetAPIAllCount(ctx context.Context) (int64, error) {
	return s.apiAllCount.get(func() (int64, error) { return s.repo.GetAPIAllCount(ctx) })
}
