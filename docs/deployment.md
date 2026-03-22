# PoolX 部署说明

本文档描述 `PoolX` 当前阶段的部署基线、启动方式和最小验证步骤。

当前文档只描述 PoolX 服务端与前端的部署，不包含 Agent、节点联调或 OpenResty 配置分发链路。

## 1. 前置条件

### 1.1 Server

* Go 1.24+
* Node.js 18+
* pnpm
* 可写 SQLite 文件目录，或可访问的 PostgreSQL 实例

## 2. 启动方式

### 2.1 构建前端静态产物

```bash
cd server/web
corepack enable
pnpm install
pnpm build
```

`pnpm build` 会生成供 Go Server 托管的管理端静态产物。

如需内置 Clash 控制台 `zashboard`，还需要构建一次它的静态资源：

```bash
cd server/zashboard
corepack enable
pnpm install
pnpm build
```

构建完成后，Go Server 会托管：

* 管理端：`/`
* Clash 控制台：`/zashboard/`

### 2.2 源码启动

推荐使用 `cmd/server` 入口启动：

```bash
cd server
export SESSION_SECRET='replace-with-random-string'
export SQLITE_PATH='./poolx.db'
export LOG_LEVEL='info'
# 可选：使用 PostgreSQL
# export DSN='postgres://template:secret@127.0.0.1:5432/poolx?sslmode=disable'
go run ./cmd/server
```

如需使用嵌入式静态资源方式，也可直接执行：

```bash
cd server
go run .
```

默认监听 `3000` 端口。

如果使用 `go run .` 的嵌入式静态资源入口，启动前需要先完成 `server/web` 和 `server/zashboard` 的构建。

### 2.3 Docker Compose 启动

可参考 `server/docker-compose.yaml`：

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
      - postgres-data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U poolx -d poolx"]
      interval: 10s
      timeout: 5s
      retries: 5

  poolx:
    image: ghcr.io/rain-kl/poolx:latest
    container_name: poolx
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "3000:3000"
    environment:
      SESSION_SECRET: replace-with-random-string
      SQLITE_PATH: /data/poolx.db
      DSN: postgres://poolx:replace-with-strong-password@postgres:5432/poolx?sslmode=disable
      GIN_MODE: release
      LOG_LEVEL: info
    volumes:
      - poolx-data:/data

volumes:
  postgres-data:
  poolx-data:
```

启动：

```bash
cd server
docker compose up -d
```

## 3. 首次登录

访问 `http://localhost:3000`

默认账号：

* 用户名：`root`
* 密码：`123456`

首次部署后应尽快修改默认密码，并配置正式的 `SESSION_SECRET`。

如需打开内置 Clash 控制台，可在运行状态页点击“打开 Clash 控制台”，或直接访问：

`http://localhost:3000/zashboard/`

该界面会通过 PoolX 服务端同源反代连接 Mihomo external-controller，继续复用管理端登录态，不需要在浏览器中填写 `secret` 或直连本机 Clash API。

## 4. Swagger

登录管理端后访问：

`http://localhost:3000/swagger/index.html`

如需重新生成文档：

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.4
cd server
swag init -g main.go -o docs
```

## 5. 本地开发配合方式

前端本地开发：

```bash
cd server/web
pnpm dev
```

`pnpm dev` 默认启动 `http://127.0.0.1:3001`，并通过 `NEXT_DEV_BACKEND_URL` 同源代理到服务端。

## 6. 最小验证步骤

1. 构建前端静态产物
2. 启动服务端
3. 登录管理端
4. 验证用户登录、文件上传、日志、系统设置页面可访问
5. 如启用邮件能力，验证邮箱验证码或密码重置流程
6. 如启用升级能力，验证升级相关页面与接口可访问

## 7. 升级说明

当前 PoolX 保留服务端升级能力：

* 检查最新版本
* 上传服务端二进制进行手动升级
* 执行服务端升级流程

默认升级仓库为 `Rain-kl/PoolX`，可在系统设置中通过 `ServerUpdateRepo` 改为自己的 GitHub 发布仓库，格式为 `owner/repo`。

当前系统设置还支持配置代理内核：

* 当前仅开放 `mihomo`
* 可填写 Mihomo 二进制目标路径
* 可通过手动上传或从官方仓库自动下载适配当前平台的发行版完成安装
* 安装完成后会立即执行版本校验，并写回运行时配置

当前文档不描述 Agent 升级、节点升级或 OpenResty 重载。

## 8. 常用验证命令

### 8.1 Server

```bash
cd server
GOCACHE=/tmp/poolx-go-cache go test ./...
```

### 8.2 Frontend

```bash
cd server/web
pnpm build
```

## 9. 文档维护要求

启动方式、部署方式、升级流程或联调步骤变化时，必须同步更新本文档和 `README.md`。
