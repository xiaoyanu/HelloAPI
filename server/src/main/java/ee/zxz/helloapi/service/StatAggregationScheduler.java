package ee.zxz.helloapi.service;

import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

/**
 * 定时触发增量统计聚合。
 *
 * <p>默认启动 10 秒后首次执行，之后每隔 10 秒执行；可通过
 * {@code helloapi.stat.initial-delay-ms} 和 {@code helloapi.stat.fixed-delay-ms} 调整。</p>
 */
@Component
public class StatAggregationScheduler {
    private final StatAggregationService statAggregationService;

    /**
     * @param statAggregationService 每次调度实际执行的增量聚合服务
     */
    public StatAggregationScheduler(StatAggregationService statAggregationService) {
        this.statAggregationService = statAggregationService;
    }

    /**
     * 持续处理已有积压批次，直到检查点追上当前日志末尾。
     */
    @Scheduled(initialDelayString = "${helloapi.stat.initial-delay-ms:10000}",
            fixedDelayString = "${helloapi.stat.fixed-delay-ms:10000}")
    public void aggregateNewLogs() {
        int processed;
        do {
            processed = statAggregationService.aggregateNextBatch();
        } while (processed > 0);
    }
}
