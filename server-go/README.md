# HelloAPI Go 后端

使用 Go 1.27.x、chi、sqlx 和 MySQL 5.7+。沿用 Java 版数据库与前端的 `/api/v1` 接口。Java 到 Go 的迁移范围和验收要求见 [重写规划](../docs/go-rewrite-plan.md)。

## 构建与运行

在 `server-go/` 目录运行。先在 MySQL 创建配置所指的数据库，再启动服务。

```powershell
Copy-Item config.example.yaml config.yaml
go build -o helloapi.exe ./cmd/helloapi
.\helloapi.exe -config config.yaml
```

配置按 `-config` 指定文件、可执行文件同级的 `config.yaml`、当前目录的 `config.yaml` 顺序查找，使用第一个存在的文件。文件缺省字段沿用默认值，环境变量覆盖文件值。

| 配置项 | 默认值 | 环境变量 |
| --- | --- | --- |
| `server.port` | `8080` | `HELLOAPI_PORT` |
| `database.host` | `localhost` | `HELLOAPI_DB_HOST` |
| `database.port` | `3306` | `HELLOAPI_DB_PORT` |
| `database.name` | `api` | `HELLOAPI_DB_NAME` |
| `database.username` | `root` | `HELLOAPI_DB_USERNAME` |
| `database.password` | `123456` | `HELLOAPI_DB_PASSWORD` |
| `jwt.secret` | 空值，每次启动生成随机密钥 | `HELLOAPI_JWT_SECRET` |
| `stat.initial-delay-ms` | `10000` | `HELLOAPI_STAT_INITIAL_DELAY_MS` |
| `stat.fixed-delay-ms` | `10000` | `HELLOAPI_STAT_FIXED_DELAY_MS` |

MySQL 连接使用 `charset=utf8mb4`、`parseTime=true`、`loc=Local`。统计调度到达初始延迟后每轮清空积压，再等待固定间隔。按 Ctrl+C 停止后台任务并等待 HTTP 请求结束。

## 数据库迁移

启动时用独立的 MySQL 连接获取 `GET_LOCK('helloapi_migrate', 60)`，执行内嵌在 `internal/migrate/sql/` 中的脚本。只有迁移连接启用多语句。

- 空库按版本执行 V1、V1.1 和 V1.2。
- 非空库没有 `flyway_schema_history` 时建立版本 1 基线，再执行更新的脚本。版本 1 的基础表需已存在。
- 已有 Flyway 历史时跳过已执行版本。校验和不一致时记录告警。任一历史记录的 `success=0` 会阻止启动。
- 以原始脚本逐行计算 Flyway CRC32。结构变更应新增 `V<版本>__<描述>.sql`，已有脚本保留原样。

## 路由对照

共 56 条独立的控制器方法与路径组合。路径区分大小写和尾斜杠。可选登录表示在隐藏或未审核 API 的详情接口读取令牌中的身份。管理员路由实时查库。未给出路径变量的别名使用 `"0"`。

### 用户（22 条）

| 方法 | 路径 | 鉴权 | Java 方法 | Go handler |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/user/Captcha` | 公开 | `GetCaptcha` | `GetCaptcha` |
| POST | `/api/v1/user/Register` | 公开 | `Register` | `Register` |
| POST | `/api/v1/user/Login` | 公开 | `Login` | `Login` |
| GET | `/api/v1/user/` | 登录 | `GetUserInfo` | `GetUserInfo` |
| PUT | `/api/v1/user/nick` | 登录 | `UpdateUserNick` | `UpdateUserNick` |
| PUT | `/api/v1/user/mail` | 登录 | `UpdateUserMail` | `UpdateUserMail` |
| PUT | `/api/v1/user/password` | 登录 | `UpdateUserPassword` | `UpdateUserPassword` |
| PUT | `/api/v1/user/setPassword` | 管理员 | `SetUserPassword` | `SetUserPassword` |
| PUT | `/api/v1/user/setMode` | 管理员 | `SetUserMode` | `SetUserMode` |
| GET | `/api/v1/user/key` | 登录 | `GetUserKey` | `GetUserKey` |
| DELETE | `/api/v1/user/{userId}` | 管理员 | `DeleteUser` | `DeleteUser` |
| DELETE | `/api/v1/user/` | 管理员 | `DeleteUser` | `DeleteUser` |
| GET | `/api/v1/user/AppList` | 登录 | `GetUserApiList` | `GetUserAPIList` |
| GET | `/api/v1/user/AppList/Search` | 登录 | `UserApiListSearch` | `UserAPIListSearch` |
| GET | `/api/v1/user/list` | 管理员 | `GetUserList` | `GetUserList` |
| GET | `/api/v1/user/list/Search` | 管理员 | `UserListSearch` | `UserListSearch` |
| PUT | `/api/v1/user/key/` | 登录 | `resetUserKey` | `ResetUserKey` |
| PUT | `/api/v1/user/key/{userId}` | 登录 | `resetUserKey` | `ResetUserKey` |
| PUT | `/api/v1/user/key` | 登录 | `resetUserKey` | `ResetUserKey` |
| GET | `/api/v1/user/CheckAppList` | 管理员 | `GetCheckApiList` | `GetCheckAPIList` |
| GET | `/api/v1/user/CheckAppList/Search` | 管理员 | `CheckApiListSearch` | `CheckAPIListSearch` |
| PUT | `/api/v1/user/CheckApp` | 管理员 | `CheckAppChange` | `CheckAppChange` |

### API 与 API Key（22 条）

| 方法 | 路径 | 鉴权 | Java 方法 | Go handler |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/app/list` | 公开 | `GetApiList` | `GetAPIList` |
| POST | `/api/v1/app/` | 登录 | `CreateApi` | `CreateAPI` |
| PUT | `/api/v1/app/{apiId}` | 登录 | `UpdateApi` | `UpdateAPI` |
| PUT | `/api/v1/app/` | 登录 | `UpdateApi` | `UpdateAPI` |
| DELETE | `/api/v1/app/{apiId}` | 登录 | `DeleteApi` | `DeleteAPI` |
| DELETE | `/api/v1/app/` | 登录 | `DeleteApi` | `DeleteAPI` |
| GET | `/api/v1/app/search` | 公开 | `SearchApiList` | `SearchAPIList` |
| GET | `/api/v1/app/{apiId}` | 可选登录 | `GetApiApp` | `GetAPIApp` |
| GET | `/api/v1/app/` | 可选登录 | `GetApiApp` | `GetAPIApp` |
| POST | `/api/v1/app/key/{api_id}` | 登录 | `CreateApiKey` | `CreateAPIKey` |
| POST | `/api/v1/app/key/` | 登录 | `CreateApiKey` | `CreateAPIKey` |
| GET | `/api/v1/app/key/list/` | 登录 | `GetUserApiKeyList` | `GetUserAPIKeyList` |
| GET | `/api/v1/app/key/list/{user_id}` | 登录 | `GetUserApiKeyList` | `GetUserAPIKeyList` |
| GET | `/api/v1/app/key/list/{user_id}/Search` | 登录 | `UserApiKeyListSearch` | `UserAPIKeyListSearch` |
| PUT | `/api/v1/app/key/{key}` | 登录 | `UpdateApiKey` | `UpdateAPIKey` |
| PUT | `/api/v1/app/key/` | 登录 | `UpdateApiKey` | `UpdateAPIKey` |
| GET | `/api/v1/app/key/{key}` | 登录 | `GetApiKey` | `GetAPIKey` |
| GET | `/api/v1/app/key/` | 登录 | `GetApiKey` | `GetAPIKey` |
| DELETE | `/api/v1/app/key/{key}` | 登录 | `DeleteApiKey` | `DeleteAPIKey` |
| DELETE | `/api/v1/app/key/` | 登录 | `DeleteApiKey` | `DeleteAPIKey` |
| PUT | `/api/v1/app/key/reset/{key}` | 登录 | `ResetApiKey` | `ResetAPIKey` |
| PUT | `/api/v1/app/key/reset/` | 登录 | `ResetApiKey` | `ResetAPIKey` |

### 友链（5 条）

| 方法 | 路径 | 鉴权 | Java 方法 | Go handler |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/link/` | 公开 | `GetLinkList` | `GetLinkList` |
| POST | `/api/v1/link/` | 管理员 | `CreateLink` | `CreateLink` |
| PUT | `/api/v1/link/{link_id}` | 管理员 | `UpdateLink` | `UpdateLink` |
| DELETE | `/api/v1/link/{link_id}` | 管理员 | `DeleteLink` | `DeleteLink` |
| GET | `/api/v1/link/Search` | 管理员 | `GetLinkListSearch` | `GetLinkListSearch` |

### 全局设置（3 条）

| 方法 | 路径 | 鉴权 | Java 方法 | Go handler |
| --- | --- | --- | --- | --- |
| GET | `/api/v1/setting/list` | 管理员 | `GetSettingList` | `GetSettingList` |
| GET | `/api/v1/setting/{settingKey}` | 公开 | `GetSetting` | `GetSetting` |
| PUT | `/api/v1/setting/` | 管理员 | `UpdateSettingValue` | `UpdateSettingValue` |

### 调用统计（4 条）

| 方法 | 路径 | 鉴权 | Java 方法 | Go handler |
| --- | --- | --- | --- | --- |
| POST | `/api/v1/stat/log` | 公开 | `logApi` | `LogAPI` |
| POST | `/api/v1/stat/log/` | 公开 | `logApi` | `LogAPI` |
| GET | `/api/v1/stat/dashboard` | 登录 | `getDashboard` | `GetDashboard` |
| GET | `/api/v1/stat/count` | 公开 | `getApiAllCount` | `GetAPIAllCount` |

Java 静态资源 `GET /Logo.psd` 也随二进制提供。

## 验证

```powershell
go build ./...
go vet ./...
go test ./...
gofmt -l cmd internal scripts
```

### MySQL 5.7 与 HTTP 冒烟

以下命令在尚未创建测试容器时执行。已存在同名容器时可运行 `docker start helloapi-mysql57-test`。等 MySQL 启动后，在同一终端配置并启动 Go 后端。

```powershell
docker run -d --name helloapi-mysql57-test -p 127.0.0.1:3307:3306 -e MYSQL_ROOT_PASSWORD=helloapi_test_only -e MYSQL_ROOT_HOST=% -e MYSQL_DATABASE=helloapi_go_test mysql:5.7

$env:HELLOAPI_DB_HOST = '127.0.0.1'
$env:HELLOAPI_DB_PORT = '3307'
$env:HELLOAPI_DB_NAME = 'helloapi_go_test'
$env:HELLOAPI_DB_PASSWORD = 'helloapi_test_only'
$env:HELLOAPI_STAT_INITIAL_DELAY_MS = '0'
$env:HELLOAPI_STAT_FIXED_DELAY_MS = '1000'
.\helloapi.exe
```

在另一个终端执行冒烟脚本。按提示读取两张验证码图片并输入字符。脚本从注册和登录开始，验证用户信息、创建和审核 API、创建计次密钥、日志和统计，最后删除 API。它还会对最后一次配额发出 8 个并发请求，要求只有 1 个成功。

```powershell
go run ./scripts/smoke -base http://127.0.0.1:8080
```

首次注册的用户自动成为管理员。复跑时，脚本仍会注册一个新用户，可通过 `-admin-user` 和 `HELLOAPI_SMOKE_ADMIN_PASSWORD` 提供已有管理员凭证用于审核。脚本运行在专用测试库中，并保留注册的测试账号。

保留测试容器时可用以下命令随时查看 Flyway 历史。需要清理时执行第二条命令。

```powershell
docker exec -e MYSQL_PWD=helloapi_test_only helloapi-mysql57-test mysql -uroot helloapi_go_test -e 'SELECT installed_rank, version, type, script, checksum, success FROM flyway_schema_history ORDER BY installed_rank'
docker rm -f -v helloapi-mysql57-test
```

## 与 Java 版的已知差异与保留行为

- 计次型 API Key 用带 `count > 0` 条件的原子 UPDATE 扣减，未更新行时返回原有耗尽响应。日志写入仍独立提交。
- Go 将 `GET /api/v1/app/key/list/` 固定交给 `GetUserAPIKeyList`。Java 对这个路径的映射歧义可触发 HTTP 500。带用户 ID 的 `/Search` 分支仍保留 Java 对 `tokenUserId` 的读取结果，普通用户通常收到无操作权限。
- 验证码为带正确媒体类型的 PNG。Java 会在 GIF 和 PNG 间选择，且统一标记为 `image/gif`。注册输错后保留验证码，登录输错或两者校验成功后清除验证码。
- 可选请求体缺失或非法时按空对象处理。必需请求体缺失或非法时返回 HTTP 200、业务码 400 和“参数错误”。Java 的框架级请求体错误可能返回 HTTP 400。Go 对未捕获错误统一返回 HTTP 500 和“服务器内部错误”。未匹配路由和方法使用 chi 的 404 或 405 响应。
- API Key 时间支持 ISO 本地日期时间和以空格分隔的本地日期时间，响应统一输出秒精度。Java 默认只将 ISO 本地日期时间用于输入解析。
- 空 JWT 密钥每次启动都会生成新签名密钥。Java 和 Go 默认各自生成签名密钥，切换进程后需重新登录。验证码会话也随进程重启清空。
- 单独删 API 和删用户时，多表清理按各自的 Java 顺序在事务中提交。
- 创建和更新 API 时，主记录与参数在同一事务中写入，参数写入失败会整体回滚并返回“参数错误”。Java 版会留下缺少参数的 API，更新时还会丢失原有参数。
- 修改密码时，旧密码不匹配返回业务码 400 和“旧密码错误”，Java 版会返回成功但实际不修改。修改密码与管理员设置密码都会按登录规则校验新密码（6～64 个字符、不含空格），避免设置后无法登录。
