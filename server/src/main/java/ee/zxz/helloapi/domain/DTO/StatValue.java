package ee.zxz.helloapi.domain.DTO;

/**
 * 前端统计卡片的通用数值结构。
 */
public class StatValue {
    private long count;
    private long change;

    /**
     * @param count  当前统计值
     * @param change 与上一时间窗口相比的变化量
     */
    public StatValue(long count, long change) {
        this.count = count;
        this.change = change;
    }

    public long getCount() { return count; }
    public void setCount(long count) { this.count = count; }
    public long getChange() { return change; }
    public void setChange(long change) { this.change = change; }
}
