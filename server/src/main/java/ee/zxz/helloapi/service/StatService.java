package ee.zxz.helloapi.service;

import jakarta.servlet.http.HttpServletRequest;

import java.util.Map;

public interface StatService {
    // 记录API日志/消耗等
    Map<String, Object> logApi(Map<String, String> requestBody, HttpServletRequest request);
}
