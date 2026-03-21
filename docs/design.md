# GinNextTemplate 设计基线

本文档定义 `GinNextTemplate` 当前有效的产品定位、系统边界、核心能力与架构约束。

当前项目已完成从历史业务工程到模板工程的主线切换。默认前提是：后续开发围绕模板工程展开，而不是继续恢复历史业务产品。

如果新增需求超出本文档范围，必须先更新本文档，再进入实现。

## 1. 产品定位

`GinNextTemplate` 是一个面向后台管理类项目的全栈模板工程，提供开箱即用的通用基础能力、管理端界面骨架和可继续扩展的工程结构。

模板当前主要价值：

* 为新的 Gin + Next.js 项目提供可直接复用的起点
* 提供统一的用户、鉴权、邮件、文件、安全、升级能力
* 提供可继续沉淀的服务端分层结构与前端 feature 组织方式
* 降低新项目从零搭建后台基础设施的成本

## 2. 范围边界

当前明确保留：

* 用户注册、登录、登出、会话、Token、自助资料维护
* 用户管理、角色与状态控制
* 邮箱验证码、密码重置、邮箱绑定
* 文件上传、删除、检索、下载与元数据管理
* 系统设置、系统公告、关于页与公共状态信息
* 应用日志记录与管理端查看
* 服务端版本检查、手动上传升级包和服务端升级流程
* 后台管理端布局、导航、请求层、主题与通用反馈组件

当前明确不纳入模板主线：

* Agent 子系统
* 节点管理、心跳与同步
* 配置版本分发
* OpenResty 代理规则
* 域名与证书分发
* 节点观测、流量分析、看板与相关生态能力

当前明确不做：

* 多租户平台化
* 工作流引擎
* 通用消息队列平台化
* 对象存储平台化
* 链路追踪与云资源编排等超出模板边界的基础设施产品化

## 3. 技术基线

### 3.1 Server

服务端当前采用：

* Go 1.24+
* Gin
* GORM
* SQLite / PostgreSQL
* Session 登录体系

### 3.2 Frontend

前端当前采用：

* Next.js App Router
* React 19
* TypeScript
* Tailwind CSS
* TanStack Query

## 4. 整体架构

```text
GinNextTemplate
├── Server
│   ├── Admin API
│   ├── Auth / Mail / File / Security / Settings / Logs / Upgrade
│   └── Static Web Host
└── Web Admin
    ├── Public Pages
    └── Dashboard Features
```

职责划分：

* Server：提供管理端 API、数据库访问、认证、上传、邮件、安全、日志与升级能力
* Web Admin：提供模板化后台界面、配置入口与交互体验

## 5. 当前工程组织

当前项目的服务端工作目录是 `server/`，并已建立以下结构基线：

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

当前约束：

* `cmd/server/` 作为服务启动入口
* `internal/handler/` 负责接口层
* `internal/service/` 负责业务逻辑编排
* `internal/model/` 负责数据模型与迁移
* `internal/middleware/` 负责通用中间件
* `internal/router/` 负责路由注册
* `internal/pkg/` 负责项目内公共能力
* `web/` 作为管理端前端工程与构建产物目录

`repository/`、`dto/`、顶层 `pkg/` 仍是允许的后续演进方向，但是否引入应以实际抽象收益为准，不做为了结构而结构的迁移。

## 6. 架构原则

### 6.1 服务端分层

服务端继续遵循 MVC 扩展分层：

* Handler：请求解析、权限入口、响应封装
* Service：业务规则、流程编排、事务边界
* Model：数据模型、迁移与持久化表达
* Middleware：横切关注点
* Router：路由组织与挂载

### 6.2 前后端同步

能力变更时，必须同步处理：

* 后端接口
* 前端页面
* 导航入口
* API client
* 类型定义
* Swagger
* 部署文档
* 配置项说明

### 6.3 数据库迁移显式化

数据库 schema 变更必须显式维护版本与迁移逻辑，迁移代码统一收敛在 `server/internal/model/migrate/`。

`AutoMigrate` 可以作为建表与补齐的底层手段，但不能替代版本化迁移方案本身。

迁移实现约束：

* `server/internal/model/migrate/` 下的公共迁移基础设施应集中维护，不按单个小方法零散拆文件
* 每个版本文件负责自身版本的升级步骤、执行顺序和校验逻辑
* `server/internal/model/main.go` 只负责迁移接入，不承载具体版本升级实现
* `server/internal/model/migrate/scheduler.go` 只负责通用调度，不承载业务化迁移步骤
* 升级过程按版本顺序逐级推进，遵循类似 Android 数据库升级的演进方式

## 7. 核心对象

模板当前长期保留的核心对象包括：

* `users`
* `files`
* `options`
* `app_logs`
* 与升级、安全、邮件流程相关的必要内部对象

历史业务对象如 `nodes`、`config_versions`、`proxy_routes`、`managed_domains`、`tls_certificates` 等，不属于模板主线对象。

## 8. 文档维护原则

以下内容变化时，应同步更新本文档：

* 产品范围或系统边界变化
* 模板保留模块变化
* 核心架构与职责划分变化
* 服务端工程组织基线变化
* 数据库迁移策略变化
