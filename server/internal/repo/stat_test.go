package repo

import (
	"testing"
	"time"
)

func TestFillDailyCountsAcrossYear(t *testing.T) {
	today := time.Date(2027, time.January, 3, 0, 0, 0, 0, time.UTC)
	rows := []dailyCountRow{
		{Date: "2027-01-03", Count: 12},
		{Date: "2026-12-31", Count: 8},
		{Date: "2026-10-06", Count: 5},
	}
	got := fillDailyCounts(today, 90, rows)
	if len(got) != 90 {
		t.Fatalf("趋势长度 = %d，期望 90", len(got))
	}
	first, err := time.Parse(time.DateOnly, "2026-10-06")
	if err != nil {
		t.Fatal(err)
	}
	counts := map[string]int64{"2026-10-06": 5, "2026-12-31": 8, "2027-01-03": 12}
	for i, point := range got {
		date := first.AddDate(0, 0, i).Format(time.DateOnly)
		if point.Date.Format("2006-01-02T15:04:05") != date+"T00:00:00" || point.Count != counts[date] {
			t.Errorf("第 %d 项 = %+v，期望日期 %s，调用量 %d", i, point, date, counts[date])
		}
	}
}
