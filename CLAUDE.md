# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

HelloAPI：拟态风格的 API 管理程序（API 展示、审核、密钥、调用统计、友链、全局设置）。前后端分离的单仓库：

- `server/`：Go 1.27 + chi + sqlx（手写 SQL）+ MySQL 5.7+，自带 Flyway 兼容的迁移器（由原 Spring Boot 版重写而来，重写规划见 `docs/go-rewrite-plan.md`）
- `web/`：Vue 3 + Vite + TypeScript + Element Plus + Tailwind CSS v4 + Pinia，使用 pnpm

## 常用命令

后端（在 `server/` 目录下）：

```bash
go run ./cmd/helloapi                  # 本地运行（需要可连接的 MySQL，配置见下文）
go build -o helloapi ./cmd/helloapi    # 编译二进制
go vet ./... && go test ./...          # 静态检查 + 全部测试
go test ./internal/service/ -run TestAggregateReturnsZeroWithoutNewLogs   # 运行单个测试
gofmt -l .                             # 检查格式（应无输出）
```

前端（在 `web/` 目录下）：

```bash
pnpm i            # 安装依赖
pnpm dev          # 开发服务器
pnpm build        # vue-tsc 类型检查 + vite 构建（并行执行）
pnpm type-check   # 仅类型检查
```

项目未配置 golangci-lint 等 lint 工具，前端也没有测试框架。

## 配置

- 后端：`config.yaml`（模板 `server/config.example.yaml`），查找顺序为 `-config` 参数 → 可执行文件同级目录 → 当前目录；`HELLOAPI_DB_PASSWORD`、`HELLOAPI_PORT` 等环境变量可覆盖（完整列表见 `server/README.md`）。统计聚合调度间隔由 `stat.initial-delay-ms` / `stat.fixed-delay-ms` 控制（默认 10 秒）。`stat.log-retention-days` 默认保留原始日志 90 天，设为 0 时关闭清理。清理轮次间隔由 `stat.log-cleanup-interval-ms` 控制（默认 1 小时）。`stat.log-payload-max-bytes` 默认限制 header 和 body 各 2048 字节，0 表示不保存，-1 表示不限。
- 前端：`web/src/config/config.ts`（`serverURL` 后端地址、分页大小、水印、页脚，以及 `folder_paths` / `file_paths`——放行给非 SPA 资源的路径，否则会被前端 404 路由拦截）。
- 生产环境 Nginx 需要 `try_files $uri $uri/ /index.html;` 伪静态。

## 后端架构

模块名 `github.com/zxz/helloapi`，入口 `cmd/helloapi/main.go`（加载配置 → 连库 → 迁移 → 启动验证码清理、统计聚合与日志清理 → HTTP，支持优雅退出）。分层为 `internal/handler`（路由与参数提取）→ `internal/service`（业务）→ `internal/repo`（SQL），service 依赖 repo 的接口以便测试替换。

- **统一响应**：`internal/httpx` 的 `Reply`/`Success` 构造 `{code, msg, data}`（data 为 nil 时省略，nil 切片输出为 `[]`）。业务错误码放在 JSON 的 `code` 字段（401 = 需重新登录，403 = 非管理员，400 = 参数错误等），HTTP 状态码通常仍是 200。提示文案常量与 URL 前缀（`/api/v1/user|app|stat|link|setting`）也在 `httpx` 中。
- **接口契约必须与前端保持兼容**：JSON 字段名沿用原 Java 版（如 `smallTitle`、`user_id`、`view_status`），日期时间用 `model.LocalDateTime` 输出不带时区的 `2006-01-02T15:04:05`。请求体用 `httpx.StringMap` 宽松解析（数字/布尔转为字符串，对应原 Jackson 行为），或用 `httpx.ObjectMap` 保留 `json.Number`。
- **鉴权**：路由在 `internal/handler/router.go` 中逐条注册，需登录的挂 `httpx.RequireLogin`，管理员接口挂 `httpx.RequireAdmin`（实时查库校验）。身份通过 `httpx.User(ctx)` 读取（测试中用 `httpx.WithUser` 注入）。可选身份的接口需自行解析 `Authorization` 头（如 `GET /api/v1/app/{id}`）。
- **JWT 密钥**：`jwt.secret` 为空时每次启动随机生成，重启后已签发 token 全部失效——这是沿用的设计，不是 bug；需要持久登录时配置固定密钥。
- **验证码**：`internal/captcha` 用内存会话 + `JSESSIONID` Cookie 保存验证码（前端依赖 `withCredentials`），进程重启即清空。
- **数据库**：DSN 开启 `clientFoundRows`（UPDATE 返回匹配行数，与 JDBC 一致，业务依赖这一点判断是否命中）；`repo.New` 对 sqlx 启用 `Unsafe`，使 `select *` 忽略多余列。多步写操作用 `Store.inTx` 包在事务中（如创建/更新 API、删除 API/用户、统计聚合）。不要使用 MySQL 8 专有语法。
- **数据库迁移**：`internal/migrate` 是自写的 Flyway 兼容迁移器，脚本在 `internal/migrate/sql/`（`V1__`、`V1.1__`、`V1.2__` …，`embed` 打包），读写 `flyway_schema_history` 并按 Flyway 算法计算校验和。表名统一前缀 `helloapi_`。修改表结构时新增版本化脚本，不要改动已有脚本。
- **调用统计链路**：
  1. 外部被管理的 API 调用 `POST /api/v1/stat/log` 上报。`service/stat.go` 校验用户密钥与 API Key（时间型校验过期，次数型用 `count > 0` 条件原子扣减）后写入原始日志表 `helloapi_api_request_logs`。header 和 body 独立校验 JSON，非法值写 NULL，超限的合法值写 `{"truncated":true,"size":N}`，其中 N 是原始字节数。
  2. `internal/stat/scheduler.go` 定时调用 `service.Aggregation#AggregateNextBatch`，基于 `helloapi_stat_checkpoint` 的 `last_log_id` 每批最多 5000 条增量聚合到 `helloapi_api_daily_stats`（每日）和 `helloapi_api_views`（累计），三者在同一事务内更新，并用 `FOR UPDATE` 行锁防止多实例重复统计。
  3. `service/stat_dashboard.go` 从聚合表读取调用量，提供 `GET /api/v1/stat/dashboard`（管理后台）和 `GET /api/v1/stat/count`（公开累计次数），结果在进程内缓存 30 秒。看板的 `apiDailyCountArray` 以 MySQL 当前日期为准返回包含当天的 90 天日趋势，按日期升序并补零。`apiWeekCountArray` 取最后 7 天。
  4. `internal/stat/cleanup.go` 调度过期日志清理，按不超过已提交统计检查点的主键窗口分批删除。清理仅处理原始日志，截止日以 MySQL 当前日期为准。
  5. 删除 API 或用户时需同步清理相关聚合数据。
- **测试**：均为纯单元测试（fake 实现 repo 接口，不连数据库）。`scripts/smoke` 是需要真实 MySQL 与人工输入验证码的端到端冒烟脚本，用法见 `server/README.md`。

## 前端架构

- **请求层**：`src/utils/module/request.ts` 是 axios 实例，自动附带 `Bearer` token；响应拦截器按 `res.data.code` 处理——`200` 原样返回，`401` 清空用户状态并跳转登录页，其余 code 自动 `ElMessage.error(msg)` 后**仍然 resolve** 返回。因此调用方需自行判断 `res.data.code === 200`，且不必再重复弹错误提示。
- **接口与类型**：`src/api/module/*.ts` 按后端模块划分接口函数，`src/types/module/*.ts` 存放对应类型，均通过各自的 `index.ts` 聚合导出（`@/api`、`@/types`）。
- **路由**：`src/router/index.ts`，前台走 `IndexLayout`（`/`、`/search/:keywords`、`/info/:id`、`/links`），后台走 `AdminLayout`（`/admin/*`）。`beforeEach` 守卫负责登录检查，管理员专属页面（`/admin/manage`、`/admin/setting`、`/admin/check`、`/admin/link`）在守卫里硬编码判断——新增管理员页面时要同步更新这里。
- **状态**：Pinia + `pinia-plugin-persistedstate`，用户 token/信息在 `src/stores/modules/user.ts`。
- **自动导入**：Element Plus 组件与 API 由 `unplugin-auto-import` / `unplugin-vue-components` 自动导入（生成 `auto-imports.d.ts`、`components.d.ts`），无需手动 import `ElMessage` 等。路径别名 `@` → `web/src`。
- `web/public/api/api_count/index.php` 是独立的 PHP 脚本，请求后端 `/api/v1/stat/count` 生成调用次数 SVG 计数图，不属于 Vue 应用。

## 约定

- 代码注释与提交信息使用中文；提交信息风格为多行 `- 描述` 列表。
