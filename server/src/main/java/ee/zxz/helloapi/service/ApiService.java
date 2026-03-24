package ee.zxz.helloapi.service;

import jakarta.servlet.http.HttpServletRequest;

import java.util.Map;

public interface ApiService {

    // 获取API列表
    Map<String, Object> getApiList(Map<String, String> requestParam, HttpServletRequest request);

    // 创建API
    Map<String, Object> createApi(Map<String, Object> requestBody, HttpServletRequest request);

    // 更新API
    Map<String, Object> updateApi(String apiId, Map<String, Object> requestBody, HttpServletRequest request);

    // 删除API
    Map<String, Object> deleteApi(String apiId, HttpServletRequest request);

    // 搜索API
    Map<String, Object> searchApiList(Map<String, String> requestParam);

    // 获取API应用详情
    Map<String, Object> getApiApp(String apiId ,String authorizationHeader);

    // 创建API密钥
    Map<String, Object> createApiKey(String finalApiID, Map<String, Object> requestBody, HttpServletRequest request);

    // 获取API密钥列表
    Map<String, Object> getUserApiKeyList(String userId, Map<String, String> requestParam, HttpServletRequest request);

    // 更新API密钥
    Map<String, Object> updateApiKey(String finalKey, Map<String, Object> requestBody, HttpServletRequest request);

    // 获取API密钥详情
    Map<String, Object> getApiKey(String finalKey, HttpServletRequest request);

    // 删除API密钥
    Map<String, Object> deleteApiKey(String finalKey, HttpServletRequest request);

    // 重置API密钥
    Map<String, Object> resetApiKey(String finalKey, HttpServletRequest request);

    // 搜索用户API密钥列表
    Map<String, Object> userApiKeyListSearch(String userId, Map<String, String> requestParam, HttpServletRequest request);
}
