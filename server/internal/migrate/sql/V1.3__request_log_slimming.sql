-- 原始调用日志只按主键区间聚合/清理、按 api_id 删除，其余二级索引均无读取方。
ALTER TABLE `helloapi_api_request_logs`
    DROP INDEX `idx_time`,
    DROP INDEX `idx_user_id`,
    DROP INDEX `idx_api_key`,
    DROP INDEX `idx_time_api_id`;
