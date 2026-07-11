package ee.zxz.helloapi.service;

import ee.zxz.helloapi.domain.DTO.ApiTodayArray;
import ee.zxz.helloapi.domain.DTO.ApiUsageSummary;
import ee.zxz.helloapi.domain.DTO.DashboardEntitySummary;
import ee.zxz.helloapi.domain.DTO.StatDashboard;
import ee.zxz.helloapi.domain.DTO.StatValue;
import ee.zxz.helloapi.mapper.StatMapper;
import org.springframework.cache.annotation.Cacheable;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;

/**
 * 组装管理后台统计数据的只读服务。
 *
 * <p>所有调用量查询都读取聚合表，不扫描原始日志；结果缓存 30 秒。</p>
 */
@Service
public class StatDashboardService {
    private final StatMapper statMapper;

    /**
     * @param statMapper Dashboard 聚合查询的数据访问接口
     */
    public StatDashboardService(StatMapper statMapper) {
        this.statMapper = statMapper;
    }

    /**
     * 获取统计页所需的全部指标、趋势和排行榜。
     *
     * @return 可直接返回给前端的 Dashboard 数据
     */
    @Cacheable("statDashboard")
    public StatDashboard getDashboard() {
        ApiUsageSummary usage = statMapper.getApiUsageSummary();
        if (usage == null) {
            usage = new ApiUsageSummary();
        }
        DashboardEntitySummary entities = statMapper.getDashboardEntitySummary();
        if (entities == null) {
            entities = new DashboardEntitySummary();
        }

        List<ApiTodayArray> topApis = statMapper.getApiTodayCountArrayFromDaily();
        topApis = topApis == null ? new ArrayList<>() : new ArrayList<>(topApis);
        while (topApis.size() < 7) {
            ApiTodayArray emptyItem = new ApiTodayArray();
            emptyItem.setName("");
            emptyItem.setCount(0L);
            topApis.add(emptyItem);
        }

        StatDashboard dashboard = new StatDashboard();
        dashboard.setUserCount(value(entities.getUserCount(), 0));
        dashboard.setUserMonthRegisterCount(value(entities.getUserMonthRegisterCount(), 0));
        dashboard.setApiAppCount(value(entities.getApiAppCount(), 0));
        dashboard.setApiAppMonthCount(value(entities.getApiAppMonthCount(), 0));
        dashboard.setApiAllCount(value(entities.getApiAllCount(), 0));
        dashboard.setApiTodayCount(value(usage.getTodayCount(), usage.getTodayChange()));
        dashboard.setApiWeekCount(value(usage.getWeekCount(), usage.getWeekChange()));
        dashboard.setApiMonthCount(value(usage.getMonthCount(), usage.getMonthChange()));
        dashboard.setApiWeekCountArray(statMapper.getApiWeekCountArrayFromDaily());
        dashboard.setApiTodayCountArray(topApis);
        return dashboard;
    }

    /** 供公开计数图片使用，只查询累计调用次数。 */
    @Cacheable("apiAllCount")
    public long getApiAllCount() {
        return statMapper.getApiAllCount();
    }

    private StatValue value(long count, long change) {
        return new StatValue(count, change);
    }
}
