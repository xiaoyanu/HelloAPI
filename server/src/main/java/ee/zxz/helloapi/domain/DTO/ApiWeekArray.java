package ee.zxz.helloapi.domain.DTO;

import java.time.LocalDateTime;

public class ApiWeekArray {
    private LocalDateTime date;
    private Long count;

    public LocalDateTime getDate() { return date; }
    public void setDate(LocalDateTime date) { this.date = date; }
    public Long getCount() { return count; }
    public void setCount(Long count) { this.count = count; }
}
