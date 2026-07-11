export interface StatValue {
    count: number
    // 当前时间窗口调用量减去上一等长时间窗口调用量
    change: number
}

export interface ApiWeekCountItem {
    date: string
    count: number
}

export interface ApiTodayCountItem {
    name: string
    count: number
}

export interface StatDashboard {
    userCount: StatValue
    userMonthRegisterCount: StatValue
    apiAppCount: StatValue
    apiAppMonthCount: StatValue
    apiAllCount: StatValue
    apiTodayCount: StatValue
    apiWeekCount: StatValue
    apiMonthCount: StatValue
    apiWeekCountArray: ApiWeekCountItem[]
    apiTodayCountArray: ApiTodayCountItem[]
}

export interface StatDashboardResponse {
    code: number
    msg: string
    data: StatDashboard
}
