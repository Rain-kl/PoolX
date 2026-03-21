<p align="right">
  <strong>中文</strong>
</p>

<div align="center">

# PoolX

基于 Gin + Next.js 的 Proxy Kernel Control Plane，首期围绕 Mihomo 交付代理池控制端能力。
</div>

## 项目定位

当前项目定位为面向代理池场景的后台管理系统。模板基础设施能力仍然保留，但主线目标已经切换为业务产品开发。

当前主线能力：

* 用户与认证
* 邮箱能力
* 文件上传
* 安全能力
* 系统设置
* 应用日志
* 服务端版本升级
* 配置导入
* 节点池与节点测试
* 工作台配置
* 运行控制与状态查看

## 当前工程基线

* 服务端工作目录为 `server/`
* 启动入口为 `server/cmd/server`
* 前端工程位于 `server/web`
* 服务端当前按 `internal/app`、`internal/handler`、`internal/service`、`internal/model`、`internal/middleware`、`internal/router`、`internal/pkg` 组织

## 快速开始

### 1. 构建前端

```bash
cd server/web
corepack enable
pnpm install
pnpm build
```

### 2. 启动服务端

```bash
cd server
export SESSION_SECRET='replace-with-random-string'
export SQLITE_PATH='./poolx.db'
export LOG_LEVEL='info'
go run ./cmd/server
```

访问地址：`http://localhost:3000`

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

## 常用命令

Server 测试：

```bash
cd server
go test ./...
```

Frontend 构建：

```bash
cd server/web
pnpm build
```

## 文档

建议按以下顺序阅读：

1. [docs/design.md](./docs/design.md)
2. [docs/development-guidelines.md](./docs/development-guidelines.md)
3. [docs/development-plan.md](./docs/development-plan.md)
4. [docs/frontend-development-guidelines.md](./docs/frontend-development-guidelines.md)
5. [docs/deployment.md](./docs/deployment.md)
6. [docs/app-config.md](./docs/app-config.md)
7. [docs/dev/Product.md](./docs/dev/Product.md)
8. [docs/dev/Demand.md](./docs/dev/Demand.md)
9. [docs/dev/Tech.md](./docs/dev/Tech.md)

## 开源协议

本项目采用 [Apache License 2.0](./LICENSE) 开源。
