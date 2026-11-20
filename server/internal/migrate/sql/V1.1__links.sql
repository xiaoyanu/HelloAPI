CREATE TABLE `helloapi_links`
(
    `link_id` int          NOT NULL AUTO_INCREMENT COMMENT '友链唯一ID',
    `name`    varchar(255) NOT NULL COMMENT '名称',
    `url`     varchar(255) NOT NULL COMMENT '链接',
    `desc`    varchar(255) NULL COMMENT '网站描述',
    `avatar`  varchar(255) NULL COMMENT '图标',
    `created` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated` datetime     NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`link_id`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4 COMMENT ='友情链接';