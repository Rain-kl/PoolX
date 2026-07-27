---
name: "new-api"
description: "Foam 项目专用：新增或修改业务 API、垂直切片、Handler、application service、repository、路由注册时必须使用。指导按 example 切片复制扩展、Wavelet 风格路径与 {error_msg,data} 响应、组合根接线与质量门禁；纠正 Wavelet apps/* 结构、custom 垃圾桶、跳过分层等错误写法。"
---

# 新增业务 API（Foam 垂直切片）

本技能是 Foam 后端新增接口的唯一落点规范。开发任何新资源 / 管理端 API 前，先按本指南决策分层与注册位置。

更细的仓库约定见：

- [`docs/architecture.md`](../../../docs/architecture.md)
- [`docs/backend-development.md`](../../../docs/backend-development.md)
- always-on：仓库根 `AGENTS.md` / `Agents.md`

---

## 先搞清：脚手架 vs 产品化

Foam 是**可运行的全栈脚手架**（Go + React 示例画廊），不是 Wavelet 那套 `internal/apps/*` 结构。

| 层级 | 含义 | 典型落点 |
| :--- | :--- | :--- |
| **平台能力** | 脚手架自带 | `adminauth`、`settings`、health、JWT 中间件 |
| **扩展模板** | 演示如何加资源 | **`example` 垂直切片**（复制它，不要从零发明分层） |
| **产品业务** | 二次开发真实域 | 复制 `example` → 改名为业务域（如 `widget`） |

**扩展靠复制**：新业务模块优先完整复制 `example` 切片，再改名与接线。禁止把业务塞进 `adminauth` / `settings` / 前端 mock 页。

---

## 分层与目录（必须遵守）

依赖方向：

```text
transport/http  →  application  →  domain
                        ↓
                   repository (接口)
                        ↑
              infra/persistence/relational (实现)
```

以资源 **`widget`** 为例，复制 `example` 后应对齐：

```text
backend/internal/
├── domain/widget/widget.go
├── application/widget/service.go          (+ service_test.go)
├── repository/widget.go                   # 接口
├── infra/persistence/relational/
│   ├── models.go                          # 增加 widgetModel
│   ├── schema.go                          # schemaModels + 索引
│   ├── mapping.go                         # toWidgetDomain
│   └── widget_repository.go
├── transport/http/widget/handler.go       (+ handler_test.go)
└── app/application.go                     # 构造 repo → service → Dependencies
    transport/http/server.go               # 注册到 /api/v1/admin 或公开组
```

| 层 | 做什么 | 不做什么 |
| :--- | :--- | :--- |
| `domain` | 实体与纯类型 | import Gin / GORM / HTTP |
| `application` | 用例、校验、编排 repo、`context.Context` | JSON binding、SQL、`*gin.Context` |
| `repository` | 接口 + `PageQuery` 等共享查询类型 | 具体驱动 |
| `infra/.../relational` | GORM model、映射、实现 | 业务规则文案散落 |
| `transport/http` | 路由、DTO、状态码、envelope | 直接拼 SQL |
| `app` | **唯一组合根** wiring | handler 里 new 全局单例 |

参考代码骨架：`references/`（`domain` / `repository` / `service` / `handler`）。

---

## 反模式（AI 最常踩的坑）

### 1. 套用 Wavelet / 其他仓库的目录

| 错误 | 正确 |
| :--- | :--- |
| `internal/apps/<domain>/routers.go` + `logics.go` | Foam 的 domain / application / repository / transport |
| `internal/router/v1/<domain>.go` | 在 `transport/http/<domain>` 写 `Register`，于 `server.go` 挂载 |
| `module github.com/.../Wavelet` | `github.com/Rain-kl/Foam/backend` |

### 2. 跳过切片、只写 Handler

| 错误 | 正确 |
| :--- | :--- |
| 只在 `server.go` 里闭包查库 | 完整切片 + 接口注入 |
| domain 里 import GORM | domain 纯类型 |
| application 里 `ShouldBindJSON` | 绑定只在 transport |

### 3. 路径与 envelope 用旧 Foam / 非 Wavelet 风格

| 错误 | 正确 |
| :--- | :--- |
| `/api/admin/v1/...`、`/api/examples` | `/api/v1/admin/...`、`/api/v1/user/...` |
| `{ "data": T }` / `{ "error": { "code", "message" } }` | `{ "error_msg": "", "data": T\|null }` |
| 随意 camelCase JSON | API JSON **snake_case**（`page_size`、`created_at`…） |

### 4. 其它防线

- 不要把业务塞进 `settings` 的 revision 配置或前端 mock 画廊顶替真 API。
- 管理端业务默认 **`middleware.AdminAuth`**（Bearer JWT）；公开接口明确标注。
- 错误响应优先 `response.Error` / `response.Abort*`；成功用 `response.Success` 或 `c.JSON(..., response.OK(...))`。
- 中性文案：Foam / demo，不引入已剥离产品话术。

---

## API 约定（对齐 Wavelet）

### 路径归属

| 目标 | 注册位置 | 鉴权 |
| :--- | :--- | :--- |
| `GET /api/health`、`/healthz`、`/readyz` | `server.go` 探针 | 公开 |
| `/api/v1/user/*` | `transport/http/adminauth` | login/refresh 公开；self / change-password 需 JWT |
| `GET /api/v1/config/public` | `server.go` | 公开 |
| `/api/v1/admin/<resource>/...` | 各 `transport/http/<name>`，挂到 `admin` 组 | **AdminAuth** |
| SPA | `registerFrontend` | 非 API 回退 |

**产品资源默认**：`/api/v1/admin/widgets`（复数资源名，REST CRUD），与 `examples` 同级。

### 响应 envelope

```json
{ "error_msg": "", "data": { } }
{ "error_msg": "请求参数无效", "data": null }
```

- 成功：HTTP 2xx + 空 `error_msg` + `data`
- 失败：HTTP 4xx/5xx + `error_msg`（body **无**独立 `error.code` 对象；`response.Error` 的 code 参数仅便于调用处标注，不写出）
- 列表建议：`{ "items", "total", "page", "page_size" }`
- 时间字段：RFC3339 字符串，`created_at` / `updated_at`

### 鉴权

- Header：`Authorization: Bearer <access_token>`
- Refresh cookie：`foam_admin_refresh`，Path=`/api/v1/user`（HttpOnly）
- 未登录文案对齐：`未登录`（401）

---

## 核心开发步骤

### 步骤 1：复制 example 切片并改名

```bash
# 概念步骤（以 widget 为例）
domain/example          → domain/widget
application/example     → application/widget
repository/example.go   → repository/widget.go
relational example_*    → widget_*（models/schema/mapping）
transport/http/example  → transport/http/widget
```

改 package 名、类型名、表名、路由前缀、错误文案。

### 步骤 2：domain

纯实体，无标签依赖。见 `references/domain_example.go`。

### 步骤 3：repository 接口 + relational 实现

- 接口放 `internal/repository/<name>.go`，复用 `PageQuery` / `NormalizePage`。
- 实现放 `relational/<name>_repository.go`；在 `models.go` 加 model；`mapping.go` 做 domain 映射。
- **表结构**：按 [database-migration](../database-migration/SKILL.md) 在 `migrator/goose/{postgres,sqlite}` 写双方言 goose SQL（**禁止**生产 AutoMigrate）。

### 步骤 4：application service

- 构造函数注入 repository 接口。
- 方法首位 `context.Context`。
- 校验输入、映射 `repository.ErrNotFound` → 本模块 `ErrNotFound`。
- 见 `references/service_example.go`；补 `service_test.go`。

### 步骤 5：HTTP handler

- `NewHandler(service)` + `Register(group *gin.RouterGroup)`。
- DTO：`json` snake_case；binding 校验。
- 成功：`response.Success(c, status, data)`。
- 失败：`response.Error(c, status, "stableCode", "人类可读中文/英文案")`。
- 见 `references/handler_example.go`；补 `handler_test.go`。

### 步骤 6：组合根接线

1. `internal/app/application.go`：`NewXxxRepository` → `NewService` → 填入 `httpserver.Dependencies`。
2. `internal/transport/http/server.go`：
   - 在 `Dependencies` 增加字段；
   - 在 `admin`（或公开）组调用 `xxxhttp.NewHandler(...).Register(admin)`。

```go
// server.go 片段（管理端资源）
admin := v1.Group("/admin")
admin.Use(middleware.AdminAuth(deps.AdminAuth))
// ...
if deps.Widgets != nil {
    widgethttp.NewHandler(deps.Widgets).Register(admin)
}
```

Handler 内注册路径应为相对 group 的 `/widgets`，最终 **`/api/v1/admin/widgets`**。

### 步骤 7：前端（可选）

- API 客户端：`frontend/src/shared/api/client.ts`（已解析 `error_msg` + `data`）。
- Feature：`frontend/src/features/<domain>/`；路径用 `/api/v1/admin/...`。
- 画廊页可继续 mock；真 CRUD 再接后端。

### 步骤 8：验证（证据再收工）

```bash
cd backend && go test ./...
cd backend && go build -o foam ./cmd/foam
# 或仓库根
make code-check
make format
make swagger   # 若改了 swag 注解 / 需更新 docs
```

手工 smoke 示例：

```bash
# 登录拿 access_token（cookie 会带 refresh）
curl -sS -c /tmp/foam.ck -X POST http://127.0.0.1:8000/api/v1/user/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"..."}'

curl -sS -b /tmp/foam.ck http://127.0.0.1:8000/api/v1/admin/widgets \
  -H "Authorization: Bearer <access_token>"
```

---

## 与平台路由的边界

| 场景 | 做法 |
| :--- | :--- |
| 扩展登录 / 改密 / me | 改 `adminauth`，路径保持 `/api/v1/user/*` |
| 运行时可改配置 | `settings` + `/api/v1/admin/settings` |
| 新业务资源 | **新切片** + `/api/v1/admin/<resources>` |
| 仅探活 / 版本 | `/api/health`、`/api/v1/admin/status`、`/version` |
| 公开只读配置 | `/api/v1/config/public` 或专用公开 handler |

---

## 质量验证门禁

1. `cd backend && go test ./...`（至少 application + transport handler）
2. `go build ./cmd/foam` 或 `make build-backend`
3. `make format`（gofmt + 前端 biome，若动了前端）
4. `make code-check`（golangci-lint + 前端 tsc/eslint，按改动范围）
5. 有 Swagger 注解变更时：`make swagger`
6. 声称完成前贴出关键命令输出（AGENTS always-on）

---

## 自检清单

- [ ] 从 `example` 完整复制切片，而非只写 handler / 套 Wavelet `apps/*`
- [ ] domain 无 HTTP/DB；application 无 gin；SQL 仅在 relational
- [ ] repository 接口在 `internal/repository`，实现经 `app` 注入
- [ ] 路由挂在 `/api/v1/admin/...`（或明确公开的 `/api/v1/...`），非旧 `/api/admin/v1` 或 `/api/examples`
- [ ] 响应为 `{error_msg,data}`；JSON 字段 snake_case
- [ ] 管理端使用 `middleware.AdminAuth`
- [ ] `schema` / models 已登记；`server.go` + `Dependencies` 已接线
- [ ] 有 service / handler 测试；`go test ./...` 通过
- [ ] 文案中性（Foam / demo）
