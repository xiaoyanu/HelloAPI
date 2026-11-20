# 统计趋势多时间范围 + 调用日志瘦身 实施计划

> 规划：Claude；实施：GPT；验收：Claude（使用线上备份还原的 MySQL 5.7 副本实测）。
> 所有改动必须兼容 MySQL 5.7，不得使用 MySQL 8 专有语法；接口契约保持向后兼容。

## 背景数据（2026-11-20 线上备份）

- `helloapi_api_request_logs`：约 2519 万行，日增约 18–22 万行，数据最早为 2026-03-05。
- 表数据约 1.5 GB，**二级索引约 2.2 GB**（比数据还大）。
- `header` + `body` 两个 JSON 字段约占行数据的 37%（header 中甚至包含完整 Cookie）。
- 原始日志表除聚合任务（按 `log_id` 区间读取）和删除 API/用户（按 `api_id` 删除）外，**没有任何读取方**；看板只读 `helloapi_api_daily_stats` / `helloapi_api_views`。
- `helloapi_api_daily_stats` 同样从 2026-03-05 开始，90 天趋势可直接从聚合表读取。
- 保留 90 天时，首次清理约删除 908 万行。

## 已确认的决策

| 项目 | 决策 |
| --- | --- |
| 趋势图范围 | 7 / 15 / 30 / 90 天，前端切换 |
| 原始日志保留 | 默认 90 天，可配置，`0` 表示永久保留（不清理） |
| header/body | 默认每个字段最多 2048 字节，超限时替换为占位 JSON；可配置 |
| 索引 | 删除无用二级索引，仅保留主键与 `idx_api_id` |

---

## 一、统计趋势多时间范围

### 1.1 后端

**设计**：`GET /api/v1/stat/dashboard` 新增字段 `apiDailyCountArray`，固定返回近 90 天（含今天）的每日调用量，按日期升序，缺失日期补 0，共 90 个元素。前端在本地截取最后 N 个。原有字段 `apiWeekCountArray` 保留（`AdminPage.vue` 仍在使用），改为由 90 天数组的最后 7 个元素派生，语义与格式不变。

- `internal/model/stat.go`
  - `StatDashboard` 新增 `APIDailyCountArray []APIWeekArray \`json:"apiDailyCountArray"\``（复用 `APIWeekArray`，`date` 仍为 `model.LocalDateTime`，输出 `2026-11-20T00:00:00`，与现有周数组一致）。
- `internal/repo/stat.go`
  - 删除 `GetAPIWeekCountArrayFromDaily`。
  - 新增 `GetAPIDailyCountArray(ctx, days int) ([]model.APIWeekArray, error)`：
    1. 先取数据库日期锚点 `SELECT CURDATE()`（避免 Go 进程与 MySQL 时区不一致）；
    2. 再查 `SELECT stat_date, SUM(call_count) FROM helloapi_api_daily_stats WHERE stat_date >= ? AND stat_date <= ? GROUP BY stat_date`，参数为 `锚点 - (days-1)` 与锚点；
    3. 补零与排序可以放在 repo，也可以放在 service，但**补零逻辑必须有单元测试覆盖**（建议抽成纯函数，例如 `fillDailyCounts(today time.Time, days int, rows) []model.APIWeekArray`，放在 service 层以便测试）。若放 service，则 repo 拆成两个方法（取锚点、取区间行），或一个方法同时返回锚点与行。
  - 注意 MySQL 5.7 没有递归 CTE，不要用 SQL 生成日期序列。
- `internal/service/stat_dashboard.go`
  - `DashboardRepository` 接口去掉周数组方法，换成新方法。
  - 常量 `dailyTrendDays = 90`。
  - `APIWeekCountArray` = 90 天数组最后 7 个元素（需要拷贝成独立切片，不要与 daily 共享底层数组导致后续误改）。
  - 两个数组在任何情况下都不能为 nil（JSON 必须输出 `[]` 或完整数组）。
- 测试 `internal/service/stat_dashboard_test.go`：更新 fake；新增用例覆盖：
  - 补零后长度恰为 90、首元素日期为 `today-89`、末元素为 `today`、升序；
  - 数据库返回的稀疏行被放到正确位置；
  - `apiWeekCountArray` 等于 daily 的最后 7 个；
  - 跨月、跨年日期正确（例如锚点为 2027-01-03）。
- `server/README.md` 的接口说明中补充 `apiDailyCountArray` 字段。

### 1.2 前端

- `web/src/types/module/stat.ts`：`StatDashboard` 新增 `apiDailyCountArray: ApiWeekCountItem[]`（可新增类型别名 `ApiDailyCountItem = ApiWeekCountItem`）。
- `web/src/views/admin/StatPage.vue`：
  - 新增 `trendDays = ref<7 | 15 | 30 | 90>(7)`，选项常量 `[7, 15, 30, 90]`。
  - 保存完整的 90 天数据，用 `computed` 截取最后 `trendDays` 个元素生成 x 轴（`date.slice(5, 10)`，即 `MM-DD`）和数值。
  - 图表卡片标题改为 `近{{ trendDays }}天调用统计`，标题右侧放一个范围切换控件，优先使用 Element Plus 的 `el-segmented`（若当前 Element Plus 版本不支持，则用 `el-radio-group` + `el-radio-button`，`size="small"`）。标题行用 flex 两端对齐，小屏幕下不能溢出或换行错乱。
  - 长范围下的可读性：
    - 范围 `> 15` 天时关闭每个点上方的数值 `label`（tooltip 仍可查看）；
    - 范围 `= 90` 天时 `showSymbol: false`；
    - x 轴标签交给 ECharts 自动间隔（不要设 `interval: 0`）。
  - 切换范围只改本地截取，不发请求；5 秒自动刷新保持不变，刷新时不能重置用户选中的范围。
  - 保持现有拟态卡片风格，不引入新依赖。
- `AdminPage.vue` 不改。
- `pnpm build` 必须通过（含 `vue-tsc` 类型检查）。

---

## 二、调用日志瘦身

### 2.1 迁移：删除无用索引

新增 `server/internal/migrate/sql/V1.3__request_log_slimming.sql`（**不得修改已有迁移脚本**）：

```sql
-- 原始调用日志只按主键区间聚合/清理、按 api_id 删除，其余二级索引均无读取方。
ALTER TABLE `helloapi_api_request_logs`
    DROP INDEX `idx_time`,
    DROP INDEX `idx_user_id`,
    DROP INDEX `idx_api_key`,
    DROP INDEX `idx_time_api_id`;
```

- 在 MySQL 5.7 上删除二级索引是 in-place 的元数据操作，大表上也很快。
- 保留 `PRIMARY` 与 `idx_api_id`（`DeleteAPILog` 按 `api_id` 删除需要）。
- 确认迁移器能正确执行（版本号 `1.3` 排在 `1.2` 之后，校验和按现有算法计算），必要时在 `migrate_test.go` 中补充对新脚本的解析/排序断言。

### 2.2 过期日志定时清理

**配置**（`internal/config/config.go`、`config.example.yaml`、`server/README.md` 配置表同步更新）：

| YAML | 默认值 | 环境变量 | 说明 |
| --- | --- | --- | --- |
| `stat.log-retention-days` | `90` | `HELLOAPI_STAT_LOG_RETENTION_DAYS` | 原始日志保留天数，`0` 表示不清理；负数报错 |
| `stat.log-cleanup-interval-ms` | `3600000` | `HELLOAPI_STAT_LOG_CLEANUP_INTERVAL_MS` | 两轮清理之间的间隔，必须为正 |

**清理算法**（每批独立自动提交，不包大事务）：

1. 读取检查点 `SELECT last_log_id FROM helloapi_stat_checkpoint WHERE job_name = 'api-daily-stats'`（普通读，**不加** `FOR UPDATE`；检查点只增不减，读到旧值也是安全的）。若检查点不存在则本轮直接结束。
2. 按主键取本批窗口上界：
   `SELECT COALESCE(MAX(log_id), 0) FROM (SELECT log_id FROM helloapi_api_request_logs WHERE log_id <= ? ORDER BY log_id LIMIT ?) AS w`
   参数为检查点与批大小（5000）。上界为 0 表示没有已聚合日志，结束本轮。
3. 删除窗口内过期日志：
   `DELETE FROM helloapi_api_request_logs WHERE log_id <= ? AND time < CURDATE() - INTERVAL ? DAY`
   参数为窗口上界与保留天数；截止日期由数据库计算（`time` 列由数据库 `CURRENT_TIMESTAMP` 写入，避免时区差异）。取受影响行数（`repo.affected`）。
4. 删除行数为 0 → 本轮结束（窗口头部已全是未过期日志）；否则暂停 100ms 后继续下一批，直到为 0 或 ctx 取消。

要点：
- **绝不能删除 `log_id > 检查点` 的日志**（尚未聚合），由第 2 步窗口上界保证。
- 不依赖 `time` 索引（2.1 中已删除），窗口与删除都走主键范围。
- 多实例同时清理是安全的（重复删除只会得到 0 行）。

**代码组织**（参考现有聚合的分层与写法）：
- `internal/repo/stat_cleanup.go`：`FindCheckpointNoLock`（或类似命名）、`FindPurgeUpperLogID(ctx, checkpoint int64, batchSize int)`、`DeleteExpiredLogs(ctx, upper int64, retentionDays int) (int64, error)`。
- `internal/service/stat_cleanup.go`：`LogCleanup` 结构体，依赖 repo 接口，提供 `PurgeExpiredBatch(ctx) (int64, error)`（执行上述 1–3 步一次）。
- `internal/stat/cleanup.go`：`RunCleanup(ctx, purger, initialDelay, interval, pause, logger)`，循环调用 `PurgeExpiredBatch` 直到返回 0 或出错，每轮结束记录一条 Info 日志（本轮删除总行数、耗时），出错记录 Error 并等待下一轮；ctx 取消时立即退出（包括暂停期间）。首次延迟可固定为 1 分钟（避免与启动迁移、聚合追赶争抢资源）。
- `cmd/helloapi/main.go`：`log-retention-days > 0` 时才启动清理 goroutine，纳入现有 `workers` 等待组，优雅退出。
- 单元测试（fake 实现，不连数据库）：
  - service：检查点不存在 / 上界为 0 时不调用删除；正确传递检查点、批大小、保留天数；返回删除行数；错误透传。
  - stat：多批循环直到 0 行停止；出错后本轮停止；ctx 取消能及时退出（可把 pause/interval 设得很短）。
  - config：默认值、YAML、环境变量覆盖、非法值（负的保留天数、非正间隔）报错。

### 2.3 header/body 体积限制

**配置**：

| YAML | 默认值 | 环境变量 | 说明 |
| --- | --- | --- | --- |
| `stat.log-payload-max-bytes` | `2048` | `HELLOAPI_STAT_LOG_PAYLOAD_MAX_BYTES` | 单个字段最大字节数；`0` 表示不保存（写 NULL）；`-1` 表示不限制；小于 `-1` 报错 |

**行为**（`internal/service/stat.go`）：
- `NewStat(store, maxPayloadBytes int)`，`main.go` 传入配置值。
- 保持现有逻辑：值不是合法 JSON 时写 NULL。
- 合法 JSON 且长度（字节）超过上限时，不做截断（截断会破坏 JSON 合法性），改为写入占位 JSON：`{"truncated":true,"size":<原始字节数>}`。
- 上限为 `0` 时 header/body 一律写 NULL；`-1` 时原样保存。
- 单元测试覆盖：未超限原样、超限占位（且占位是合法 JSON，size 正确）、0、-1、非法 JSON 仍为 NULL。

### 2.4 磁盘空间回收（文档，不自动执行）

InnoDB 的 DELETE / DROP INDEX 释放的页只在表空间内复用，`.ibd` 文件不会变小。在 `server/README.md` 新增「调用日志瘦身与空间回收」小节，说明：
- 升级后首次清理会分批删除超出保留期的历史日志（以线上量级约 900 万行，耗时为分钟级），期间服务正常可用。
- 首次清理完成后，在低峰期手动执行 `OPTIMIZE TABLE helloapi_api_request_logs;`（5.7 下为在线重建，需要约等于剩余表大小的空闲磁盘；需开启 `innodb_file_per_table`）。
- 实测耗时由验收阶段补充。

---

## 三、文档与约定

- `CLAUDE.md`：在「配置」和「调用统计链路」中补充日志保留清理、payload 上限、`apiDailyCountArray`。
- `server/README.md`：配置表、接口字段、空间回收小节。
- 代码注释使用中文；不新增第三方依赖。

## 四、实施自检（GPT 完成后必须全部通过）

```bash
cd server && gofmt -l . && go vet ./... && go test ./...
cd web && pnpm build
```

- 可使用 `helloapi-mysql57-test` 容器（127.0.0.1:3307，root / `helloapi_test_only`，库 `helloapi_go_test`；若已停止可 `docker start`）做小规模联调。
- **禁止操作 `helloapi-mysql57-prod` 容器**（线上备份副本，由验收方使用）。
- 不要提交 git，由验收方审查后提交。

## 五、验收清单（Claude）

1. 代码审查：检查点约束、批量删除走主键、事务边界、时区、nil 切片、配置校验、前端刷新不重置范围。
2. 独立复跑 `gofmt` / `go vet` / `go test` / `pnpm build`。
3. 从 prod 副本克隆一份独立容器，实测：
   - V1.3 迁移耗时与索引结果；
   - 清理任务首次运行的删除量、耗时，确认 `log_id > 检查点` 的日志未被删除，聚合表数据不变；
   - dashboard 接口返回 90 个点且与聚合表一致；
   - `OPTIMIZE TABLE` 耗时与回收后的表大小，回填到 README。
4. 前端页面实际查看 7/15/30/90 切换效果。
