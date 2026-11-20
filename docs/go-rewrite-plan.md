# HelloAPI 后端 Go 重写规划

> **状态：已完成。** Go 版已通过线上数据副本验证并正式替换 Java 版；文中的 `server-go/` 现为 `server/`，Java 源码已删除（可在 git 历史中查看）。

> 规划：Claude；实施：GPT。本文是实施方的唯一需求来源，遇到本文未覆盖的细节，以 **Java 源码的实际行为** 为准。

## 0. 目标与边界

- 用 Go 1.27.x 重写 `server/`（Spring Boot 4 + MyBatis）的全部功能，数据库仍为 MySQL 5.7+，前端 `web/` **零改动**即可切换到新后端。
- 新代码放在 `server-go/`，**不修改** `server/`（Java 版保留用于对照，确认无误后再统一切换删除）和 `web/`。
- 首要原则是 **接口契约与数据库完全兼容**：同一份已有数据库可直接被 Go 版接管；前端请求、响应 JSON 结构、字段名、业务 code、提示文案与 Java 版一致。
- 不顺手"优化"业务语义。唯一允许的行为变更见第 9 节。

## 1. 技术选型

| 用途 | 选型 | 说明 |
|---|---|---|
| 路由 | `github.com/go-chi/chi/v5` | 精确匹配路径，不做尾斜杠重定向（Spring Boot 4 默认也不做尾斜杠匹配） |
| 数据库 | `github.com/go-sql-driver/mysql` + `github.com/jmoiron/sqlx` | 手写 SQL，对应 MyBatis 注解 SQL，SQL 尽量原样搬运 |
| JWT | `github.com/golang-jwt/jwt/v5` | HS256 |
| 验证码 | `github.com/mojocn/base64Captcha`（或同等库） | 需能直接输出图片二进制流 |
| 配置 | `gopkg.in/yaml.v3` | |
| 日志 | 标准库 `log/slog` | |
| 缓存 | 自行实现的小型 TTL 缓存 | 只有两个缓存项，不引入依赖 |
| 测试 | 标准库 `testing`，必要时 `github.com/DATA-DOG/go-sqlmock` | |

`go.mod` 中 `go 1.27`，模块名建议 `github.com/zxz/helloapi`（或 `helloapi`）。

## 2. 目录结构

```
server-go/
  cmd/helloapi/main.go          # 入口：加载配置 → 连库 → 迁移 → 启动调度器 → 启动 HTTP（优雅退出）
  internal/config/              # 配置加载
  internal/migrate/             # Flyway 兼容迁移器
  internal/migrate/sql/         # 从 server/src/main/resources/db/migration 原样复制的 SQL，用 embed 打包
  internal/httpx/               # 统一响应、宽松参数解析、CORS、鉴权中间件、recover
  internal/auth/                # JWT 签发与解析
  internal/captcha/             # 验证码生成 + 基于 Cookie 的会话存储
  internal/model/               # 领域结构体、DTO、LocalDateTime 类型
  internal/repo/                # 对应 mapper/*，一个文件一个 Mapper
  internal/service/             # 对应 service/impl/*、StatAggregationService、StatDashboardService
  internal/handler/             # 对应 controller/*，只做参数提取和调用 service
  internal/stat/                # 聚合调度器（对应 StatAggregationScheduler）
  config.example.yaml
  README.md
```

分层与 Java 一一对应：`handler → service → repo`，service 依赖 repo 的 **接口**，便于单元测试替换。

## 3. 配置

`config.yaml`（查找顺序：`-config` 命令行参数 → 可执行文件同级目录 → 当前工作目录；均不存在则用默认值）：

```yaml
server:
  port: 8080
database:
  host: localhost
  port: 3306
  name: api
  username: root
  password: "123456"
jwt:
  secret: ""            # 留空 = 每次启动随机生成（与 Java 版一致：重启后旧 token 失效）
stat:
  initial-delay-ms: 10000
  fixed-delay-ms: 10000
```

DSN 固定附加 `charset=utf8mb4&parseTime=true&loc=Local`。支持环境变量覆盖数据库密码等敏感项（如 `HELLOAPI_DB_PASSWORD`），其余项可选。

## 4. 数据库迁移（必须兼容 Flyway）

已有部署的库里存在 Flyway 的 `flyway_schema_history` 表，Go 版必须能无缝接管，所以 **不要用 goose / golang-migrate**，自写一个最小 Flyway 兼容迁移器：

1. 迁移脚本用 `embed` 打包，文件名规则 `V<版本>__<描述>.sql`，版本按点分数字比较（`1` < `1.1` < `1.2` < `1.10`）。
2. `flyway_schema_history` 不存在时：
   - 当前 schema 中 **没有任何表** → 建历史表，从头执行所有脚本；
   - schema 非空 → 等价于 `baseline-on-migrate: true`、`baselineVersion=1`：建历史表并写入一条 `type='BASELINE'`、`version='1'`、`description='<< Flyway Baseline >>'`、`script='<< Flyway Baseline >>'` 的记录，然后只执行版本 > 1 的脚本。
3. 历史表结构与 Flyway 一致：`installed_rank INT PK, version VARCHAR(50), description VARCHAR(200), type VARCHAR(20), script VARCHAR(1000), checksum INT, installed_by VARCHAR(100), installed_on TIMESTAMP DEFAULT CURRENT_TIMESTAMP, execution_time INT, success TINYINT(1)`，外加 Flyway 同名索引 `flyway_schema_history_s_idx (success)`。
4. 执行成功后插入 `type='SQL'` 记录。`description` 为文件名描述部分，下划线替换为空格；`checksum` 按 Flyway 算法计算（去除 BOM，逐行、不含换行符，对每行 UTF-8 字节做 CRC32 累加，结果转 int32）。
5. 已执行版本的 checksum 不一致时 **只打 warning，不中止启动**。存在 `success=0` 的记录时拒绝启动并给出明确提示。
6. 迁移连接使用 `multiStatements=true`（只在迁移用的连接上开启），整个迁移过程用 `GET_LOCK('helloapi_migrate', 60)` 防止多实例并发。
7. 新增表结构变更的方式不变：在 `internal/migrate/sql/` 新增版本化脚本，禁止改动已有脚本。

## 5. HTTP 层契约（重点）

### 5.1 统一响应

- 结构 `{"code":..,"msg":..,"data":..}`，字段顺序 code、msg、data；`data` 为 nil 时 **省略该字段**（对应 `ResponseUtil`）。
- HTTP 状态码一律 200（包括鉴权失败的 401/403），与 Java 一致。panic 由 recover 中间件兜底返回 HTTP 500 + `{"code":500,"msg":"服务器内部错误"}`。
- 空列表必须序列化为 `[]`，**绝不能是 `null`**（Go nil slice 的常见坑，前端会直接 `.map`）。
- 所有 `Finals` 中的常量文案原样搬到 Go 常量中。

### 5.2 JSON 字段与类型

- 字段名以 Java getter 推导出的 Jackson 名称为准，例如 `smallTitle`、`sendType`、`user_id`、`view_status`、`api_id`。逐个对照 `domain/` 与 `domain/DTO/` 的 getter 写 `json` tag，不要凭感觉改成驼峰或下划线。
- 日期时间（Java `LocalDateTime`）输出为 **不带时区** 的 `"2006-01-02T15:04:05"`。实现一个 `model.LocalDateTime` 类型，负责 JSON 编解码与 `sql.Scanner`/`driver.Valuer`；反序列化需接受 `2006-01-02T15:04:05`、`2006-01-02 15:04:05`（前端 KeyPage 发送前者）。日期（`LocalDate`）输出 `"2006-01-02"`。
- 可空列（如 `desc`、`avatar`、`started`、`expired`、`ip`、`header`、`body`、`api_key`）用指针类型，序列化为 `null`（Jackson 默认会输出 null 字段）。
- 计数类字段（Java `long`）用 `int64`。

### 5.3 请求参数的宽松解析

Java 大量使用 `@RequestBody Map<String,String>`，Jackson 会把 JSON 中的数字、布尔值自动转成字符串。Go 版必须做同样的宽松处理：

- 提供 `httpx.StringMap`：把请求体解码为 `map[string]any`，数字转成其文本形式（用 `json.Number` 保证 `1` 不变成 `1e+00`），布尔转 `"true"/"false"`，null 视为缺失。
- `Map<String,Object>` 的接口（`ApiController` 中的 create/update 等）保持 `map[string]any`，service 里按 Java 的取值/类型转换逻辑逐一还原。
- `@RequestParam Map<String,String>` 对应 query 参数，取第一个值。
- `Tools.strToInt` 失败返回 -1 的语义保持一致。
- 请求体缺失或不是合法 JSON 时：对 `required=false` 的接口当作空 map；对 `required=true` 的接口（ApiController 的 POST/PUT）返回 400 参数错误。

### 5.4 路由

对 `controller/*.java` 中每一个映射路径 **逐条** 注册，包括同时声明了 `"/{id}"` 与 `"/"` 两种形式的（`"/"` 时 id 取 `"0"`）。注意：

- `ApiController` 中 `GET /key/list/` 同时出现在 `GetUserApiKeyList` 与 `UserApiKeyListSearch` 的映射里，Go 版统一路由到 `GetUserApiKeyList`。
- `UserController` 的 `PUT /key/`、`/key/{userId}`、`/key` 三种形式都要注册。
- 大小写敏感，`/Captcha`、`/AppList/Search` 等保持原样。
- 在 `server-go/README.md` 中列出 **路由对照表**（方法、路径、鉴权级别、Java 方法名、Go handler 名）。

### 5.5 鉴权（对应 JwtInterceptor + @RequiresLogin）

- 中间件形式：`RequireLogin` 与 `RequireAdmin` 两种，挂在对应路由上。
- Token 从 `Authorization: Bearer <token>` 读取；解析失败或 id < 1 → `{code:401, msg:"Token过期/无效"}`。
- Admin：实时查库（`UserMapper.checkUserIsAdmin`），不是管理员 → `{code:403, msg:"非管理员用户，没有操作权限"}`；是管理员但 token 中 mode=0 时，上下文中的 userMode 提升为 1。
- 通过后把 `userId`、`userMode` 放进 `context`，service 通过帮助函数读取。
- 未加鉴权但需要可选身份的接口（如 `GET /api/v1/app/{apiId}`）自行解析 header，逻辑同 Java。
- JWT：HS256，claims 为 `id`、`name`、`mode`、`exp`（签发后 70 天），与 `JwtUtil` 一致。

### 5.6 CORS（对应 WebConfig）

允许任意来源且携带凭证：回显请求的 `Origin`，`Access-Control-Allow-Credentials: true`，允许所有方法、回显 `Access-Control-Request-Headers`，`Access-Control-Max-Age: 3600`，正确处理 OPTIONS 预检（直接 204/200，不经过鉴权）。

### 5.7 验证码

- Java 版把验证码存在 HttpSession（Cookie `JSESSIONID`，30 分钟超时），前端用 `<img src=".../Captcha?时间戳">` 获取，登录/注册时靠 Cookie 关联。
- Go 版：内存会话存储（`sync.Map` 或带锁 map + 定期清理过期项，TTL 30 分钟），Cookie 名沿用 `JSESSIONID`，`Path=/`、`HttpOnly`，不设置 SameSite（与 Java 默认行为一致）。
- 图片 120×32、4 位、数字+大写字母，直接输出图片二进制并设置正确的 `Content-Type` 与禁止缓存头。
- 校验时忽略大小写（与 `CaptchaJakartaUtil.ver` 一致），校验后无论成败都清除（逐一核对 Java 中 register/login 各分支 clear 的时机并保持一致）。

## 6. 业务逻辑迁移

逐个文件翻译，**保持判断顺序、错误码、提示文案、返回 data 内容完全一致**：

| Java | Go |
|---|---|
| `UserServiceImpl` + `UserMapper` | `service/user.go` + `repo/user.go` |
| `ApiServiceImpl` + `ApiMapper` | `service/api.go` + `repo/api.go` |
| `LinkServiceImpl` + `LinkMapper` | `service/link.go` + `repo/link.go` |
| `SettingServiceImpl` + `SettingMapper` | `service/setting.go` + `repo/setting.go` |
| `StatServiceImpl` + `ApiLogManager` + `StatMapper` | `service/stat.go` + `repo/stat.go` |
| `StatAggregationService` + `StatAggregationMapper` | `service/stat_aggregation.go` + `repo/stat_aggregation.go` |
| `StatDashboardService` + `CacheConfig` | `service/stat_dashboard.go`（TTL 30 秒缓存，key 对应 `statDashboard`、`apiAllCount`） |
| `StatAggregationScheduler` | `stat/scheduler.go`（`time.Ticker`，fixed-delay 语义：上一次结束后再等待间隔；支持 ctx 取消） |

要点：

- 密码保持 **无盐 MD5 32 位小写十六进制**（`Tools.getTextMd5`），否则已有用户无法登录。
- 密钥生成保持 `md5(id + 毫秒时间戳)` 并在冲突时重试的逻辑。
- `Tools.deleteDuplicatesList`、分页偏移 `(page-1)*pageSize` 等工具函数照搬。
- 事务：`StatAggregationService#aggregateNextBatch` 的"锁检查点行（`SELECT ... FOR UPDATE`）→ 读最多 5000 条新日志 → 写每日统计 → 写累计 → 更新检查点"必须在 **同一个事务** 中完成；其他 Java 中带 `@Transactional` 的地方同样用事务。
- 删除 API / 用户时同步清理聚合数据，按 Java 实现的 SQL 与顺序搬运（涉及多表删除的放在事务里）。
- `StatServiceImpl#logApi` 中 header/body 合法性校验用 `json.Valid`，不合法存 NULL。
- SQL 中的 `ON DUPLICATE KEY UPDATE ... VALUES(...)` 写法保留（MySQL 5.7 与 8.0 均可用）。不要使用 MySQL 8 专有语法（窗口函数、CTE、`AS new` 别名等）。
- MyBatis 的 `#{}` 一律改为 `?` 占位符；`${}`（如果存在）需要审查后用白名单处理，禁止字符串拼接用户输入。LIKE 搜索按 Java 的拼接方式实现（`CONCAT('%', ?, '%')`）。

## 7. 测试要求

- 单元测试：把 `server/src/test` 中两个 Mockito 测试的用例全部移植为 Go 测试（依赖 repo 接口的 fake 实现）。
- 新增单元测试覆盖：`LocalDateTime` 编解码、`StringMap` 宽松解析、JWT 签发/解析/过期、迁移器的版本排序与 Flyway checksum 计算（可用已知 Flyway 计算结果做对照，或至少对算法分支做测试）、空切片序列化为 `[]`。
- 集成验证（必须做）：用 Docker 起 `mysql:5.7` 容器（容器名 `helloapi-mysql57-test`、映射到宿主 `3307` 端口，避免和用户本机冲突），分别验证：
  1. 空库启动：迁移从 V1 执行到最新版本，`flyway_schema_history` 记录正确；
  2. 用 Java 版跑过迁移的库（可先用本机 Java 22 / `server/mvnw.cmd` 启动 Java 版建库；若 Java 版因 JDK 版本等原因起不来，就手工按 Flyway 格式模拟历史表）启动 Go 版：不重复执行、不报错；
  3. 用 curl 或 Go 编写的冒烟脚本跑通主要流程：验证码 → 注册 → 登录 → 获取用户信息 → 创建 API → 审核 → 创建 API Key → `POST /api/v1/stat/log` 上报 → 等待聚合 → `GET /api/v1/stat/dashboard` 与 `/count` 有数据 → 删除 API。
  4. 有条件时，对同一请求分别打 Java 版与 Go 版，比对响应 JSON（忽略 token、时间、密钥等动态值）。
- 冒烟脚本保存在 `server-go/scripts/` 下，便于复跑。验证结束后容器保留不删，在汇报中说明如何清理。

## 8. 交付物与验收

- `server-go/` 下：`go build ./...`、`go vet ./...`、`go test ./...` 全部通过，`gofmt` 无差异。
- `server-go/README.md`：构建/运行方式、配置说明、路由对照表、与 Java 版的已知差异清单。
- `server-go/config.example.yaml`。
- 不提交 git commit，由规划方审查后统一提交。
- 汇报内容：完成情况、集成验证的实际输出摘要、与 Java 版的差异或未决问题。

## 9. 允许的行为变更（仅此一处）

`logApi` 中计次型 API Key 的"先查剩余次数再扣减"存在并发超扣问题。Go 版改为原子扣减：`UPDATE helloapi_api_keys SET count = count - 1 WHERE key = ? AND count > 0`，影响行数为 0 时返回原有的"API密钥 / APIKey耗尽"响应。其余行为一律与 Java 保持一致，发现其他 Java bug 只记录在 README 的"已知差异/待办"里，不擅自修复。
