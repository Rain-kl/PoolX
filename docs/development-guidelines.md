# 模板工程开发规范

本文档描述 `GinNextTemplate` 在模板工程阶段的开发基线。

当前前提是：模板化改造已基本完成，后续开发默认以复用、维护和模板内扩展为目标，而不是继续沿历史业务方向演化。

如果需求超出 `docs/design.md` 的边界，必须先更新设计文档，再同步更新相关规范与计划文档后进入实现。

## 1. 技术基线

### 1.1 Server

服务端基线：

* Go 1.24+
* Gin
* GORM
* SQLite / PostgreSQL
* 现有 Session 登录体系

### 1.2 Frontend

前端基线以 `server/web` 为准：

* Next.js 15 App Router
* React 19
* TypeScript
* Tailwind CSS 4
* TanStack Query
* React Hook Form + Zod
* Zustand 仅用于轻量客户端 UI 状态

前端细则见 `docs/frontend-development-guidelines.md`。

## 2. 工程结构规范

当前服务端以 `server/` 为工作根目录，已建立如下结构：

```text
server/
├── cmd/server/
├── internal/app/
├── internal/handler/
├── internal/service/
├── internal/model/
├── internal/middleware/
├── internal/router/
├── internal/pkg/
├── docs/
└── web/
```

要求：

* 新增服务端代码优先落到现有 `server/internal/*` 分层
* `cmd/server/` 继续作为默认启动入口
* 不继续扩张旧的平铺结构
* 如确有必要，可按模块演进出 `repository/`、`dto/`，但必须有清晰收益
* 目录调整要与模块边界、测试和文档同步完成
* 新增的方法、类型、函数、结构和模块，必须放在符合其功能定义的包下
* 开发时必须持续关注代码仓库整洁度，优先保证可维护性，而不是临时堆叠代码

禁止：

* 在语义含糊、职责不清的包下堆放任意代码
* 因为“先跑起来”而把本应归属明确的实现塞进临时包、杂项包或不匹配的目录
* 用模糊包名掩盖分层问题和职责边界问题

## 3. 分层约束

服务端必须继续遵循当前分层职责：

* Handler：请求入参解析、权限检查入口、响应封装
* Service：业务规则、流程编排、事务边界
* Model：数据模型、数据库访问表达、迁移逻辑
* Middleware：认证、限流、跨域、校验等横切能力
* Router：路由声明与模块挂载

禁止：

* 在 Handler 中堆积核心业务逻辑
* 在 Middleware 中实现具体业务流程
* 在 Router 中实现业务判断
* 在 Model 中编排跨模块接口流程

要求：

* 接口层只负责接口问题
* 业务规则优先收敛在 Service
* 持久化结构与迁移逻辑收敛在 Model
* 横切能力不得演变为隐式业务入口
* 新增实现必须优先判断职责归属，再决定落包位置；不能先写进“差不多能放”的目录，后续再长期搁置

## 4. 模板模块约束

当前模板长期保留模块：

* 用户
* 邮箱
* 文件上传
* 安全
* 服务端升级
* 系统设置
* 应用日志

当前不纳入模板主线的模块：

* Agent
* 节点管理与心跳同步
* 配置版本分发
* OpenResty 代理规则
* 域名与证书分发
* 观测分析与相关后台能力

要求：

* 新增模块应优先建立在通用模板能力上
* 不新增新的历史业务对象或历史业务入口
* 删除或替换能力时，必须同步处理后端、前端、Swagger、配置项与文档
* 不允许保留失效导航、失效 API、失效类型或残留说明

## 5. 数据模型与数据库规范

要求：

* 数据模型应持续围绕模板长期保留对象收敛
* 表结构或内部元数据变更必须同步处理数据库版本与迁移
* 空库初始化、旧库升级和迁移失败回滚都要纳入考虑

### 5.1 迁移规范

* 迁移实现统一放在 `server/internal/model/migrate/`
* 迁移基础设施应集中维护在统一的公共实现中，不要按单个小方法零散拆文件
* 迁移按版本拆分文件，例如 `v0.go`、`v1.go`
* 每个版本文件必须内聚该版本的迁移步骤、执行顺序和校验逻辑
* 迁移目录的实现应借鉴 Android 数据库升级思想，按版本顺序逐级升级
* `server/internal/model/main.go` 只负责迁移接入与启动装配，不承载具体版本升级步骤
* `server/internal/model/migrate/scheduler.go` 只负责通用调度，不承载业务化迁移实现
* 不得仅依赖 `AutoMigrate` 隐式升级存量数据库
* 每次提升 schema 版本时，必须补充显式迁移与校验
* 迁移失败时，启动流程必须中止

## 6. API 与鉴权规范

### 6.1 API

要求：

* 管理端 API 统一使用 JSON
* 响应统一包含 `success`、`message`、`data`
* 变更类接口统一使用 `POST`
* 只读接口优先使用 `GET`
* 接口变更必须同步处理前端请求封装、类型定义和 Swagger

统一响应结构：

```json
{
  "success": true,
  "message": "",
  "data": {}
}
```

### 6.2 鉴权

管理端继续基于现有登录、角色与 Session。

安全要求：

* 不暴露远程 shell 或任意命令执行入口
* 不在日志中打印完整 Token、验证码或敏感密钥
* 不绕过统一鉴权中间件新增临时后门接口

## 7. 配置规范

要求：

* 配置项必须围绕模板保留模块收敛
* 新增配置项时必须同步更新 `docs/app-config.md`
* 配置变更影响部署方式时，必须同步更新 `docs/deployment.md` 与 `README.md`
* 已移除业务的兼容配置只能用于迁移与兜底，不应重新成为正式能力入口

## 8. 测试与交付要求

* 关键业务逻辑必须有单元测试或等效回归测试
* 服务端改动至少完成一次 `go test ./...` 或等效验证
* 前端改动至少完成一次 `pnpm build`，必要时补充 `pnpm test`
* 涉及页面与接口联动的改动，必须验证前后端是否同步
* 合入前至少保证服务可启动、核心保留模块可用

## 9. 文档维护要求

以下内容变化时，必须同步更新对应文档：

* 模板范围或系统边界变化：更新 `docs/design.md`
* 开发约束、分层规则或目录结构变化：更新本文档
* 前端工程约束变化：更新 `docs/frontend-development-guidelines.md`
* 配置项、部署方式或启动方式变化：更新 `docs/app-config.md`、`docs/deployment.md` 和 `README.md`
