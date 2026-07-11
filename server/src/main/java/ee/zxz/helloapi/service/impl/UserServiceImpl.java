package ee.zxz.helloapi.service.impl;

import com.pig4cloud.captcha.GifCaptcha;
import com.pig4cloud.captcha.SpecCaptcha;
import com.pig4cloud.captcha.base.Captcha;
import com.pig4cloud.captcha.utils.CaptchaJakartaUtil;
import ee.zxz.helloapi.domain.ApiApp;
import ee.zxz.helloapi.domain.User;
import ee.zxz.helloapi.domain.UserKey;
import ee.zxz.helloapi.mapper.ApiMapper;
import ee.zxz.helloapi.mapper.SettingMapper;
import ee.zxz.helloapi.mapper.StatMapper;
import ee.zxz.helloapi.mapper.UserMapper;
import ee.zxz.helloapi.service.UserService;
import ee.zxz.helloapi.utils.Finals;
import ee.zxz.helloapi.utils.JwtUtil;
import ee.zxz.helloapi.utils.ResponseUtil;
import ee.zxz.helloapi.utils.Tools;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Service;

import java.util.*;

@Service
public class UserServiceImpl implements UserService {
    private final UserMapper userMapper;
    private final ApiMapper apiMapper;
    private final StatMapper statMapper;
    private final SettingMapper settingMapper;

    public UserServiceImpl(UserMapper userMapper, ApiMapper apiMapper, StatMapper statMapper, SettingMapper settingMapper) {
        this.userMapper = userMapper;
        this.apiMapper = apiMapper;
        this.statMapper = statMapper;
        this.settingMapper = settingMapper;
    }

    @Override
    public void getCaptcha(HttpServletRequest request, HttpServletResponse response) throws Exception {
        GifCaptcha gifCaptcha = new GifCaptcha(120, 32, 4);
        gifCaptcha.setCharType(Captcha.TYPE_NUM_AND_UPPER);

        SpecCaptcha captcha = new SpecCaptcha(120, 32, 4);
        captcha.setCharType(Captcha.TYPE_NUM_AND_UPPER);

        if (Math.random() > 0.5) {
            CaptchaJakartaUtil.out(gifCaptcha, request, response);
        } else {
            CaptchaJakartaUtil.out(captcha, request, response);
        }
    }

    @Override
    public Map<String, Object> register(Map<String, String> requestBody, HttpServletRequest request) {
        // 1. 基础非空校验
        if (requestBody == null || requestBody.isEmpty()) {
            return ResponseUtil.response(400, "请求参数不能为空");
        }

        String username = requestBody.get("username");
        String password = requestBody.get("password");
        String code = requestBody.get("code");
        String mail = requestBody.get("mail");


        // 参数校验
        if (username == null || username.trim().isEmpty()) {
            return ResponseUtil.response(400, "账号不能为空");
        }
        if (username.equals(password)) {
            return ResponseUtil.response(400, "账号和密码不能相同");
        }
        if (username.length() < 4) {
            return ResponseUtil.response(400, "账号不能少于4个字符");
        }
        if (password == null || password.trim().isEmpty()) {
            return ResponseUtil.response(400, "密码不能为空");
        }
        if (password.length() < 6) {
            return ResponseUtil.response(400, "密码不能少于6个字符");
        }
        // 检查密码是否包含空格
        if (password.contains(" ")) {
            return ResponseUtil.response(400, "密码不能包含空格");
        }
        // 检查用户名是否包含空格
        if (username.contains(" ")) {
            return ResponseUtil.response(400, "账号不能包含空格");
        }
        if (username.length() > 32) {
            return ResponseUtil.response(400, "账号不能超过32个字符");
        }
        if (password.length() > 64) {
            return ResponseUtil.response(400, "密码不能超过64个字符");
        }
        if (code == null || code.isEmpty()) {
            return ResponseUtil.response(400, "验证码不能为空");
        }
        if (code.length() != 4) {
            return ResponseUtil.response(400, "验证码长度必须为4个字符");
        }
        if (mail == null || mail.trim().isEmpty()) {
            return ResponseUtil.response(400, "邮箱不能为空");
        }
        if (!mail.matches("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z0-9]{2,6}$")) {
            return ResponseUtil.response(400, "邮箱格式错误");
        }

        // 3. 验证码校验
        if (!CaptchaJakartaUtil.ver(code, request)) {
            return ResponseUtil.response(400, "验证码错误/过期，请重新输入");
        }
        CaptchaJakartaUtil.clear(request);

        // 全局设置检测
        String settingValue = settingMapper.getSettingValue("register");
        if (settingValue != null && settingValue.equals("false")) {
            return ResponseUtil.response(400, "注册已关闭");
        }

        try {
            // 4. 数据库校验与写入
            if (userMapper.checkUsernameExists(username) > 0) {
                return ResponseUtil.response(400, "账号已存在");
            }

            String md5Password = Tools.getTextMd5(password);
            int result;
            if (userMapper.getUserCount() < 1) {
                // 第一个注册用户为管理员
                result = userMapper.register(username, md5Password, username, 1, mail);
            } else {
                result = userMapper.register(username, md5Password, username, 0, mail);
            }

            if (result <= 0) {
                return ResponseUtil.response(500, "注册失败，数据库插入失败");
            }

            return ResponseUtil.response(200, "注册成功");
        } catch (Exception e) {
            return ResponseUtil.response(500, "注册失败");
        }
    }

    @Override
    public Map<String, Object> login(Map<String, String> requestBody, HttpServletRequest request) {

        // 获取JSON中的username和password
        String username = requestBody.get("username");
        String password = requestBody.get("password");
        String code = requestBody.get("code");
        // 参数校验
        if (username == null || username.trim().isEmpty()) {
            return ResponseUtil.response(400, "账号不能为空");
        }
        if (password == null || password.trim().isEmpty()) {
            return ResponseUtil.response(400, "密码不能为空");
        }
        if (username.contains(" ")) {
            return ResponseUtil.response(400, "账号不能包含空格");
        }
        if (password.contains(" ")) {
            return ResponseUtil.response(400, "密码不能包含空格");
        }
        if (username.length() > 32) {
            return ResponseUtil.response(400, "账号不能超过32个字符");
        }
        if (username.length() < 4) {
            return ResponseUtil.response(400, "账号不能少于4个字符");
        }
        if (password.length() < 6) {
            return ResponseUtil.response(400, "密码不能少于6个字符");
        }
        if (password.length() > 64) {
            return ResponseUtil.response(400, "密码不能超过64个字符");
        }
        if (code == null || code.isEmpty()) {
            return ResponseUtil.response(400, "验证码不能为空");
        }
        if (code.length() != 4) {
            return ResponseUtil.response(400, "验证码长度必须为4个字符");
        }

        // 验证验证码
        if (!CaptchaJakartaUtil.ver(code, request)) {
            CaptchaJakartaUtil.clear(request);  // 清除session中的验证码
            return ResponseUtil.response(400, "验证码错误/过期，请重新输入");
        }
        CaptchaJakartaUtil.clear(request);  // 清除session中的验证码

        password = Tools.getTextMd5(password);

        if (userMapper.checkUsernameExists(username) <= 0) {
            return ResponseUtil.response(400, "账号不存在");
        }

        // 获取用户信息
        User user = userMapper.login(username, password);
        if (user == null) {
            return ResponseUtil.response(400, "密码错误");
        }

        // 生成JWT token
        String token = JwtUtil.getToken(user.getId(), user.getUsername(), user.getMode());

        // 构建返回
        Map<String, Object> data = new LinkedHashMap<>();
        Map<String, Object> info = new LinkedHashMap<>();
        data.put("token", token);
        data.put("user", info);
        info.put("id", user.getId());
        info.put("username", user.getUsername());
        info.put("nick", user.getNick());
        info.put("mode", user.getMode());
        info.put("mail", user.getMail());
        return ResponseUtil.response(200, "登陆成功", data);
    }

    @Override
    public Map<String, Object> getUserInfo(Map<String, String> requestParam, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }
        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            // 检查Token是否为管理员
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
            }
            userId = Tools.strToInt(requestParam.get("id"));
        }

        // 获取用户信息
        User user = userMapper.getUserInfo(userId);
        if (user == null) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }

        Map<String, Object> data = new LinkedHashMap<>();
        data.put("id", user.getId());
        data.put("username", user.getUsername());
        data.put("nick", user.getNick());
        data.put("mode", user.getMode());
        data.put("mail", user.getMail());

        return ResponseUtil.response(200, "获取成功", data);
    }

    @Override
    public Map<String, Object> getUserKey(Map<String, String> requestParam, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }
        UserKey user_key = userMapper.getUserKey(userId);

        Map<String, Object> data = new LinkedHashMap<>();
        data.put("id", userId);
        if (user_key == null) {
            data.put("key", "");
        } else {
            data.put("key", user_key.getKey());
        }

        return ResponseUtil.response(200, "获取成功", data);
    }

    @Override
    public Map<String, Object> deleteUser(String userId, HttpServletRequest request) {
        try {
            int intUserId = Tools.strToInt(userId);
            if (intUserId != 0) {
                if (intUserId == (int) request.getAttribute("userId")) {
                    return ResponseUtil.response(403, "你不能删除自己");
                }
                // 检查用户ID是否存在
                if (userMapper.checkUserIdExists(intUserId) < 1) {
                    return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
                }
                // 删除用户
                userMapper.deleteUser(intUserId);
                // 删除用户密钥
                userMapper.deleteUserKey(intUserId);
                // 删除用户API
                List<ApiApp> apiAppList = userMapper.getUserApiListAll(intUserId);
                if (!apiAppList.isEmpty()) {
                    for (ApiApp apiApp : apiAppList) {
                        // 删除APIKey
                        apiMapper.deleteApiKeyAll(apiApp.getId());
                        // 删除API参数
                        apiMapper.deleteApiParamAll(apiApp.getId());
                        // 删除APIlog
                        statMapper.deleteApiLog(apiApp.getId());
                        // 删除API统计
                        statMapper.deleteApiCount(apiApp.getId());
                        statMapper.deleteApiDailyStats(apiApp.getId());
                        // 删除API应用
                        apiMapper.deleteApiApp(apiApp.getId());
                    }
                }
                return ResponseUtil.success();
            } else {
                return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
            }
        } catch (Exception e) {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
        }
    }

    @Override
    public Map<String, Object> getUserApiList(Map<String, String> requestParam, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
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

        // 如果携带了ID，则验证是不是管理员，只有管理才可以带ID
        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
            }
            userId = Tools.strToInt(requestParam.get("id"));
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }
        }

        // 获取用户API列表
        Map<String, Object> resultMap = new LinkedHashMap<>();
        List<Map<String, Object>> apiInfoList = new ArrayList<>();
        // 获取用户API总数量
        int total = userMapper.getUserApiListAllCount(userId);
        for (ApiApp apiApp : userMapper.getUserApiList(userId, pageSize, Tools.getPageOffset(page, pageSize))) {
            Map<String, Object> map = new LinkedHashMap<>();
            map.put("id", apiApp.getId());
            map.put("title", apiApp.getTitle());
            map.put("type", apiApp.getType());
            map.put("status", apiApp.getStatus());
            map.put("created", apiApp.getCreated());
            map.put("updated", apiApp.getUpdated());
            map.put("user_id", apiApp.getUser_id());
            map.put("view_status", apiApp.getView_status());
            apiInfoList.add(map);
        }
        resultMap.put("total", total);
        resultMap.put("list", apiInfoList);

        return ResponseUtil.response(200, "获取成功", resultMap);
    }

    @Override
    public Map<String, Object> resetUserKey(String userId, HttpServletRequest request) {
        // 全局设置检测
        int tokenUserMode = (int) request.getAttribute("userMode");
        String settingValue = settingMapper.getSettingValue("user_key");
        if (settingValue != null && tokenUserMode != Finals.Admin && settingValue.equals("false")) {
            return ResponseUtil.response(400, "重置密钥已关闭");
        }
        int intUserId = Tools.strToInt(userId);
        int tokenUserId = (int) request.getAttribute("userId");
        if (intUserId != tokenUserId) {
            if (tokenUserMode != Finals.Admin) {
                return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
            }
        }
        // 验证用户ID是否存在
        if (userMapper.checkUserIdExists(intUserId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }

        // key 生成
        String key = Tools.getTextMd5(String.valueOf(intUserId) + System.currentTimeMillis());

        // 重置用户密钥
        userMapper.resetUserKey(intUserId, key);
        return ResponseUtil.success(key);
    }

    @Override
    public Map<String, Object> userApiListSearch(Map<String, String> requestParam, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        // 如果携带了ID，则验证是不是管理员，只有管理才可以带ID
        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
            }
            userId = Tools.strToInt(requestParam.get("id"));
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }
        }
        String keyword = requestParam.get("keywords");
        // status，view_status，type 他们的-1表示不筛选
        int type = Tools.strToInt(requestParam.get("type"));
        int status = Tools.strToInt(requestParam.get("status"));
        int view_status = Tools.strToInt(requestParam.get("view_status"));
        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }

        Map<String, Object> resultMap = new LinkedHashMap<>();
        List<Map<String, Object>> apiInfoList = new ArrayList<>();
        // 获取用户API总数量
        int total = userMapper.getUserApiListSearchCount(userId, keyword, type, status, view_status);
        // 获取用户API列表
        for (ApiApp apiApp : userMapper.getUserApiListSearch(userId, keyword, type, status, view_status, pageSize, Tools.getPageOffset(page, pageSize))) {
            Map<String, Object> map = new LinkedHashMap<>();
            map.put("id", apiApp.getId());
            map.put("title", apiApp.getTitle());
            map.put("type", apiApp.getType());
            map.put("status", apiApp.getStatus());
            map.put("created", apiApp.getCreated());
            map.put("updated", apiApp.getUpdated());
            map.put("user_id", apiApp.getUser_id());
            map.put("view_status", apiApp.getView_status());
            apiInfoList.add(map);
        }
        resultMap.put("total", total);
        resultMap.put("list", apiInfoList);
        return ResponseUtil.response(200, "获取成功", resultMap);
    }

    @Override
    public Map<String, Object> updateUserNick(Map<String, String> requestParam, Map<String, String> requestBody, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }

        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            // 检查Token是否为管理员
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
            }
            userId = Tools.strToInt(String.valueOf(requestParam.get("id")));
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }

        }

        String nick = requestBody.get("nick");
        if (nick == null || nick.trim().isEmpty()) {
            return ResponseUtil.response(400, "昵称不能为空");
        }
        if (nick.length() > 32) {
            return ResponseUtil.response(400, "昵称不能超过32个字符");
        }


        try {
            userMapper.updateUserNick(userId, nick);
            return ResponseUtil.success();
        } catch (Exception e) {
            return ResponseUtil.response(500, "修改昵称失败");
        }
    }


    @Override
    public Map<String, Object> updateUserMail(Map<String, String> requestParam, Map<String, String> requestBody, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }

        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            // 检查Token是否为管理员
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
            }
            userId = Tools.strToInt(String.valueOf(requestParam.get("id")));
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }

        }

        String mail = requestBody.get("mail");
        if (mail == null || mail.trim().isEmpty()) {
            return ResponseUtil.response(400, "邮箱不能为空");
        }
        if (!mail.matches("^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\\.[a-zA-Z0-9-]+)*\\.[a-zA-Z0-9]{2,6}$")) {
            return ResponseUtil.response(400, "邮箱格式错误");
        }

        try {
            userMapper.updateUserMail(userId, mail);
            return ResponseUtil.success();
        } catch (Exception e) {
            return ResponseUtil.response(500, "修改邮箱失败");
        }
    }

    @Override
    public Map<String, Object> updateUserPassword(Map<String, String> requestParam, Map<String, String> requestBody, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }

        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            // 检查Token是否为管理员
            if ((int) request.getAttribute("userMode") != Finals.Admin) {
                return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
            }
            userId = Tools.strToInt(String.valueOf(requestParam.get("id")));
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }

        }

        String oldPassword = Tools.getTextMd5(requestBody.get("oldPassword"));
        String newPassword = Tools.getTextMd5(requestBody.get("newPassword"));
        if (Objects.equals(oldPassword, newPassword)) {
            return ResponseUtil.response(400, "新旧密码相同");
        }
        userMapper.updateUserPassword(userId, oldPassword, newPassword);
        return ResponseUtil.success();
    }

    @Override
    public Map<String, Object> setUserPassword(Map<String, String> requestParam, Map<String, String> requestBody, HttpServletRequest request) {
        // 检查Token是否为管理员
        if ((int) request.getAttribute("userMode") != Finals.Admin) {
            return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
        }

        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            int userId = Tools.strToInt(String.valueOf(requestParam.get("id")));
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }
            String password = requestBody.get("password");
            if (password.isEmpty()) {
                return ResponseUtil.response(400, "密码不能为空");
            }
            userMapper.setUserPassword(userId, Tools.getTextMd5(password));
            return ResponseUtil.success();
        } else {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
        }
    }

    @Override
    public Map<String, Object> getUserList(Map<String, String> requestParam, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }
        if ((int) request.getAttribute("userMode") != Finals.Admin) {
            return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
        }

        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }
        // 获取用户列表
        Map<String, Object> resultMap = new LinkedHashMap<>();
        List<User> userList = userMapper.getUserList(pageSize, Tools.getPageOffset(page, pageSize));
        // 获取用户API总数量
        int total = userMapper.getUserListAllCount();
        resultMap.put("total", total);
        resultMap.put("list", userList);
        return ResponseUtil.response(200, "获取成功", resultMap);
    }

    @Override
    public Map<String, Object> userListSearch(Map<String, String> requestParam, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }
        if ((int) request.getAttribute("userMode") != Finals.Admin) {
            return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
        }
        String keyword = requestParam.get("keywords");
        // -1表示不筛选
        int mode = Tools.strToInt(requestParam.get("mode"));
        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }

        Map<String, Object> resultMap = new LinkedHashMap<>();
        // 获取总数量
        int total = userMapper.getUserListSearchCount(keyword, mode);
        // 获取列表
        List<User> userList = userMapper.getUserListSearch(keyword, mode, pageSize, Tools.getPageOffset(page, pageSize));
        resultMap.put("total", total);
        resultMap.put("list", userList);
        return ResponseUtil.success(resultMap);
    }

    @Override
    public Map<String, Object> setUserMode(Map<String, String> requestParam, Map<String, String> requestBody, HttpServletRequest request) {
        int userId = (int) request.getAttribute("userId");
        if (userMapper.checkUserIdExists(userId) < 1) {
            return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
        }
        if ((int) request.getAttribute("userMode") != Finals.Admin) {
            return ResponseUtil.response(403, Finals.MESSAGES_ERROR_NOT_ADMIN);
        }

        if (requestParam.get("id") != null && !requestParam.get("id").trim().isEmpty()) {
            userId = Tools.strToInt(String.valueOf(requestParam.get("id")));
            if (userMapper.checkUserIdExists(userId) < 1) {
                return ResponseUtil.response(401, Finals.MESSAGES_ERROR_USER_NOT_FOUND);
            }
            int mode = Tools.strToInt(requestBody.get("mode"));
            if (mode < 0 || mode > 1) {
                return ResponseUtil.response(400, "权限错误");
            }
            userMapper.setUserMode(userId, mode);
            return ResponseUtil.success();
        } else {
            return ResponseUtil.response(400, Finals.MESSAGES_ERROR_PARAM);
        }
    }

    @Override
    public Map<String, Object> getCheckApiList(Map<String, String> requestParam, HttpServletRequest request) {
        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }

        // 获取待审核API列表
        Map<String, Object> resultMap = new LinkedHashMap<>();
        List<Map<String, Object>> apiInfoList = new ArrayList<>();
        // 获取待审核API总数量
        int total = userMapper.getCheckApiListAllCount();
        for (ApiApp apiApp : userMapper.getCheckApiList(pageSize, Tools.getPageOffset(page, pageSize))) {
            Map<String, Object> map = new LinkedHashMap<>();
            map.put("id", apiApp.getId());
            map.put("title", apiApp.getTitle());
            map.put("type", apiApp.getType());
            map.put("status", apiApp.getStatus());
            map.put("created", apiApp.getCreated());
            map.put("updated", apiApp.getUpdated());
            map.put("user_id", apiApp.getUser_id());
            map.put("view_status", apiApp.getView_status());
            apiInfoList.add(map);
        }
        resultMap.put("total", total);
        resultMap.put("list", apiInfoList);

        return ResponseUtil.response(200, "获取成功", resultMap);
    }

    @Override
    public Map<String, Object> checkApiListSearch(Map<String, String> requestParam, HttpServletRequest request) {
        String keyword = requestParam.get("keywords");
        // statu，type 他们的-1表示不筛选
        int type = Tools.strToInt(requestParam.get("type"));
        int status = Tools.strToInt(requestParam.get("status"));
        int pageSize = Tools.strToInt(requestParam.get("pageSize"));
        int page = Tools.strToInt(requestParam.get("page"));
        if (pageSize < 1) {
            pageSize = 30;
        }
        if (page < 1) {
            page = 1;
        }

        Map<String, Object> resultMap = new LinkedHashMap<>();
        List<Map<String, Object>> apiInfoList = new ArrayList<>();
        // 获取待审核API总数量
        int total = userMapper.getCheckApiListSearchCount(keyword, type, status);
        // 获取待审核API列表
        for (ApiApp apiApp : userMapper.getCheckApiListSearch(keyword, type, status, pageSize, Tools.getPageOffset(page, pageSize))) {
            Map<String, Object> map = new LinkedHashMap<>();
            map.put("id", apiApp.getId());
            map.put("title", apiApp.getTitle());
            map.put("type", apiApp.getType());
            map.put("status", apiApp.getStatus());
            map.put("created", apiApp.getCreated());
            map.put("updated", apiApp.getUpdated());
            map.put("user_id", apiApp.getUser_id());
            map.put("view_status", apiApp.getView_status());
            apiInfoList.add(map);
        }
        resultMap.put("total", total);
        resultMap.put("list", apiInfoList);
        return ResponseUtil.response(200, "获取成功", resultMap);
    }

    @Override
    public Map<String, Object> checkAppChange(Map<String, String> requestBody, HttpServletRequest request) {
        int api_id = Tools.strToInt(requestBody.get("api_id"));
        int view_status = Tools.strToInt(requestBody.get("view_status"));

        // 检测API是否存在
        if (apiMapper.checkApiExist(api_id) < 1) {
            return ResponseUtil.response(400, "API不存在");
        }
        if (view_status < 0 || view_status > 2) {
            view_status = 2;
        }
        userMapper.checkAppChange(api_id, view_status);
        return ResponseUtil.success();
    }


}
