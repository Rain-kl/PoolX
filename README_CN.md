<p align="right">
  <a href="./README.md">English</a>
</p>

<div align="center">

# PoolX

基于 Gin 和 Next.js 的代理池控制面，核心目标是把 Clash/Mihomo 节点组织成可复用的代理池，服务爬虫、抓取、代理请求和自动化出网场景。
</div>

## 项目用途

PoolX 面向这类需求而设计：

* 导入和管理大量 Clash 兼容节点
* 将节点组织成可复用的代理池
* 为爬虫、自动化任务和代理请求系统暴露稳定的本地代理入口
* 对节点应用负载均衡、自动回退和基于延迟的选择策略
* 通过可视化控制面管理 Mihomo 运行内核

可以把它理解为：将原始订阅和零散节点配置，转成可运营、可复用、可编排的代理池基础设施。

## 功能特性

* Web 管理端，内置登录鉴权、系统设置、日志与文件上传能力
* Clash 兼容配置导入链路
* 节点池管理，支持节点测试、筛选和跨工作台复用
* 基于工作台的端口配置管理，用于构建代理池入口
* 支持 `Mixed` 和 `SOCKS/HTTP` 两种监听模式
* 运行阶段自动聚合多个工作台配置，生成最终 Mihomo 配置
* 内置 `select`、`url-test`、`fallback`、`load-balance` 等策略能力
* 基于 JSON 的代理设置扩展，可承载监听鉴权、UDP、测速参数和负载均衡调优
* 设置页支持 Mihomo 二进制安装与管理
* 内置 zashboard，并通过 PoolX 鉴权同源反代到 `/zashboard/`，不向浏览器暴露 Clash secret

## 典型使用场景

* 爬虫和抓取服务的代理池
* 自动化平台的统一出网代理网关
* 数据采集任务的轮换代理或回退代理入口
* 面向不同目标站点的工作台式代理编排
* 团队内部使用的 Mihomo 代理基础设施控制台


## 快速开始

### Docker 部署

```yaml
services:
  postgres:
    image: postgres:17-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: poolx
      POSTGRES_USER: poolx
      POSTGRES_PASSWORD: replace-with-strong-password
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U poolx -d poolx"]
      interval: 10s
      timeout: 5s
      retries: 5

  poolx:
    image: ghcr.io/rain-kl/poolx:latest
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "3000:3000"
#     开放代理监听端口
    environment:
      SESSION_SECRET: replace-with-random-string
      SQLITE_PATH: /data/poolx.db
      DSN: postgres://poolx:replace-with-strong-password@postgres:5432/poolx?sslmode=disable
      GIN_MODE: release
      LOG_LEVEL: info

    volumes:
      - ./data/poolx:/data
```

### 环境要求

* Go 1.24+
* Node.js 20+
* pnpm

### 构建前端静态资源

```bash
cd server/web
corepack enable
pnpm install
pnpm build
```

```bash
cd server/zashboard
corepack enable
pnpm install
pnpm build
```

### 启动服务端

```bash
cd server
export SESSION_SECRET='replace-with-random-string'
export SQLITE_PATH='./poolx.db'
export LOG_LEVEL='info'
go run ./cmd/server
```

访问地址：`http://localhost:3000`

内置 Clash 控制台：`http://localhost:3000/zashboard/`

默认账号：

* 用户名：`root`
* 密码：`123456`

## 本地开发

Server：

```bash
cd server
go run ./cmd/server
```

Frontend：

```bash
cd server/web
pnpm install
pnpm dev
```

Zashboard：

```bash
cd server/zashboard
pnpm install
pnpm build
```

服务端测试：

```bash
cd server
go test ./...
```

前端构建：

```bash
cd server/web
pnpm build
```

## 配置说明

PoolX 不要求你手工维护一个固定的 Mihomo 主配置文件。

它的核心流程是：

* 导入节点
* 组织节点池
* 定义工作台监听配置
* 自动渲染最终 Mihomo 运行配置

运行参数、部署方式和系统配置请参考：

* [docs/app-config.md](./docs/app-config.md)
* [docs/deployment.md](./docs/deployment.md)

## License

本项目采用 [Apache License 2.0](./LICENSE) 开源。
