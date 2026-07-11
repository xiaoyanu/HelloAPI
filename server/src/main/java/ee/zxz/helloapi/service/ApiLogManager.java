package ee.zxz.helloapi.service;

import ee.zxz.helloapi.mapper.StatMapper;
import jakarta.annotation.Resource;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

@Component
public class ApiLogManager {

    @Resource
    private StatMapper statMapper;

    /**
     * 可靠写入原始日志；统计计数由增量聚合任务批量更新。
     *
     * @param appId  被调用的 API ID
     * @param ip     调用方 IP 地址
     * @param header 请求头 JSON；无效或未提供时为 null
     * @param body   请求体 JSON；无效或未提供时为 null
     * @param apiKey 本次调用使用的 API Key；未使用时为 null
     * @param userId 发起调用的用户 ID
     */
    @Transactional
    public void saveLog(int appId, String ip, String header, String body, String apiKey, int userId) {
        statMapper.insertApiLog(appId, ip, header, body, apiKey, userId);
    }
}
