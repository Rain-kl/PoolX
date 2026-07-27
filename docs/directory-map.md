# 目录地图

按目录说明职责与预期内容。路径以 **PoolX** 命名为准。

## 仓库根

| 路径 | 用途 |
| --- | --- |
| `AGENTS.md` | Agent/人类开发入口；always-on 规则 + docs 索引 |
| `config.example.yaml` | 配置模板；复制为 `config.yaml`（勿提交密钥） |
| `Makefile` | 本地 `run` 等快捷目标 |
| `Dockerfile` / `docker-compose.yml` / `docker/` | 容器构建与编排 |
| `docs/` | 架构与开发文档（按需阅读） |
| `backend/` | Go 服务 |
| `frontend/` | React 管理端 |
| `VERSION` | 版本号源（与 buildinfo 注入配合） |

## backend/

| 路径 | 用途 | 预期内容 |
| --- | --- | --- |
| `cmd/foam/` | 进程入口 | `main` → `cli.Run` |
| `internal/cli/` | 启动参数 | `--config`、`--listen`、版本 |
| `internal/app/` | 组合根 | DB、service 装配、HTTP server 生命周期、AutoStartKernel |
| `internal/buildinfo/` | 构建版本 | `CurrentVersion()` 等 |
| `internal/domain/` | 领域实体 | 每资源一子包；含 `admin` / `clash` / `settings` / `example` |
| `internal/domain/admin/` | 管理员实体 | 认证主体 |
| `internal/domain/clash/` | 代理池核心实体 | 节点、端口配置、内核实例、运行时配置 |
| `internal/domain/settings/` | 运行设置实体 | 通用可热更新配置 |
| `internal/domain/example/` | 示例实体 | 扩展模板 |
| `internal/application/` | 应用服务 | 每资源一子包；薄编排 |
| `internal/application/adminauth/` | 登录/刷新/会话用例 | |
| `internal/application/clash/` | 节点、端口配置、内核生命周期用例 | 含 AutoStartKernel |
| `internal/application/settings/` | 运行设置读写 | revision 乐观锁 |
| `internal/application/example/` | 示例 CRUD 用例 | 含测试 |
| `internal/repository/` | 持久化接口 | 接口 + `PageQuery` 等；无 SQL |
| `internal/infra/config/` | YAML 加载与校验 | |
| `internal/infra/security/` | JWT、密码、cipher | |
| `internal/infra/clash/` | 内核控制、订阅拉取、配置渲染 | kernel / proxy / runtimeconfig / sourcefetch |
| `internal/infra/persistence/relational/` | GORM 读写实现 | repo、models（无 AutoMigrate） |
| `internal/infra/persistence/migrator/` | goose SQL 迁移 | `goose/postgres` + `goose/sqlite` 双方言 |
| `internal/infra/observability/` | 日志等 | |
| `internal/transport/http/` | HTTP 服务器 | `server.go`、frontend static、middleware |
| `internal/transport/http/middleware/` | 鉴权、超时、request ID… | |
| `internal/transport/http/adminauth/` | 管理认证 handlers | |
| `internal/transport/http/clash/` | 节点、端口配置、内核控制 handlers | |
| `internal/transport/http/system/` | 管理端 system 信息 | |
| `internal/transport/http/example/` | `/api/v1/admin/examples` handlers | |
| `internal/transport/http/settings/` | `/api/v1/admin/settings` | |
| `internal/shared/response/` | 成功/错误 envelope | |

**依赖方向：** transport → application → domain；infra 实现 repository 接口。

## frontend/

| 路径 | 用途 | 预期内容 |
| --- | --- | --- |
| `src/main.tsx` | 浏览器入口 | |
| `src/app/` | 应用壳与路由 | providers、shell、router、auth boundary |
| `src/features/auth/` | 登录 feature | |
| `src/features/clash/` | 代理池核心 | 节点管理、端口配置、运行时控制、dashboard |
| `src/features/settings/` | 系统设置页 | 对接真实 settings API |
| `src/features/example/` | 示例 CRUD（扩展模板参考） | |
| `src/components/ui/` | 原子 UI | shadcn 风格，无业务耦合 |
| `src/shared/api/` | API 客户端 | envelope decode、token、错误类型 |
| `src/shared/auth/` | 会话上下文 | AuthProvider、useAuth |
| `src/shared/components/` | 跨页复合组件 | PageHeader、DataTableShell… |
| `src/shared/hooks/` | 通用 hooks | |
| `src/shared/lib/` | 纯工具 | cn、format、sort |
| `src/shared/config/` | 运行时配置 | API base 等 |
| `src/shared/i18n/` | 文案 | zh-CN / en |
| `src/types/` | 全局 TS 类型补充 | |
| `public/` | 静态资源 | |
| `dist/` | 构建输出 | 由后端 `frontend.staticPath` 托管 |

## docs/

| 路径 | 何时读 |
| --- | --- |
| `architecture.md` | 总览、请求流、加模块 |
| `backend-development.md` | 后端分层、example、测试、配置 |
| `frontend-development.md` | 路由、组件复用 |
| `directory-map.md` | 本文件：目录职责 |
| `superpowers/` | 设计/计划归档（演进记录，非日常开发必读） |

## 约定目录中的 README

以下目录保留短 `README.md`，说明层职责与如何复制 `example`：

- `backend/internal/domain/README.md`
- `backend/internal/application/README.md`
- `backend/internal/infra/README.md`
- `backend/internal/transport/http/README.md`

不要用"空 package 但无法编译"的占位；要么有真实代码，要么只有文档说明。

## 不应再出现的内容

- 前端已下线的 product feature 文件夹作为可路由模块
- 含真实密钥的 `config.yaml` 进入版本库
