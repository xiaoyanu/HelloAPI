<h1 align="center">

  <img height="100" src="./web/src/assets/images/logo.png" />

HelloAPI

[![GitHub release](https://img.shields.io/github/v/release/xiaoyanu/helloapi.svg)](../../releases/latest)

</h1>

## 简介

采用拟态风格，简约的API管理程序

- 支持正常、维护、未公开、收费、隐藏等状态
- 自适应移动端、PC端
- 支持用户注册，可自定义全局设置
- 支持审核API接口
- 数据统计

## 技术栈

- Go 1.27（chi + sqlx）
- Vue 3 + Vite

## 系统要求

- Go 1.27+（仅自行编译时需要）
- MySQL 5.7+

## 截图

![Home](./assets/1.jpg)

![Admin](./assets/2.jpg)

## Demo演示站

https://api.zxz.ee

## 部署

### 后端

- 目录：`server`
- 配置文件：将 `server/config.example.yaml` 复制为 `config.yaml`，放在可执行文件同级目录后修改数据库等配置（也可用 `-config` 参数指定路径，或用 `HELLOAPI_DB_PASSWORD` 等环境变量覆盖）
- 编译并运行
  ```
  cd server
  go build -o helloapi ./cmd/helloapi
  ./helloapi
  ```
- 首次启动会自动建表；已在使用旧版（Java）的数据库可直接接管，无需迁移数据，切换后用户需重新登录
- 详细配置项与接口说明见 [server/README.md](./server/README.md)
- 不想自己编译？
    - 直接前往 [releases](../../releases/latest) 下载编译好的

### 前端

- 目录 `web`
  ```
  cd web
  ```
- 修改配置文件 `web/src/config/config.ts`
- 下载依赖 `pnpm i`
- 直接运行 `pnpm dev`
- 编译/打包 `pnpm build`
- 伪静态 Nginx
  ```
    location / {
    try_files $uri $uri/ /index.html;
    }
    ```

## 部署视频教程

哔哩哔哩：
https://www.bilibili.com/video/BV1KqNgziE5p
