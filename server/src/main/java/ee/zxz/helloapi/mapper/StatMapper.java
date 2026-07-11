package ee.zxz.helloapi.mapper;

import ee.zxz.helloapi.domain.DTO.ApiTodayArray;
import ee.zxz.helloapi.domain.DTO.ApiUsageSummary;
import ee.zxz.helloapi.domain.DTO.ApiWeekArray;
import ee.zxz.helloapi.domain.DTO.DashboardEntitySummary;
import org.apache.ibatis.annotations.Delete;
import org.apache.ibatis.annotations.Insert;
import org.apache.ibatis.annotations.Mapper;
import org.apache.ibatis.annotations.Select;

import java.util.List;

@Mapper
public interface StatMapper {
    @Delete("delete from helloapi_api_views where api_id = #{apiId}")
    void deleteApiCount(int apiId);

    @Delete("delete from helloapi_api_request_logs where api_id = #{apiId}")
    void deleteApiLog(int apiId);

    @Delete("delete from helloapi_api_daily_stats where api_id = #{apiId}")
    void deleteApiDailyStats(int apiId);

    /**
     * @param apiId  API ID
     * @param ip     调用方 IP
     * @param header 请求头 JSON
     * @param body   请求体 JSON
     * @param apiKey 调用使用的 API Key
     * @param userId 调用方用户 ID
     */
    @Insert("insert into helloapi_api_request_logs (`api_id`,`ip`, `header`, `body`,`api_key`,`user_id`) values (#{apiId}, #{ip}, #{header}, #{body},#{apiKey},#{userId})")
    void insertApiLog(int apiId, String ip, Object header, String body, String apiKey, int userId);

    @Select("SELECT COALESCE(SUM(`count`), 0) FROM `helloapi_api_views`")
    long getApiAllCount();

    /** 从每日聚合表计算当前时间窗口及上一等长窗口的调用量。 */
    @Select("SELECT " +
            "COALESCE(SUM(CASE WHEN `stat_date` = CURDATE() THEN `call_count` ELSE 0 END), 0) AS todayCount, " +
            "COALESCE(SUM(CASE WHEN `stat_date` = CURDATE() THEN `call_count` WHEN `stat_date` = CURDATE() - INTERVAL 1 DAY THEN -`call_count` ELSE 0 END), 0) AS todayChange, " +
            "COALESCE(SUM(CASE WHEN `stat_date` >= CURDATE() - INTERVAL 6 DAY THEN `call_count` ELSE 0 END), 0) AS weekCount, " +
            "COALESCE(SUM(CASE WHEN `stat_date` >= CURDATE() - INTERVAL 6 DAY THEN `call_count` WHEN `stat_date` >= CURDATE() - INTERVAL 13 DAY THEN -`call_count` ELSE 0 END), 0) AS weekChange, " +
            "COALESCE(SUM(CASE WHEN `stat_date` >= CURDATE() - INTERVAL 29 DAY THEN `call_count` ELSE 0 END), 0) AS monthCount, " +
            "COALESCE(SUM(CASE WHEN `stat_date` >= CURDATE() - INTERVAL 29 DAY THEN `call_count` WHEN `stat_date` >= CURDATE() - INTERVAL 59 DAY THEN -`call_count` ELSE 0 END), 0) AS monthChange " +
            "FROM `helloapi_api_daily_stats` WHERE `stat_date` >= CURDATE() - INTERVAL 59 DAY")
    ApiUsageSummary getApiUsageSummary();

    /** 查询最近 7 天逐日调用量，缺少数据的日期返回 0。 */
    @Select("SELECT CAST(Days.Date AS DATETIME) AS date, COALESCE(SUM(s.call_count), 0) AS count " +
            "FROM (SELECT CURDATE() - INTERVAL 6 DAY AS Date UNION ALL " +
            "SELECT CURDATE() - INTERVAL 5 DAY UNION ALL SELECT CURDATE() - INTERVAL 4 DAY UNION ALL " +
            "SELECT CURDATE() - INTERVAL 3 DAY UNION ALL SELECT CURDATE() - INTERVAL 2 DAY UNION ALL " +
            "SELECT CURDATE() - INTERVAL 1 DAY UNION ALL SELECT CURDATE()) AS Days " +
            "LEFT JOIN `helloapi_api_daily_stats` s ON s.stat_date = Days.Date " +
            "GROUP BY Days.Date ORDER BY Days.Date")
    List<ApiWeekArray> getApiWeekCountArrayFromDaily();

    /** 查询今日调用量最高的 7 个 API。 */
    @Select("SELECT a.title AS name, COALESCE(s.call_count, 0) AS count " +
            "FROM `helloapi_api_apps` a LEFT JOIN `helloapi_api_daily_stats` s " +
            "ON s.api_id = a.id AND s.stat_date = CURDATE() " +
            "ORDER BY count DESC LIMIT 7")
    List<ApiTodayArray> getApiTodayCountArrayFromDaily();

    /** 一次查询 Dashboard 所需的用户数、API 数和累计调用次数。 */
    @Select("SELECT " +
            "(SELECT COUNT(*) FROM `helloapi_users`) AS userCount, " +
            "(SELECT COUNT(*) FROM `helloapi_users` WHERE `created` >= CURDATE() - INTERVAL 29 DAY AND `created` < CURDATE() + INTERVAL 1 DAY) AS userMonthRegisterCount, " +
            "(SELECT COUNT(*) FROM `helloapi_api_apps`) AS apiAppCount, " +
            "(SELECT COUNT(*) FROM `helloapi_api_apps` WHERE `created` >= CURDATE() - INTERVAL 29 DAY AND `created` < CURDATE() + INTERVAL 1 DAY) AS apiAppMonthCount, " +
            "(SELECT COALESCE(SUM(`count`), 0) FROM `helloapi_api_views`) AS apiAllCount")
    DashboardEntitySummary getDashboardEntitySummary();
}
