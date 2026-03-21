<p align="right">
  <strong>中文</strong>
</p>

<div align="center">

# GinNextTemplate

Gin + Next.js 的后台管理模板工程，提供用户、邮箱、文件上传、安全、系统设置、应用日志和服务端升级等基础能力，适合作为新项目的起点。
</div>

## 项目定位

`GinNextTemplate` 是一个已经完成主线收口的全栈模板工程。后续默认工作方向是围绕模板基线继续开发、维护与扩展，而不是回到历史业务工程模式。

当前长期保留模块：

* 用户与认证
* 邮箱能力
* 文件上传
* 安全能力
* 系统设置
* 应用日志
* 服务端版本升级

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
export SQLITE_PATH='./ginnexttemplate.db'
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

## 开源协议

本项目采用 [Apache License 2.0](./LICENSE) 开源。
