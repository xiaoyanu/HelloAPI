package ee.zxz.helloapi.mapper;

import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Param;
import org.apache.ibatis.annotations.Select;
import org.apache.ibatis.annotations.Update;

/**
 * 增量统计聚合的数据访问接口。
 *
 * <p>聚合任务以日志主键 {@code log_id} 作为游标，将尚未处理的原始日志
 * 汇总到每日统计表和总调用次数表。</p>
 */
@Mapper
public interface StatAggregationMapper {

    /**
     * 锁定并读取聚合检查点，防止多个应用实例重复处理同一批日志。
     *
     * @return 上一次成功聚合的最大日志 ID
     */
    @Select("SELECT `last_log_id` FROM `helloapi_stat_checkpoint` " +
            "WHERE `job_name` = 'api-daily-stats' FOR UPDATE")
    long findCheckpoint();

    /**
     * 查询本批次最后一条日志的 ID。
     *
     * @param lastLogId 已处理的最大日志 ID，仅查询大于该值的日志
     * @param batchSize 单批最多读取的日志条数
     * @return 本批次最大日志 ID；没有新日志时返回 {@code lastLogId}
     */
    @Select("SELECT COALESCE(MAX(`log_id`), #{lastLogId}) FROM (" +
            "SELECT `log_id` FROM `helloapi_api_request_logs` " +
            "WHERE `log_id` > #{lastLogId} ORDER BY `log_id` LIMIT #{batchSize}" +
            ") AS batch")
    long findBatchUpperLogId(@Param("lastLogId") long lastLogId, @Param("batchSize") int batchSize);

    /**
     * 将指定日志 ID 区间按日期、API 分组并累加到每日统计表。
     *
     * @param fromLogId 区间下界，不包含该日志 ID
     * @param toLogId   区间上界，包含该日志 ID
     */
    @Insert("INSERT INTO `helloapi_api_daily_stats` (`stat_date`, `api_id`, `call_count`) " +
            "SELECT DATE(`time`), `api_id`, COUNT(*) FROM `helloapi_api_request_logs` " +
            "WHERE `log_id` > #{fromLogId} AND `log_id` <= #{toLogId} " +
            "GROUP BY DATE(`time`), `api_id` " +
            "ON DUPLICATE KEY UPDATE `call_count` = `call_count` + VALUES(`call_count`)")
    void aggregateDaily(@Param("fromLogId") long fromLogId, @Param("toLogId") long toLogId);

    /**
     * 将指定日志 ID 区间按 API 分组并累加到历史总调用次数表。
     *
     * @param fromLogId 区间下界，不包含该日志 ID
     * @param toLogId   区间上界，包含该日志 ID
     */
    @Insert("INSERT INTO `helloapi_api_views` (`api_id`, `count`) " +
            "SELECT `api_id`, COUNT(*) FROM `helloapi_api_request_logs` " +
            "WHERE `log_id` > #{fromLogId} AND `log_id` <= #{toLogId} GROUP BY `api_id` " +
            "ON DUPLICATE KEY UPDATE `count` = `count` + VALUES(`count`)")
    void aggregateTotals(@Param("fromLogId") long fromLogId, @Param("toLogId") long toLogId);

    /**
     * 在当前聚合事务完成后推进检查点。
     *
     * @param lastLogId 已成功写入聚合表的最大日志 ID
     */
    @Update("UPDATE `helloapi_stat_checkpoint` SET `last_log_id` = #{lastLogId} " +
            "WHERE `job_name` = 'api-daily-stats'")
    void advanceCheckpoint(@Param("lastLogId") long lastLogId);
}
