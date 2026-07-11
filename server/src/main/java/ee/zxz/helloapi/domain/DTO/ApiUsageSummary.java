package ee.zxz.helloapi.domain.DTO;

/**
 * API 调用量的时间窗口汇总结果。
 * change 字段表示当前窗口调用量减去上一个等长窗口调用量。
 */
public class ApiUsageSummary {
    private long todayCount;
    private long todayChange;
    private long weekCount;
    private long weekChange;
    private long monthCount;
    private long monthChange;

    public long getTodayCount() { return todayCount; }
    public void setTodayCount(long todayCount) { this.todayCount = todayCount; }
    public long getTodayChange() { return todayChange; }
    public void setTodayChange(long todayChange) { this.todayChange = todayChange; }
    public long getWeekCount() { return weekCount; }
    public void setWeekCount(long weekCount) { this.weekCount = weekCount; }
    public long getWeekChange() { return weekChange; }
    public void setWeekChange(long weekChange) { this.weekChange = weekChange; }
    public long getMonthCount() { return monthCount; }
    public void setMonthCount(long monthCount) { this.monthCount = monthCount; }
    public long getMonthChange() { return monthChange; }
    public void setMonthChange(long monthChange) { this.monthChange = monthChange; }
}
