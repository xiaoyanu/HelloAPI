package ee.zxz.helloapi.service;

import ee.zxz.helloapi.domain.DTO.ApiTodayArray;
import ee.zxz.helloapi.domain.DTO.ApiUsageSummary;
import ee.zxz.helloapi.domain.DTO.DashboardEntitySummary;
import ee.zxz.helloapi.domain.DTO.StatDashboard;
import ee.zxz.helloapi.mapper.StatMapper;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.ArrayList;
import java.util.List;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
class StatDashboardServiceTest {

    @Mock
    private StatMapper statMapper;

    @Test
    void combinesDashboardMetricsAndPadsTopApisToSevenItems() {
        ApiUsageSummary usage = new ApiUsageSummary();
        usage.setTodayCount(120);
        usage.setTodayChange(20);
        usage.setWeekCount(700);
        usage.setWeekChange(100);
        usage.setMonthCount(3000);
        usage.setMonthChange(500);

        ApiTodayArray first = new ApiTodayArray();
        first.setName("天气接口");
        first.setCount(80L);
        ApiTodayArray second = new ApiTodayArray();
        second.setName("短链接口");
        second.setCount(40L);

        DashboardEntitySummary entities = new DashboardEntitySummary();
        entities.setUserCount(15);
        entities.setUserMonthRegisterCount(3);
        entities.setApiAppCount(8);
        entities.setApiAppMonthCount(2);
        entities.setApiAllCount(9000);

        when(statMapper.getApiUsageSummary()).thenReturn(usage);
        when(statMapper.getDashboardEntitySummary()).thenReturn(entities);
        when(statMapper.getApiWeekCountArrayFromDaily()).thenReturn(new ArrayList<>());
        when(statMapper.getApiTodayCountArrayFromDaily()).thenReturn(new ArrayList<>(List.of(first, second)));

        StatDashboard dashboard = new StatDashboardService(statMapper).getDashboard();

        assertEquals(15L, dashboard.getUserCount().getCount());
        assertEquals(3L, dashboard.getUserMonthRegisterCount().getCount());
        assertEquals(120L, dashboard.getApiTodayCount().getCount());
        assertEquals(20L, dashboard.getApiTodayCount().getChange());
        assertEquals(700L, dashboard.getApiWeekCount().getCount());
        assertEquals(3000L, dashboard.getApiMonthCount().getCount());
        assertEquals(9000L, dashboard.getApiAllCount().getCount());
        assertEquals(7, dashboard.getApiTodayCountArray().size());
        assertEquals("", dashboard.getApiTodayCountArray().get(6).getName());
        assertEquals(0L, dashboard.getApiTodayCountArray().get(6).getCount());
    }

    @Test
    void returnsPublicApiAllCount() {
        when(statMapper.getApiAllCount()).thenReturn(9000L);

        long count = new StatDashboardService(statMapper).getApiAllCount();

        assertEquals(9000L, count);
    }
}
