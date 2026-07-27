# 后端开发

Module：`github.com/Rain-kl/Foam/backend`  
Binary：`foam`（`backend/cmd/foam`）

## 技术栈

- Go、Gin、GORM
- SQLite（本地/单实例）或 PostgreSQL
- JWT + cookie 会话（admin auth）
- 统一响应：`internal/shared/response`

## 分层职责

| 层 | 路径 | 做什么 | 不做什么 |
| --- | --- | --- | --- |
| domain | `internal/domain/<name>` | 实体与纯领域类型 | import HTTP/DB |
| application | `internal/application/<name>` | 用例、输入校验、编排 repo | JSON binding、SQL |
| repository | `internal/repository` | 接口与共享查询类型 | 具体驱动实现 |
| infra | `internal/infra/**` | config、security、GORM 实现、日志 | 业务用例逻辑 |
| transport | `internal/transport/http/<name>` | 路由、DTO、状态码 | 直接拼 SQL |
| app | `internal/app` | 组合根、生命周期 | 散落业务规则 |

层间约定 README 也在各目录下（`domain/`、`application/`、`infra/`、`transport/http/`）。

## example 垂直切片（模板）

最小字段：`id`, `name`, `description`, `created_at`, `updated_at`。

| 文件 | 角色 |
| --- | --- |
| `domain/example/example.go` | 实体 |
| `application/example/service.go` | List/Get/Create/Update/Delete |
| `repository/example.go` | `ExampleRepository` 接口 |
| `infra/persistence/relational/example_repository.go` | GORM 实现 |
| `infra/persistence/relational/models.go` / `schema.go` | 表模型与迁移 |
| `transport/http/example/handler.go` | `/api/v1/admin/examples` CRUD |

### HTTP 契约（示例，对齐 Wavelet）

- `GET /api/v1/admin/examples?page=&page_size=&search=`
- `GET /api/v1/admin/examples/:id`
- `POST /api/v1/admin/examples` body: `{ "name", "description" }`
- `PUT /api/v1/admin/examples/:id`
- `DELETE /api/v1/admin/examples/:id`

均需 **admin JWT**（`middleware.AdminAuth`，`Authorization: Bearer`）。  
成功 / 错误体：`{ "error_msg": "", "data": ... }` / `{ "error_msg": "...", "data": null }`（HTTP 状态码区分结果）。

### 复制切片清单

新增资源时按顺序：

1. 复制 domain / application / repository 接口 / relational 实现 / HTTP handler
2. 改 package 名、类型名、表名、路由前缀
3. 在 `app.New` 创建 repo + service
4. 在 `httpserver.Dependencies` 增加字段并在 `server.go` 注册
5. 补 `*_test.go`（handler 与 service 至少各有覆盖）
6. `go test ./...` 与 `go build -o foam ./cmd/foam`

## 组合根

`internal/app/application.go` 负责：

1. 按 `database.driver` 打开 SQLite / Postgres
2. `InitializeSchema`
3. 校验 `credentialEncryptionKey`（cipher 可用）
4. Bootstrap 管理员（仅库中无管理员时）
5. 装配 `adminauth` + `clash` + `settings` + `example` 等 service
6. 构造 Gin engine 与 `http.Server`
7. goroutine 调用 `AutoStartKernel`（配置与二进制可用时自动拉起内核）

新依赖只在这里 wiring，避免在 handler 里 new 全局单例。

## 数据库迁移（goose + SQL）

对齐 Wavelet：版本化 SQL 嵌入二进制，启动时执行。

| 路径 | 说明 |
| --- | --- |
| `internal/infra/persistence/migrator/` | `Up(ctx, gormDB, driver)` |
| `migrator/goose/postgres/*.sql` | PostgreSQL |
| `migrator/goose/sqlite/*.sql` | SQLite（同版本号/文件名） |

- 调用链：`app.New` → `database.InitializeSchema` → `migrator.Up`
- 版本表：`goose_db_version`
- 新表/改列：新增 goose 文件 + 更新 `relational` model；详见 `.agent/skills/database-migration/SKILL.md`
- 管理员仍由 `bootstrapAdmin` 配置 bootstrap，不写进 SQL seed

## 配置

优先级：

```text
CLI 标志（--listen 等）
  > 环境变量 FOAM_*（含 .env 仅填充未设置键）
  > config.yaml
  > 代码默认值
```

从仓库根复制：

```bash
cp config.example.yaml config.yaml
cp .env.example .env   # 可选；适合密钥与部署差异
```

必改（YAML 或环境变量二选一）：

- `secrets.jwtSecret` / `FOAM_SECRETS_JWT_SECRET`（≥ 32 字符）
- `secrets.credentialEncryptionKey` / `FOAM_SECRETS_CREDENTIAL_ENCRYPTION_KEY`（base64 32 bytes）
- `bootstrapAdmin.password` / `FOAM_BOOTSTRAP_ADMIN_PASSWORD`

常用：

- `server.listen` / `FOAM_SERVER_LISTEN` 默认 `127.0.0.1:8000`
- `database.driver` / `FOAM_DATABASE_DRIVER`: `sqlite` | `postgres`
- `frontend.staticPath` / `FOAM_FRONTEND_STATIC_PATH`
- `frontend.publicApiBaseURL` / `FOAM_FRONTEND_PUBLIC_API_BASE_URL`
- 配置文件路径：`FOAM_CONFIG` 或 `CONFIG_PATH`（CLI `--config` 更高）

完整环境变量表见 [`.env.example`](../.env.example)。加载与校验：`internal/infra/config`。

## 管理员认证

- 应用服务：`application/adminauth`
- HTTP：`transport/http/adminauth`
  - `POST /api/v1/user/login` · `POST /api/v1/user/refresh` · `GET /api/v1/user/logout`
  - `GET /api/v1/user/self` · `GET /api/v1/user-info` · `POST /api/v1/user/change-password`

## 系统设置（runtime settings）

通用可热更新配置：

| 层 | 路径 |
| --- | --- |
| domain | `domain/settings` |
| application | `application/settings` |
| repository | `repository/settings.go` + relational 实现 |
| HTTP | `GET/PUT /api/v1/admin/settings` |

字段：`app.display_name`、`frontend.public_api_base_url`。更新使用 `revision` 乐观锁；空值回退到 `config.yaml` / 默认 `PoolX`。
- Token：`infra/security`（JWT）
- 会话持久化：relational admin session repo

前端登录页依赖这组 API；不要为其他页面再发明第二套鉴权。

## 测试

```bash
cd backend
go test ./...
go test -race ./...
go vet ./...
go build -o foam ./cmd/foam
```

建议：

- **service 测试**：假 repo / 内存实现，覆盖校验与错误分支
- **handler 测试**：`httptest` + Gin engine，断言状态码与 envelope
- **middleware / config**：边界与安全相关逻辑单独测

## 本地运行

```bash
# 仓库根（Makefile 目标以仓库当前文件为准）
make run

# 或
cd backend && go run ./cmd/foam --config ../config.yaml
```

需要前端静态文件时先：

```bash
cd frontend && pnpm install && pnpm build
```

## 反模式

- 在 `domain` 引用 `gin` / `gorm`
- 在 handler 里写业务规则或直接操作 DB
- 不复制 example 就新建"特殊"目录结构
- 只改代码不跑 `go test` / `go build` 就宣称完成
