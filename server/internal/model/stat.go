package model

type APIUsageSummary struct {
	TodayCount  int64 `db:"todayCount" json:"todayCount"`
	TodayChange int64 `db:"todayChange" json:"todayChange"`
	WeekCount   int64 `db:"weekCount" json:"weekCount"`
	WeekChange  int64 `db:"weekChange" json:"weekChange"`
	MonthCount  int64 `db:"monthCount" json:"monthCount"`
	MonthChange int64 `db:"monthChange" json:"monthChange"`
}

type DashboardEntitySummary struct {
	UserCount              int64 `db:"userCount" json:"userCount"`
	UserMonthRegisterCount int64 `db:"userMonthRegisterCount" json:"userMonthRegisterCount"`
	APIAppCount            int64 `db:"apiAppCount" json:"apiAppCount"`
	APIAppMonthCount       int64 `db:"apiAppMonthCount" json:"apiAppMonthCount"`
	APIAllCount            int64 `db:"apiAllCount" json:"apiAllCount"`
}

type APITodayArray struct {
	Name  string `db:"name" json:"name"`
	Count int64  `db:"count" json:"count"`
}

type APIWeekArray struct {
	Date  LocalDateTime `db:"date" json:"date"`
	Count int64         `db:"count" json:"count"`
}

type StatValue struct {
	Count  int64 `json:"count"`
	Change int64 `json:"change"`
}

type StatDashboard struct {
	UserCount              StatValue       `json:"userCount"`
	UserMonthRegisterCount StatValue       `json:"userMonthRegisterCount"`
	APIAppCount            StatValue       `json:"apiAppCount"`
	APIAppMonthCount       StatValue       `json:"apiAppMonthCount"`
	APIAllCount            StatValue       `json:"apiAllCount"`
	APITodayCount          StatValue       `json:"apiTodayCount"`
	APIWeekCount           StatValue       `json:"apiWeekCount"`
	APIMonthCount          StatValue       `json:"apiMonthCount"`
	APIWeekCountArray      []APIWeekArray  `json:"apiWeekCountArray"`
	APIDailyCountArray     []APIWeekArray  `json:"apiDailyCountArray"`
	APITodayCountArray     []APITodayArray `json:"apiTodayCountArray"`
}
