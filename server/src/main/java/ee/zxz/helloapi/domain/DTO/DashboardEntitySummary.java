package ee.zxz.helloapi.domain.DTO;

/**
 * Dashboard 所需的用户、API 和历史调用总量。
 * 这些数据由一条轻量 SQL 中的子查询一次性返回。
 */
public class DashboardEntitySummary {
    private long userCount;
    private long userMonthRegisterCount;
    private long apiAppCount;
    private long apiAppMonthCount;
    private long apiAllCount;

    public long getUserCount() { return userCount; }
    public void setUserCount(long userCount) { this.userCount = userCount; }
    public long getUserMonthRegisterCount() { return userMonthRegisterCount; }
    public void setUserMonthRegisterCount(long value) { this.userMonthRegisterCount = value; }
    public long getApiAppCount() { return apiAppCount; }
    public void setApiAppCount(long apiAppCount) { this.apiAppCount = apiAppCount; }
    public long getApiAppMonthCount() { return apiAppMonthCount; }
    public void setApiAppMonthCount(long value) { this.apiAppMonthCount = value; }
    public long getApiAllCount() { return apiAllCount; }
    public void setApiAllCount(long apiAllCount) { this.apiAllCount = apiAllCount; }
}
