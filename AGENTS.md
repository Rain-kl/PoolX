# AGENTS.md — Foam AI 助手工作操作手册

本文件面向 AI 开发助手，定义其职责与操作规范。

**Foam** 是可二次开发的全栈脚手架（Go 后端 + React 管理端示例画廊），不是空目录模板：

- 后端：薄组合根 + 管理员认证 + 系统设置 + `example` 垂直切片 CRUD
- 前端：admin shell 上的组件 / 页面示例画廊（`/example/*`）
- 扩展方式：复制 `example` 切片，而不是从零发明分层

| 项 | 值 |
| --- | --- |
| Go module | `github.com/Rain-kl/Foam/backend` |
| 二进制 / cmd | `foam`（`backend/cmd/foam`） |
| 前端包名 | `foam-frontend` |
| 默认配置 | 仓库根 `config.yaml`（从 `config.example.yaml` 复制） |

更细的长文按需阅读：`docs/architecture.md`、`docs/backend-development.md`、`docs/frontend-development.md`、`docs/directory-map.md`。

## Git 提交规范

遵循 Conventional Commits：`<type>(<scope>): <subject>`（例：`feat(example): add list filter`）。

- `type`：`feat` / `fix` / `docs` / `refactor` / `test` / `chore` 等
- 文案中性：使用 Foam / 通用 demo 用语，不要引入已剥离的业务产品话术

## 务必阅读匹配的 Skill

技能目录：`.agent/skills/<name>/SKILL.md`。

| Skill | 何时使用 |
| :--- | :--- |
| `new-api` | 添加或修改业务 API、垂直切片、Handler、application service、repository、路由注册（**优先读**） |
| `go-testing` | 编写或改造 Go 测试、表驱动、集成测试约定 |
| `go-logging` | 结构化日志、上下文日志字段 |
| `shadcn` | 添加、修改或组合 shadcn/ui 组件 |
| `code-review-skill` | 代码评审、PR 检查清单 |
| `release-guide` | 版本发布、Version Bump、Release 说明 |
| `database-migration` | 新增/修改表结构、索引、goose SQL（`migrator/goose/{postgres,sqlite}`）；**禁止生产 AutoMigrate** |

## 严格遵循事项 (Guardrails)

- 切勿删除 `frontend/node_modules`。
- **分层依赖**：`transport → application → domain`；`infra` / `repository` 实现通过接口注入；`domain` 禁止依赖 HTTP / DB / Gin / GORM。
- **扩展靠复制**：新业务资源优先完整复制 `example` 切片（domain / application / repository 接口 / relational 实现 / HTTP handler），再改名并接线到 `internal/app` 与 `transport/http/server.go`。
- **禁止**套用 Wavelet 的 `internal/apps/*`、`logics.go`、`internal/router/v1/*` 结构；Foam 的 HTTP 入口是 `backend/internal/transport/http/server.go`。
- 测试禁止硬编码相对路径创建临时目录，统一使用 `t.TempDir()`。
- API JSON 字段使用 **snake_case**；统一响应信封 `{ "error_msg": "", "data": ... }`（`internal/shared/response`）。
- 成功用 `response.Success` / `response.OK`；失败用 `response.Error` 或 `response.Abort*`（HTTP 状态码表达错误，body 不写独立 `error.code` 对象）。
- 管理端业务路由默认挂在 `/api/v1/admin/...`，使用 `middleware.AdminAuth`（`Authorization: Bearer`）；用户鉴权在 `/api/v1/user/*`。
- 组合根只在 `internal/app` wiring；禁止在 handler 里 new 全局单例或直接拼 SQL。
- 中性文案：UI、注释、提交说明使用 Foam / demo 用语。
- **证据再收工**：声称完成 / 通过 / 修好前，必须实际跑验证命令并贴出关键输出（`go test` / build / 路由 smoke）。
- 修改 API Handler 后视需要运行 `make swagger`；完成开发后应运行 `make format` 与 `make code-check`（或等价命令）。

## 技术栈与项目目录结构

### 技术栈

- **后端**：Go 1.26、Gin、GORM、SQLite / PostgreSQL、JWT + refresh cookie、Swaggo（可选）。
- **前端**：React 19、Vite、TypeScript、Tailwind CSS、pnpm、shadcn/ui、React Router、TanStack Query、i18next。

### 顶层目录

- `AGENTS.md`：本文件（AI 操作手册入口）。
- `config.example.yaml` / `config.yaml`：配置模板与本地配置（勿提交 secrets）。
- `.env.example` / `.env`：环境变量模板与本地密钥（`FOAM_*`；优先级高于 YAML；勿提交 `.env`）。
- `Makefile`：`dev` / `format` / `code-check` / `build-*` / `run` / `swagger` 等。
- `backend/`：Go module（`github.com/Rain-kl/Foam/backend`）。
- `frontend/`：React 管理端 SPA。
- `docs/`：架构与开发指南（人工维护）。
- `docker/` · `Dockerfile` · `docker-compose.yml`：容器化。
- `.agent/skills/`：按任务触发的技能说明。
- `bin/` · `data/` · `VERSION`：本地二进制、运行数据、版本号。

### 后端目录 (`backend/internal/`)

- `cmd/foam/`：进程入口。
- `cli/`：flags / `--config` / 启动参数。
- `app/`：**唯一组合根**（DB、bootstrap admin、DI、HTTP Server）。
- `domain/<name>/`：实体与纯领域类型。
- `application/<name>/`：用例服务（`context.Context`，无 Gin）。
- `repository/`：持久化接口与分页等共享查询类型。
- `infra/config` · `security` · `persistence/relational` · `observability`：配置、JWT/密码、GORM 实现、日志。
- `transport/http/`：Gin 引擎、中间件、各域 handler（`adminauth` / `example` / `settings` / `system`）。
- `shared/response/`：Wavelet 风格 API 信封。
- `buildinfo/`：版本注入。

### 前端目录 (`frontend/`)

- `src/app/`：providers、auth boundary、app shell、router、懒加载出口。
- `src/features/auth` · `example` · `settings`：登录、示例画廊、系统设置。
- `src/components/ui/`：shadcn/ui 原子组件。
- `src/shared/api` · `auth` · `components` · `config` · `i18n` · `lib`：HTTP 客户端、鉴权、跨页组件、运行时配置。
- `public/` · `vite.config.ts`：静态资源；开发服 `:8010` 反代 `/api` → 后端 `:8000`。

## 后端开发规范

### API 路径与响应

- **前缀**：`/api` + `/api/v1/...`；管理端 `/api/v1/admin/...`；用户鉴权 `/api/v1/user/...`。
- **探活**：`GET /api/health`（信封）；`/healthz`、`/readyz` 为探针别名。
- **信封**：

```json
{ "error_msg": "", "data": {} }
{ "error_msg": "未登录", "data": null }
```

- **成功**：`response.Success(c, status, data)` 或 `c.JSON(status, response.OK(data))`。
- **失败**：`response.Error(c, status, code, message)` 或 `Abort*`；禁止用 HTTP 200 携带业务失败。
- **列表**：`data` 内建议 `{ "items", "total", "page", "page_size" }`。
- **鉴权**：`Authorization: Bearer <access_token>`；refresh cookie `foam_admin_refresh` Path=`/api/v1/user`。

### 分层与新增资源

1. 复制 `example` 垂直切片并改名（见 skill `new-api`）。
2. `relational`：`models.go` / `schema.go` / mapping / repository 实现。
3. `app.New` 注入 service → `httpserver.Dependencies`。
4. `server.go` 将 handler 挂到 `admin`（或明确公开的）组。
5. 补 application / handler 测试；`cd backend && go test ./...`。

### 数据库

- 驱动：`sqlite`（本地默认）或 `postgres`（`config.yaml` → `database.driver`）。
- **Schema 迁移**：`pressly/goose` + 嵌入 SQL（`internal/infra/persistence/migrator/goose/{postgres,sqlite}`），启动时 `InitializeSchema` → `migrator.Up`。
- GORM model 仅做读写映射；**生产路径禁止 AutoMigrate**。双方言同版本号、同文件名；无物理外键。
- 禁止在 Handler 写复杂 SQL；映射错误经 repository 边界处理，勿把底层错误细节直接返回客户端。

## 前端开发规范

- 开发：`pnpm dev`（`:8010` HMR + API 反代）；出包：`pnpm build` → 后端 `frontend.staticPath` 托管 `dist`（嵌入模式）。
- 画廊页优先 **mock/static**；真 API 仅登录会话、设置、可选 example CRUD。
- 路由与导航：`src/app/router.tsx`、`app-shell.tsx`；设置入口在侧栏底部「更多」右侧。
- API 调用走 `shared/api/client.ts`（解析 `error_msg` + `data`）；路径与后端 Wavelet 风格一致。
- 原子 UI 放 `components/ui`；跨页模式放 `shared/components`；域内逻辑放 `features/<area>`。
- 优先 shadcn `variant` 与 CSS 变量，避免业务里硬编码颜色。
- 文案走 i18n（`zh-CN` + `en`）；产品名 **Foam**。

## 常用命令

```bash
# 配置
cp config.example.yaml config.yaml   # 填 secrets + bootstrapAdmin

# 开发
make dev              # 前端 :8010 + 后端 :8000
make dev-f / make dev-b
make run              # 仅后端

# 质量
make format
make code-check
cd backend && go test ./...
cd frontend && pnpm exec tsc -b

# 构建
make build-frontend   # → frontend/dist
make build-backend    # → bin/foam
make build-embedded   # dist + 二进制
make cross-build
make swagger
```

## 优先级

1. 用户当前明确指令  
2. 本文件（`AGENTS.md`）与匹配的 `.agent/skills/*`  
3. `docs/*` 中的详细约定  
4. 现有 `example` 切片与邻近代码风格  
