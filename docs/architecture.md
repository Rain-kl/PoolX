# 架构总览

## 项目定位

PoolX 是**代理池控制面**：

- 后端提供管理员登录、系统健康/版本、可热更新的系统设置，以及节点管理、端口配置、内核生命周期控制
- 前端是 React 管理端 SPA，覆盖节点导入、代理池组织、运行时控制、内核日志等场景
- 扩展方式：复制 `example` 切片，改名并接线到组合根与路由

## 仓库布局

```text
PoolX/
  AGENTS.md                 # Agent / 开发入口（router）
  config.example.yaml       # 配置模板
  Makefile                  # 本地启动等
  backend/
    cmd/foam/               # 进程入口
    internal/
      app/                  # 组合根：DI + 生命周期
      cli/                  # flags / --config
      domain/               # 实体（admin, clash, settings, example…）
      application/          # 用例服务
      repository/           # 持久化接口
      infra/                # config, security, persistence, observability
      transport/http/       # Gin 路由、middleware、handlers
      shared/response/      # 统一 API envelope
  frontend/
    src/
      app/                  # providers, shell, router, auth boundary
      features/auth/        # 登录
      features/clash/       # 节点、端口配置、运行时控制
      features/settings/    # 系统设置
      components/ui/        # shadcn 原子组件
      shared/               # api, auth, hooks, lib, components
  docs/                     # 本目录：按需文档
```

## 运行时请求流

```text
CLI (cmd/foam → cli.Run)
  → config.Load
  → app.New
       → DB open + schema
       → bootstrap admin
       → wire adminauth + clash + settings + example services
       → httpserver.New (Gin)
       → AutoStartKernel (goroutine, if config available)
  → http.Server.Listen
```

### HTTP 分层

```text
Client
  → middleware (request ID, security headers, body limit, timeout, access log)
  → route group
       公开：/healthz, /readyz, /api/system/*, /api/admin/v1/auth/*
       鉴权：Admin JWT / session middleware
       handler (transport/http/<module>)
         → application service
           → repository interface
             → infra/persistence/relational
  → 未匹配 API 之外：frontend static (SPA)

API 约定（对齐 Wavelet）：
- 前缀：`/api` + `/api/v1/...`；管理端 `/api/v1/admin/...`；用户鉴权 `/api/v1/user/...`
- 探活：`GET /api/health`（envelope）与 `/healthz`/`/readyz`（探针别名）
- 响应体：`{ "error_msg": "", "data": T|null }`；错误用 HTTP 状态 + `error_msg`
```

### 依赖方向

```text
transport  →  application  →  domain
                 ↓
            repository (interfaces)
                 ↑
         infra/persistence (implements)
```

- `domain`：纯类型与规则，禁止 import Gin / GORM / 外部 IO
- `application`：编排用例、校验输入、映射错误；不绑 HTTP、不写 SQL
- `repository`：接口 + 分页等共享查询类型
- `infra`：配置、JWT/密码、GORM 实现
- `transport`：绑定 JSON、状态码、`shared/response` envelope
- `app`：唯一组合根，组装以上依赖

## 可运行 API 面

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| GET | `/api/health`、`/healthz` | 存活（envelope / 探针别名） |
| GET | `/readyz` | 就绪快照 |
| POST | `/api/v1/user/login` · `/refresh` | 管理员登录 / 刷新 |
| GET | `/api/v1/user/logout` · `/self` | 登出 / 当前用户 |
| GET | `/api/v1/config/public` | 公开运行时配置 |
| * | `/api/v1/admin/...` | 鉴权后的 admin（status / settings / examples） |
| CRUD | `/api/v1/admin/examples` | 示例资源（需 admin JWT） |
| * | `/api/v1/admin/clash/...` | 节点、端口配置、内核控制 |
| static | `frontend/dist` | SPA 托管 |

> 响应：`{ "error_msg": "", "data": ... }`。扩展时复制 `example` 切片；流程见 [`.agent/skills/new-api/SKILL.md`](../.agent/skills/new-api/SKILL.md)。

## 如何新增一个业务模块

以资源 `widget` 为例（完整步骤见 [`backend-development.md`](backend-development.md)）：

1. **复制 example 切片**
   - `domain/example` → `domain/widget`
   - `application/example` → `application/widget`
   - `repository/example.go` → `repository/widget.go`
   - `infra/persistence/relational/example_*` → `widget_*`，并更新 schema / models
   - `transport/http/example` → `transport/http/widget`
2. **在 `internal/app` 接线**  
   构造 repo → service → 传入 `httpserver.Dependencies`
3. **在 `transport/http/server.go` 注册路由**  
   挂到合适 group（通常需 `AdminAuth`）
4. **前端（可选）**  
   在 `features/` 下加 feature；路由挂 `app/router.tsx`；API 走 `shared/api`
5. **验证**  
   `go test ./...`、`go build ./cmd/foam`、手工或测试打通 CRUD

## 前端在架构中的位置

- **开发**：`pnpm dev` 在 `:8010` 提供 HMR，Vite 将 `/api` 等反代到后端 `:8000`（不必先 build）
- **出包 / 嵌入**：`pnpm build` → 后端静态托管（`frontend.staticPath`，默认 `./frontend/dist`）
- 浏览器同源访问 API；`publicApiBaseURL` 供前端运行时配置
- 路由约定见 [`frontend-development.md`](frontend-development.md)

## 配置边界

保留在 YAML 的：`server`、`auth`、`secrets`、`bootstrapAdmin`、`frontend`、`database`。  
新业务配置应有明确归属（YAML 启动项 vs 应用内设置）。

## 相关文档

- 后端细节：[`backend-development.md`](backend-development.md)
- 前端细节：[`frontend-development.md`](frontend-development.md)
- 目录地图：[`directory-map.md`](directory-map.md)
