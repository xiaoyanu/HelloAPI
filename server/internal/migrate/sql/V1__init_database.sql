-- =============================================================================
-- Flyway 初始版本脚本: V1__Initial_Setup.sql
-- 描述: 初始化 API 管理系统基础表结构
-- =============================================================================

-- 1. 用户基础信息表
CREATE TABLE `helloapi_users`
(
    `id`       int(10)      NOT NULL AUTO_INCREMENT COMMENT '用户唯一ID',
    `username` varchar(255) NOT NULL COMMENT '用户名',
    `password` varchar(255) NOT NULL COMMENT '密码',
    `nick`     varchar(255) NOT NULL COMMENT '用户昵称',
    `mode`     int(2)       NOT NULL COMMENT '用户类型0为普通，1为管理员',
    `mail`     varchar(255) NOT NULL COMMENT '邮箱',
    `created`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated`  datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='用户基础信息表';

-- 2. 接口主表
CREATE TABLE `helloapi_api_apps`
(
    `id`            int(10)      NOT NULL AUTO_INCREMENT COMMENT '唯一ID',
    `title`         varchar(255) NOT NULL COMMENT '接口名',
    `smallTitle`    varchar(255) NOT NULL COMMENT '接口描述',
    `status`        int(2)       NOT NULL COMMENT '状态，正常0，异常1，维护2，隐藏3',
    `type`          int(2)       NOT NULL COMMENT '接口类型，免费0，付费1',
    `url`           varchar(255) NOT NULL COMMENT '请求地址',
    `sendType`      int(2)       NOT NULL COMMENT '请求类型，0 GET、1 POST、2 PUT、3 DELETE',
    `returnType`    varchar(255) NOT NULL COMMENT '返回类型',
    `returnContent` longtext     NOT NULL COMMENT '返回内容',
    `created`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated`       datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `user_id`       int(10)      NOT NULL COMMENT '这个接口的用户',
    `view_status`   int(2)       NOT NULL COMMENT '展示状态，0正常，1审核中，2拒绝',
    PRIMARY KEY (`id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='接口主表';

-- 3. API 授权密钥表
CREATE TABLE `helloapi_api_keys`
(
    `api_id`  int(10)      NOT NULL COMMENT '对应的API的ID',
    `key`     varchar(255) NOT NULL COMMENT 'api密钥',
    `created` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    `type`    int(2)       NOT NULL COMMENT '计费类型：0时间，1次数',
    `started` datetime              DEFAULT NULL COMMENT '开始时间',
    `expired` datetime              DEFAULT NULL COMMENT '过期时间',
    `count`   bigint(13)   NOT NULL COMMENT '可用次数',
    `desc`    varchar(255)          DEFAULT NULL COMMENT '备注',
    PRIMARY KEY (`key`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='API授权密钥表';

-- 4. 接口请求参数表
CREATE TABLE `helloapi_api_params`
(
    `api_id`   int(10)      NOT NULL COMMENT '对应的Api唯一ID',
    `name`     varchar(255) NOT NULL COMMENT '参数名',
    `required` int(2)       NOT NULL COMMENT '是否必填，0否，1是',
    `type`     varchar(255) NOT NULL COMMENT '参数类型',
    `msg`      varchar(255) NOT NULL COMMENT '描述'
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='接口请求参数表';

-- 5. API 调用日志表
CREATE TABLE `helloapi_api_request_logs`
(
    `log_id`  bigint(20) NOT NULL AUTO_INCREMENT COMMENT '主键',
    `api_id`  int(10)    NOT NULL COMMENT '对应的API ID',
    `ip`      varchar(255)        DEFAULT NULL COMMENT '请求 IP',
    `time`    datetime   NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '请求时间',
    `header`  json                DEFAULT NULL COMMENT '请求头',
    `body`    json                DEFAULT NULL COMMENT '请求参数',
    `api_key` varchar(255)        DEFAULT NULL COMMENT '请求APIKey',
    `user_id` int(10)    NOT NULL COMMENT '用户ID',
    PRIMARY KEY (`log_id`),
    INDEX `idx_api_id` (`api_id`),
    INDEX `idx_time` (`time`),
    INDEX `idx_user_id` (`user_id`),
    INDEX `idx_api_key` (`api_key`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='API调用日志表';

-- 6. 接口调用次数统计
CREATE TABLE `helloapi_api_views`
(
    `api_id` int(10)    NOT NULL,
    `count`  bigint(13) NOT NULL,
    PRIMARY KEY (`api_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='接口调用次数统计';

-- 7. 系统全局设置
CREATE TABLE `helloapi_settings`
(
    `key`     varchar(255) NOT NULL,
    `value`   varchar(255) NOT NULL,
    `updated` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`key`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='系统全局设置';

-- 8. 用户身份凭证表
CREATE TABLE `helloapi_user_keys`
(
    `user_id` int(10)      NOT NULL COMMENT '对应的用户id',
    `key`     varchar(255) NOT NULL COMMENT '用户接口key',
    `created` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`user_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='用户身份凭证表';