package service

import (
	"context"
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/zxz/helloapi/internal/model"
)

type dashboardFake struct {
	countQueries int
	trendDays    int
	daily        []model.APIWeekArray
}

func (*dashboardFake) GetAPIUsageSummary(context.Context) (*model.APIUsageSummary, error) {
	return &model.APIUsageSummary{TodayCount: 120, TodayChange: 20, WeekCount: 700, WeekChange: 100, MonthCount: 3000, MonthChange: 500}, nil
}
func (*dashboardFake) GetDashboardEntitySummary(context.Context) (*model.DashboardEntitySummary, error) {
	return &model.DashboardEntitySummary{UserCount: 15, UserMonthRegisterCount: 3, APIAppCount: 8, APIAppMonthCount: 2, APIAllCount: 9000}, nil
}
func (d *dashboardFake) GetAPIDailyCountArray(_ context.Context, days int) ([]model.APIWeekArray, error) {
	d.trendDays = days
	return d.daily, nil
}
func (*dashboardFake) GetAPITodayCountArrayFromDaily(context.Context) ([]model.APITodayArray, error) {
	return []model.APITodayArray{{Name: "天气接口", Count: 80}, {Name: "短链接口", Count: 40}}, nil
}
func (d *dashboardFake) GetAPIAllCount(context.Context) (int64, error) {
	d.countQueries++
	return 9000, nil
}

func TestDashboardMetricsAndPadding(t *testing.T) {
	dashboard, err := NewDashboard(&dashboardFake{}).GetDashboard(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if dashboard.UserCount.Count != 15 || dashboard.UserMonthRegisterCount.Count != 3 ||
		dashboard.APIAppCount.Count != 8 || dashboard.APIAppMonthCount.Count != 2 ||
		dashboard.APITodayCount != (model.StatValue{Count: 120, Change: 20}) ||
		dashboard.APIWeekCount != (model.StatValue{Count: 700, Change: 100}) ||
		dashboard.APIMonthCount != (model.StatValue{Count: 3000, Change: 500}) || dashboard.APIAllCount.Count != 9000 {
		t.Fatalf("指标不符：%+v", dashboard)
	}
	if len(dashboard.APITodayCountArray) != 7 || dashboard.APITodayCountArray[0].Count != 80 ||
		dashboard.APITodayCountArray[1].Count != 40 || dashboard.APITodayCountArray[6] != (model.APITodayArray{}) {
		t.Fatalf("排行榜不符：%+v", dashboard.APITodayCountArray)
	}
	encoded, err := json.Marshal(dashboard)
	if err != nil || !strings.Contains(string(encoded), `"apiWeekCountArray":[]`) ||
		!strings.Contains(string(encoded), `"apiDailyCountArray":[]`) {
		t.Fatalf("空趋势 = %s，错误 = %v", encoded, err)
	}
}

func TestDashboardWeekIsIndependentTailOfDailyTrend(t *testing.T) {
	daily := make([]model.APIWeekArray, 90)
	for i := range daily {
		daily[i].Count = int64(i)
	}
	fake := &dashboardFake{daily: daily}
	dashboard, err := NewDashboard(fake).GetDashboard(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if fake.trendDays != 90 || !slices.Equal(dashboard.APIDailyCountArray, daily) ||
		!slices.Equal(dashboard.APIWeekCountArray, daily[83:]) {
		t.Fatalf("查询范围 = %d，日趋势长度 = %d，周趋势 = %v",
			fake.trendDays, len(dashboard.APIDailyCountArray), dashboard.APIWeekCountArray)
	}
	dashboard.APIWeekCountArray[0].Count = -1
	if got := dashboard.APIDailyCountArray[83].Count; got != 83 {
		t.Fatalf("修改周趋势后对应日趋势 = %d，期望 83", got)
	}
}

func TestDashboardPublicCount(t *testing.T) {
	fake := &dashboardFake{}
	service := NewDashboard(fake)
	for range 2 {
		count, err := service.GetAPIAllCount(t.Context())
		if err != nil || count != 9000 {
			t.Fatalf("累计 = %d，错误 = %v", count, err)
		}
	}
	if fake.countQueries != 1 {
		t.Fatalf("累计查询 %d 次", fake.countQueries)
	}
}
