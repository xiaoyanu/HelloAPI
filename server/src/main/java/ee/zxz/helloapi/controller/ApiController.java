package ee.zxz.helloapi.controller;

import ee.zxz.helloapi.annotation.RequiresLogin;
import ee.zxz.helloapi.service.ApiService;
import ee.zxz.helloapi.utils.Finals;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping(Finals.apiUrl)
public class ApiController {
    private final ApiService apiService;

    public ApiController(ApiService apiService) {
        this.apiService = apiService;
    }

    // GetApiList - 获取API列表 - GET
    @GetMapping("/list")
    public Map<String, Object> GetApiList(@RequestParam(required = false) Map<String, String> requestParam, HttpServletRequest request) {
        return apiService.getApiList(requestParam, request);
    }

    // CreateApi - 创建API - POST
    @PostMapping("/")
    @RequiresLogin
    public Map<String, Object> CreateApi(@RequestBody Map<String, Object> requestBody, HttpServletRequest request) {
        return apiService.createApi(requestBody, request);
    }

    // UpdateApi - 更新API - PUT
    @PutMapping({"/{apiId}", "/"})
    @RequiresLogin
    public Map<String, Object> UpdateApi(@PathVariable(required = false) String apiId, @RequestBody Map<String, Object> requestBody, HttpServletRequest request) {
        String finalApiID = (apiId == null) ? "0" : apiId;
        return apiService.updateApi(finalApiID, requestBody, request);
    }

    // DeleteApi - 删除API - DELETE
    @DeleteMapping({"/{apiId}", "/"})
    @RequiresLogin
    public Map<String, Object> DeleteApi(@PathVariable(required = false) String apiId, HttpServletRequest request) {
        String finalApiID = (apiId == null) ? "0" : apiId;
        return apiService.deleteApi(finalApiID, request);
    }

    // SearchApiList - 搜索Api接口 - GET
    @GetMapping("/search")
    public Map<String, Object> SearchApiList(@RequestParam Map<String, String> requestParam) {
        return apiService.searchApiList(requestParam);
    }

    // GetApiApp - 获取Api接口详情 - GET
    @GetMapping({"/{apiId}", "/"})
    public Map<String, Object> GetApiApp(@PathVariable(required = false) String apiId,@RequestHeader(value = "Authorization", required = false) String authorizationHeader) {
        String finalApiID = (apiId == null) ? "0" : apiId;
        return apiService.getApiApp(finalApiID,authorizationHeader);
    }

    // CreateApiKey - 创建API密钥 - POST
    @PostMapping({"/key/{api_id}", "/key/"})
    @RequiresLogin
    public Map<String, Object> CreateApiKey(@PathVariable(required = false) String api_id, @RequestBody Map<String, Object> requestBody, HttpServletRequest request) {
        String finalApiID = (api_id == null) ? "0" : api_id;
        return apiService.createApiKey(finalApiID, requestBody, request);
    }

    // GetUserApiKeyList - 获取用户API密钥列表 - GET
    @GetMapping({"/key/list/", "/key/list/{user_id}"})
    @RequiresLogin
    public Map<String, Object> GetUserApiKeyList(@PathVariable(required = false) String user_id, @RequestParam(required = false) Map<String, String> requestParam, HttpServletRequest request) {
        String finalUserId = (user_id == null) ? "0" : user_id;
        return apiService.getUserApiKeyList(finalUserId, requestParam, request);
    }

    // UserApiKeyListSearch - 搜索用户API密钥列表 - GET
    @GetMapping({"/key/list/", "/key/list/{user_id}/Search"})
    @RequiresLogin
    public Map<String, Object> UserApiKeyListSearch(@PathVariable(required = false) String user_id, @RequestParam(required = false) Map<String, String> requestParam, HttpServletRequest request) {
        String finalUserId = (user_id == null) ? "0" : user_id;
        return apiService.userApiKeyListSearch(finalUserId, requestParam, request);
    }

    // UpdateApiKey - 更新API密钥 - PUT
    @PutMapping({"/key/{key}", "/key/"})
    @RequiresLogin
    public Map<String, Object> UpdateApiKey(@PathVariable(required = false) String key, @RequestBody Map<String, Object> requestBody, HttpServletRequest request) {
        String finalKey = (key == null) ? "0" : key;
        return apiService.updateApiKey(finalKey, requestBody, request);
    }

    // GetApiKey - 获取API密钥 - GET
    @GetMapping({"/key/{key}", "/key/"})
    @RequiresLogin
    public Map<String, Object> GetApiKey(@PathVariable(required = false) String key, HttpServletRequest request) {
        String finalKey = (key == null) ? "0" : key;
        return apiService.getApiKey(finalKey, request);
    }

    // DeleteApiKey - 删除API密钥 - DELETE
    @DeleteMapping({"/key/{key}", "/key/"})
    @RequiresLogin
    public Map<String, Object> DeleteApiKey(@PathVariable(required = false) String key, HttpServletRequest request) {
        String finalKey = (key == null) ? "0" : key;
        return apiService.deleteApiKey(finalKey, request);
    }

    // ResetApiKey - 重置API密钥 - PUT
    @PutMapping({"/key/reset/{key}", "/key/reset/"})
    @RequiresLogin
    public Map<String, Object> ResetApiKey(@PathVariable(required = false) String key, HttpServletRequest request) {
        String finalKey = (key == null) ? "0" : key;
        return apiService.resetApiKey(finalKey, request);
    }
}