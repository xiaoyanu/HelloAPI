package ee.zxz.helloapi.service;

import ee.zxz.helloapi.mapper.StatAggregationMapper;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/**
 * 原始 API 日志的增量聚合服务。
 *
 * <p>每日统计、总计数和检查点在同一事务内更新。任一步骤失败时事务回滚，
 * 下次调度会从原检查点重新处理，避免漏算。</p>
 */
@Service
public class StatAggregationService {
    /** 单次事务最多处理的日志条数，限制锁持有时间和事务体积。 */
    private static final int BATCH_SIZE = 5000;

    private final StatAggregationMapper statAggregationMapper;

    /**
     * @param statAggregationMapper 增量统计所需的数据库访问接口
     */
    public StatAggregationService(StatAggregationMapper statAggregationMapper) {
        this.statAggregationMapper = statAggregationMapper;
    }

    /**
     * 聚合检查点之后的一批日志。
     *
     * @return 本批日志 ID 跨度；没有待处理日志时返回 0
     */
    @Transactional
    public int aggregateNextBatch() {
        long checkpoint = statAggregationMapper.findCheckpoint();
        long upperLogId = statAggregationMapper.findBatchUpperLogId(checkpoint, BATCH_SIZE);
        if (upperLogId <= checkpoint) {
            return 0;
        }

        statAggregationMapper.aggregateDaily(checkpoint, upperLogId);
        statAggregationMapper.aggregateTotals(checkpoint, upperLogId);
        statAggregationMapper.advanceCheckpoint(upperLogId);
        return Math.toIntExact(upperLogId - checkpoint);
    }
}
