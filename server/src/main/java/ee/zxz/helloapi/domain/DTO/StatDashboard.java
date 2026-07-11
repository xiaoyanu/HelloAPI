package ee.zxz.helloapi.domain.DTO;

import java.util.List;

/**
 * 管理后台统计页的一次性响应数据。
 * 字段名称与前端 {@code StatDashboard} 接口保持一致。
 */
public class StatDashboard {
    private StatValue userCount;
    private StatValue userMonthRegisterCount;
    private StatValue apiAppCount;
    private StatValue apiAppMonthCount;
    private StatValue apiAllCount;
    private StatValue apiTodayCount;
    private StatValue apiWeekCount;
    private StatValue apiMonthCount;
    private List<ApiWeekArray> apiWeekCountArray;
    private List<ApiTodayArray> apiTodayCountArray;

    public StatValue getUserCount() { return userCount; }
    public void setUserCount(StatValue userCount) { this.userCount = userCount; }
    public StatValue getUserMonthRegisterCount() { return userMonthRegisterCount; }
    public void setUserMonthRegisterCount(StatValue value) { this.userMonthRegisterCount = value; }
    public StatValue getApiAppCount() { return apiAppCount; }
    public void setApiAppCount(StatValue apiAppCount) { this.apiAppCount = apiAppCount; }
    public StatValue getApiAppMonthCount() { return apiAppMonthCount; }
    public void setApiAppMonthCount(StatValue value) { this.apiAppMonthCount = value; }
    public StatValue getApiAllCount() { return apiAllCount; }
    public void setApiAllCount(StatValue apiAllCount) { this.apiAllCount = apiAllCount; }
    public StatValue getApiTodayCount() { return apiTodayCount; }
    public void setApiTodayCount(StatValue apiTodayCount) { this.apiTodayCount = apiTodayCount; }
    public StatValue getApiWeekCount() { return apiWeekCount; }
    public void setApiWeekCount(StatValue apiWeekCount) { this.apiWeekCount = apiWeekCount; }
    public StatValue getApiMonthCount() { return apiMonthCount; }
    public void setApiMonthCount(StatValue apiMonthCount) { this.apiMonthCount = apiMonthCount; }
    public List<ApiWeekArray> getApiWeekCountArray() { return apiWeekCountArray; }
    public void setApiWeekCountArray(List<ApiWeekArray> value) { this.apiWeekCountArray = value; }
    public List<ApiTodayArray> getApiTodayCountArray() { return apiTodayCountArray; }
    public void setApiTodayCountArray(List<ApiTodayArray> value) { this.apiTodayCountArray = value; }
}
