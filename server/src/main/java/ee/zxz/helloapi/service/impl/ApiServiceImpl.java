package ee.zxz.helloapi.service.impl;

import com.alibaba.fastjson.JSON;
import ee.zxz.helloapi.domain.ApiApp;
import ee.zxz.helloapi.domain.ApiKey;
import ee.zxz.helloapi.domain.ApiParam;
import ee.zxz.helloapi.mapper.ApiMapper;
import ee.zxz.helloapi.mapper.SettingMapper;
import ee.zxz.helloapi.mapper.StatMapper;
import ee.zxz.helloapi.mapper.UserMapper;
import ee.zxz.helloapi.service.ApiService;
import ee.zxz.helloapi.utils.Finals;
import ee.zxz.helloapi.utils.ResponseUtil;
import ee.zxz.helloapi.utils.Tools;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.*;

@Service
public class ApiServiceImpl implements ApiService {
    private final ApiMapper apiMapper;
    private final UserMapper userMapper;
    private final StatMapper statMapper;
    private final SettingMapper settingMapper;

    public ApiServiceImpl(ApiMapper apiMapper, UserMapper userMapper, StatMapper statMapper, SettingMapper settingMapper) {
        this.apiMapper = apiMapper;
        this.userMapper = userMapper;
        this.statMapper = statMapper;
        this.settingMapper = settingMapper;
    }

    /**
     * 格式化API应用列表，适用于首页获取列表、搜索列表
     *
     * @param list   用来存放结果的变量
     * @param apiApp API应用对象
     */
    private void formatApiAppList(List<Map<String, Object>> list, ApiApp apiApp) {
        Map<String, Object> map = new LinkedHashMap<>();
        map.put("id", apiApp.getId());
        map.put("title", apiApp.getTitle());
        map.put("smallTitle", apiApp.getSmallTitle());
        map.put("status", apiApp.getStatus());
        map.put("type", apiApp.getType());
        list.add(map);
    }

    @Override
    public Map<String, Object> getApiList(Map<String, String> requestParam, HttpServletRequest request) {
        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }
        Map<String, Object> response = new LinkedHashMap<>();
        List<Map<String, Object>> list = new ArrayList<>();
        int total = apiMapper.getApiListAllCount();
        for (ApiApp apiApp : apiMapper.getApiList(pageSize, Tools.getPageOffset(page, pageSize))) {
            formatApiAppList(list, apiApp);
        }
        response.put("total", total);
        response.put("list", list);
        return ResponseUtil.response(200, "获取成功", response);
    }

    @Override
    public Map<String, Object> createApi(Map<String, Object> requestBody, HttpServletRequest request) {
        try {
            // 全局设置检测
            int tokenMode = (int) request.getAttribute("userMode");
            String settingValue = settingMapper.getSettingValue("api");
            if (settingValue != null && tokenMode != Finals.Admin && settingValue.equals("false")) {
                return ResponseUtil.response(400, "发布API已关闭");
            }

            ApiApp apiApp = new ApiApp();
            int user_id = (int) request.getAttribute("userId");
            getRequestApiApp(requestBody, apiApp, user_id);
            if (userMapper.checkUserIdExists(user_id) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }
            // 检查审核状态
            if (apiApp.getView_status() != 2) {
                // 检测是不是管理员
                if (tokenMode != Finals.Admin) {
                    return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_ADMIN);
                }
            }
            Object param = requestBody.get("params");
            List<ApiParam> params = JSON.parseArray(JSON.toJSONString(param), ApiParam.class);
            List<ApiParam> uniqueParams = Tools.deleteDuplicatesList(params, ApiParam::getName);

            // 校验参数
            if (!uniqueParams.isEmpty()) {
                for (int i = 0; i < uniqueParams.size(); i++) {
                    ApiParam p = uniqueParams.get(i);
                    if (p.getName() == null || p.getName().trim().isEmpty()) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的参数名不能为空");
                    }
                    if (p.getRequired() != 0 && p.getRequired() != 1) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的必填项不能为空");
                    }
                    if (p.getType() == null || p.getType().trim().isEmpty()) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的类型不能为空");
                    }
                    if (p.getMsg() == null || p.getMsg().trim().isEmpty()) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的描述不能为空");
                    }
                    if (p.getMsg().length() > 245) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的描述不能超过245个字符");
                    }
                }
            }
            // 创建应用
            if (apiMapper.createApiApp(apiApp) > 0) {
                for (ApiParam apiParam : uniqueParams) {
                    apiParam.setApi_id(apiApp.getId());
                    if (apiMapper.createApiParam(apiParam) < 1) {
                        return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
                    }
                }
            }
            return ResponseUtil.success();
        } catch (Exception e) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
        }
    }

    @Override
    public Map<String, Object> updateApi(String apiId, Map<String, Object> requestBody, HttpServletRequest request) {
        try {
            int intApiId = Tools.strToInt(apiId);
            // 检查API是否存在
            if (apiMapper.checkApiExist(intApiId) < 1) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_API_NOT_FOUND);
            }

            // 检查是否是创建人或管理员
            int userId = (int) request.getAttribute("userId");
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }
            int userMode = (int) request.getAttribute("userMode");
            if (apiMapper.checkApiCreator(intApiId, userId) < 1) {
                if (userMode != Finals.Admin) {
                    return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
                }
            }

            ApiApp apiApp = new ApiApp();
            getRequestApiApp(requestBody, apiApp, userId);
            // 校验审核状态
            if (apiApp.getView_status() != 2) {
                // 检测是不是管理员
                if (userMode != Finals.Admin) {
                    return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_ADMIN);
                }
            }
            apiApp.setId(intApiId);
            Object param = requestBody.get("params");
            List<ApiParam> params = JSON.parseArray(JSON.toJSONString(param), ApiParam.class);
            List<ApiParam> uniqueParams = Tools.deleteDuplicatesList(params, ApiParam::getName);
            // 校验参数
            if (!uniqueParams.isEmpty()) {
                for (int i = 0; i < uniqueParams.size(); i++) {
                    ApiParam p = uniqueParams.get(i);
                    if (p.getName() == null || p.getName().trim().isEmpty()) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的参数名不能为空");
                    }
                    if (p.getRequired() != 0 && p.getRequired() != 1) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的必填项不能为空");
                    }
                    if (p.getType() == null || p.getType().trim().isEmpty()) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的类型不能为空");
                    }
                    if (p.getMsg() == null || p.getMsg().trim().isEmpty()) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的描述不能为空");
                    }
                    if (p.getMsg().length() > 245) {
                        return ResponseUtil.response(400, "第 " + (i + 1) + " 行参数的描述不能超过245个字符");
                    }
                }
            }

            // 删除所有参数
            apiMapper.deleteApiParamAll(intApiId);

            // 更新应用
            if (apiMapper.updateApiApp(apiApp) > 0) {
                for (ApiParam apiParam : uniqueParams) {
                    apiParam.setApi_id(apiApp.getId());
                    // 检查参数是否存在
                    if (apiMapper.checkApiParamExistObj(apiParam) > 0) {
                        // 更新参数
                        if (apiMapper.updateApiParam(apiParam) < 1) {
                            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
                        }
                    } else {
                        // 创建参数
                        if (apiMapper.createApiParam(apiParam) < 1) {
                            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
                        }
                    }

                }
            }
            return ResponseUtil.success();
        } catch (Exception e) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);

        }
    }

    @Override
    public Map<String, Object> deleteApi(String apiId, HttpServletRequest request) {
        try {
            int intApiId = Tools.strToInt(apiId);
            // 检查API是否存在
            if (apiMapper.checkApiExist(intApiId) < 1) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_API_NOT_FOUND);
            }

            // 检查是否是创建人或管理员
            int userId = (int) request.getAttribute("userId");
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }
            if (apiMapper.checkApiCreator(intApiId, userId) < 1) {
                int userMode = (int) request.getAttribute("userMode");
                if (userMode != Finals.Admin) {
                    return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
                }
            }


            apiMapper.deleteApiApp(intApiId);
            apiMapper.deleteApiParamAll(intApiId);
            statMapper.deleteApiCount(intApiId);
            statMapper.deleteApiLog(intApiId);
            apiMapper.deleteApiKeyAll(intApiId);
            return ResponseUtil.success();

        } catch (Exception e) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);

        }
    }

    @Override
    public Map<String, Object> searchApiList(Map<String, String> requestParam) {
        String keyword = requestParam.get("keywords");
        if (keyword == null || keyword.isEmpty()) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
        }
        int page = Tools.strToInt(requestParam.get("page"));
        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        if (page < 1) {
            page = 1;
        }
        if (pageSize < 1) {
            pageSize = 30;
        }

        Map<String, Object> response = new LinkedHashMap<>();
        List<Map<String, Object>> list = new ArrayList<>();
        int total = apiMapper.searchApiListCount(keyword);
        for (ApiApp apiApp : apiMapper.searchApiList(keyword, pageSize, Tools.getPageOffset(page, pageSize))) {
            formatApiAppList(list, apiApp);
        }
        response.put("total", total);
        response.put("list", list);
        return ResponseUtil.success(response);
    }

    @Override
    public Map<String, Object> getApiApp(String apiId, String authorizationHeader) {
        int intApiId = Tools.strToInt(apiId);
        ApiApp apiApp = apiMapper.getApiApp(intApiId);
        if (apiApp == null) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_API_NOT_FOUND);
        }
        Map<String, Object> map = new LinkedHashMap<>();

        // 验证接口是否处于非过审 或 隐藏状态，如果是的话验证访问者是否是作者或管理员，否则返回空内容
        if (apiApp.getStatus() == 3 || apiApp.getView_status() != 0) {
            int userId = Tools.tokenToUserId(authorizationHeader);
            int userMode = Tools.tokenToUserMode(authorizationHeader);
            if (userId < 1 || apiApp.getUser_id() != userId) {
                if (userMode != Finals.Admin) {
                    map.put("status", apiApp.getStatus());
                    map.put("view_status", apiApp.getView_status());
                    return ResponseUtil.success(map);
                }
            }
        }

        List<ApiParam> apiParams = apiMapper.getApiParam(intApiId);
        List<Object> paramList = new ArrayList<>();
        map.put("id", apiApp.getId());
        map.put("title", apiApp.getTitle());
        map.put("smallTitle", apiApp.getSmallTitle());
        map.put("status", apiApp.getStatus());
        map.put("type", apiApp.getType());
        map.put("url", apiApp.getUrl());
        map.put("sendType", apiApp.getSendType());
        map.put("returnType", apiApp.getReturnType());
        map.put("returnContent", apiApp.getReturnContent());
        map.put("view_status", apiApp.getView_status());
        for (ApiParam apiParam : apiParams) {
            Map<String, Object> paramMap = new LinkedHashMap<>();
            paramMap.put("name", apiParam.getName());
            paramMap.put("required", apiParam.getRequired());
            paramMap.put("type", apiParam.getType());
            paramMap.put("msg", apiParam.getMsg());
            paramList.add(paramMap);
        }
        map.put("params", paramList);
        map.put("user_id", apiApp.getUser_id());
        return ResponseUtil.success(map);
    }

    @Override
    public Map<String, Object> createApiKey(String finalApiID, Map<String, Object> requestBody, HttpServletRequest request) {
        // 全局设置检测
        int tokenMode = (int) request.getAttribute("userMode");
        String settingValue = settingMapper.getSettingValue("api_key");
        if (settingValue != null && tokenMode != Finals.Admin && settingValue.equals("false")) {
            return ResponseUtil.response(400, "创建APIKey已关闭");
        }


        int intApiId = Tools.strToInt(finalApiID);
        // 检查API是否存在
        if (apiMapper.checkApiExist(intApiId) < 1) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_API_NOT_FOUND);
        }
        // 检查是否是创建人或管理员
        int userId = (int) request.getAttribute("userId");
        if (apiMapper.checkApiCreator(intApiId, userId) < 1) {
            if (tokenMode != Finals.Admin) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
            }
        }

        int type = (int) requestBody.get("type");
        LocalDateTime started = null;
        LocalDateTime expired = null;
        if (requestBody.get("started") != null && !((String) requestBody.get("started")).isEmpty() &&
                requestBody.get("expired") != null && !((String) requestBody.get("expired")).isEmpty()) {
            try {
                started = LocalDateTime.parse((String) requestBody.get("started"));
                expired = LocalDateTime.parse((String) requestBody.get("expired"));
            } catch (Exception e) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
            // 如果过期时间早于开始时间，返回错误
            if (!expired.isAfter(started)) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
        }

        int count = (int) requestBody.get("count");

        if (type == 0) {
            if (started == null) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
        }
        if (type == 1) {
            if (count < 0) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
        }

        // 生成密钥
        String key = Tools.getTextMd5(String.valueOf(intApiId) + System.currentTimeMillis());
        while (apiMapper.checkApiKeyExist(key) > 0) {
            key = Tools.getTextMd5(String.valueOf(intApiId) + System.currentTimeMillis());
        }

        String desc = (String) requestBody.get("desc");
        if (desc != null && !desc.isEmpty()) {
            if (desc.length() > 245) {
                return ResponseUtil.response(400, "备注不能超过245个字符");
            }
        }
        apiMapper.createApiKey(intApiId, key, type, started, expired, count, desc);

        return ResponseUtil.success(key);
    }

    @Override
    public Map<String, Object> getUserApiKeyList(String userId, Map<String, String> requestParam, HttpServletRequest request) {
        int intUserId = Tools.strToInt(userId);
        int tokenUserId = (int) request.getAttribute("userId");
        if (tokenUserId != intUserId) {
            int userMode = (int) request.getAttribute("userMode");
            if (userMode != Finals.Admin) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
            }
        }

        // 检查用户是否存在
        if (userMapper.checkUserIdExists(intUserId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }

        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }
        int total = apiMapper.getApiKeyListAllCount(intUserId);
        List<ApiKey> apiKeyList = apiMapper.getApiKeyList(intUserId, pageSize, Tools.getPageOffset(page, pageSize));
        Map<String, Object> result = new LinkedHashMap<>();
        result.put("total", total);
        result.put("list", apiKeyList);
        return ResponseUtil.success(result);
    }

    @Override
    public Map<String, Object> updateApiKey(String finalKey, Map<String, Object> requestBody, HttpServletRequest request) {
        // 检查密钥是否存在
        if (apiMapper.checkApiKeyExist(finalKey) < 1) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_KEY_NOT_FOUND);
        }

        // 检查是否是创建人或管理员
        int userId = (int) request.getAttribute("userId");
        int intApiId = apiMapper.getApiIdByKey(finalKey);
        if (apiMapper.checkApiCreator(intApiId, userId) < 1) {
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
            }
        }

        int type = (int) requestBody.get("type");
        LocalDateTime started = null;
        LocalDateTime expired = null;
        if (requestBody.get("started") != null && !((String) requestBody.get("started")).isEmpty() &&
                requestBody.get("expired") != null && !((String) requestBody.get("expired")).isEmpty()) {
            try {
                started = LocalDateTime.parse((String) requestBody.get("started"));
                expired = LocalDateTime.parse((String) requestBody.get("expired"));
            } catch (Exception e) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
            // 如果过期时间早于开始时间，返回错误
            if (!expired.isAfter(started)) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
        }

        int count = (int) requestBody.get("count");
        if (type == 0) {
            if (started == null) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
        }
        if (type == 1) {
            if (count < 0) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
        }
        String desc = (String) requestBody.get("desc");
        if (desc != null && !desc.isEmpty()) {
            if (desc.length() > 245) {
                return ResponseUtil.response(400, "备注不能超过245个字符");
            }
        }
        apiMapper.updateApiKey(finalKey, type, started, expired, count, desc);

        return ResponseUtil.success();
    }

    @Override
    public Map<String, Object> getApiKey(String finalKey, HttpServletRequest request) {
        // 检查密钥是否存在
        if (apiMapper.checkApiKeyExist(finalKey) < 1) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_KEY_NOT_FOUND);
        }

        // 检查是否是创建人或管理员
        int userId = (int) request.getAttribute("userId");
        int intApiId = apiMapper.getApiIdByKey(finalKey);
        if (apiMapper.checkApiCreator(intApiId, userId) < 1) {
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
            }
        }

        // 获取密钥详情
        ApiKey apiKey = apiMapper.getApiKey(finalKey);
        return ResponseUtil.success(apiKey);
    }

    @Override
    public Map<String, Object> deleteApiKey(String finalKey, HttpServletRequest request) {
        // 检查密钥是否存在
        if (apiMapper.checkApiKeyExist(finalKey) < 1) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_KEY_NOT_FOUND);
        }

        // 检查是否是创建人或管理员
        int userId = (int) request.getAttribute("userId");
        int intApiId = apiMapper.getApiIdByKey(finalKey);
        if (apiMapper.checkApiCreator(intApiId, userId) < 1) {
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
            }
        }

        // 删除密钥
        apiMapper.deleteApiKey(finalKey);

        return ResponseUtil.success();
    }

    @Override
    public Map<String, Object> resetApiKey(String finalKey, HttpServletRequest request) {
        // 检查密钥是否存在
        if (apiMapper.checkApiKeyExist(finalKey) < 1) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_KEY_NOT_FOUND);
        }

        // 检查是否是创建人或管理员
        int userId = (int) request.getAttribute("userId");
        int intApiId = apiMapper.getApiIdByKey(finalKey);
        if (apiMapper.checkApiCreator(intApiId, userId) < 1) {
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
            }
        }

        // 生成密钥
        String key = Tools.getTextMd5(String.valueOf(intApiId) + System.currentTimeMillis());
        while (apiMapper.checkApiKeyExist(key) > 0) {
            key = Tools.getTextMd5(String.valueOf(intApiId) + System.currentTimeMillis());
        }

        // 重置密钥
        apiMapper.resetApiKey(finalKey, key);

        return ResponseUtil.success(key);
    }

    @Override
    public Map<String, Object> userApiKeyListSearch(String userId, Map<String, String> requestParam, HttpServletRequest request) {
        int intUserId = Tools.strToInt(userId);
        int tokenUserId = Tools.strToInt((String) request.getAttribute("tokenUserId"));
        if (tokenUserId != intUserId) {
            int userMode = (int) request.getAttribute("userMode");
            if (userMode != Finals.Admin) {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_NOT_CREATOR);
            }
        }

        // 检查用户是否存在
        if (userMapper.checkUserIdExists(intUserId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }

        String keyword = requestParam.get("keywords");
        // status，type 他们的-1表示不筛选
        int type = Tools.strToInt(requestParam.get("type"));
        int status = Tools.strToInt(requestParam.get("status")); // 过没过期/有没有次数，0没有，1过期
        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }

        Map<String, Object> resultMap = new LinkedHashMap<>();
        int total = apiMapper.ApiKeyListSearchCount(intUserId, keyword, type, status);
        List<ApiKey> apiKeyList = apiMapper.ApiKeyListSearch(intUserId, keyword, type, status, pageSize, Tools.getPageOffset(page, pageSize));
        resultMap.put("total", total);
        resultMap.put("list", apiKeyList);

        return ResponseUtil.success(resultMap);
    }

    /**
     * 从请求体中获取API应用信息
     *
     * @param requestBody 请求体
     * @param apiApp      ApiApp对象
     */
    private void getRequestApiApp(Map<String, Object> requestBody, ApiApp apiApp, int userId) {
        apiApp.setTitle((String) requestBody.get("title"));
        apiApp.setSmallTitle((String) requestBody.get("smallTitle"));
        apiApp.setType((Integer) requestBody.get("type"));
        apiApp.setStatus((Integer) requestBody.get("status"));
        apiApp.setUrl((String) requestBody.get("url"));
        apiApp.setSendType((Integer) requestBody.get("sendType"));
        apiApp.setReturnType((String) requestBody.get("returnType"));
        apiApp.setReturnContent((String) requestBody.get("returnContent"));
        apiApp.setView_status((Integer) requestBody.get("view_status"));
        apiApp.setUser_id(userId);
    }
}
