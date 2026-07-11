package ee.zxz.helloapi.controller;

import ee.zxz.helloapi.annotation.RequiresLogin;
import ee.zxz.helloapi.service.StatDashboardService;
import ee.zxz.helloapi.service.StatService;
import ee.zxz.helloapi.utils.Finals;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping(Finals.statUrl)
public class StatController {
    private final StatService statService;
    private final StatDashboardService statDashboardService;

    /**
     * @param statService          日志记录服务
     * @param statDashboardService Dashboard 聚合查询服务
     */
    public StatController(StatService statService, StatDashboardService statDashboardService) {
        this.statService = statService;
        this.statDashboardService = statDashboardService;
    }

    // LogApi - 记录API日志/消耗等 - POST
    @PostMapping({"/log", "/log/"})
    public Map<String, Object> logApi(@RequestBody(required = false) Map<String, String> requestBody, HttpServletRequest request) {
        return statService.logApi(requestBody, request);
    }

    /**
     * 一次获取管理后台统计页所需的全部数据。
     *
     * @return 统一成功响应，data 为缓存后的 Dashboard 聚合结果
     */
    @GetMapping("/dashboard")
    @RequiresLogin
    public Map<String, Object> getDashboard() {
        return ee.zxz.helloapi.utils.ResponseUtil.success(statDashboardService.getDashboard());
    }

    /**
     * 获取系统累计 API 调用次数，供公开计数图片使用。
     */
    @GetMapping("/count")
    public Map<String, Object> getApiAllCount() {
        return ee.zxz.helloapi.utils.ResponseUtil.success(
                Map.of("count", statDashboardService.getApiAllCount())
        );
    }
}
