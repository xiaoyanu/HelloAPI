-- 每个 API 每天只保留一条聚合记录，Dashboard 不再扫描原始日志表。
CREATE TABLE `helloapi_api_daily_stats`
(
    `stat_date`  date       NOT NULL COMMENT '统计日期',
    `api_id`     int(10)    NOT NULL COMMENT 'API ID',
    `call_count` bigint(20) NOT NULL DEFAULT 0 COMMENT '调用次数',
    PRIMARY KEY (`stat_date`, `api_id`),
    INDEX `idx_daily_api_date` (`api_id`, `stat_date`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='API每日调用统计';

-- 保存增量任务最后成功处理的日志 ID，应用重启后可从断点继续。
CREATE TABLE `helloapi_stat_checkpoint`
(
    `job_name`    varchar(64) NOT NULL COMMENT '聚合任务名称',
    `last_log_id` bigint(20)  NOT NULL DEFAULT 0 COMMENT '最后处理的日志ID',
    `updated_at`  datetime    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`job_name`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='统计聚合检查点';

-- 首次升级时，将现有原始日志回填为每日统计数据。
INSERT INTO `helloapi_api_daily_stats` (`stat_date`, `api_id`, `call_count`)
SELECT DATE(`time`), `api_id`, COUNT(*)
FROM `helloapi_api_request_logs`
GROUP BY DATE(`time`), `api_id`;

-- 以原始日志回填结果为准，重新校准历史总调用次数。
DELETE FROM `helloapi_api_views`;

INSERT INTO `helloapi_api_views` (`api_id`, `count`)
SELECT `api_id`, SUM(`call_count`)
FROM `helloapi_api_daily_stats`
GROUP BY `api_id`;

-- 将检查点初始化到迁移时的最大日志 ID；之后只处理新产生的日志。
INSERT INTO `helloapi_stat_checkpoint` (`job_name`, `last_log_id`)
SELECT 'api-daily-stats', COALESCE(MAX(`log_id`), 0)
FROM `helloapi_api_request_logs`;

-- 覆盖按时间范围、API 分组的查询，并保留原 idx_time 以便平滑上线。
ALTER TABLE `helloapi_api_request_logs`
    ADD INDEX `idx_time_api_id` (`time`, `api_id`);

-- 优化最近 30 天用户注册数查询。
ALTER TABLE `helloapi_users`
    ADD INDEX `idx_users_created` (`created`);

-- 优化最近 30 天 API 发布数查询。
ALTER TABLE `helloapi_api_apps`
    ADD INDEX `idx_apps_created` (`created`);
