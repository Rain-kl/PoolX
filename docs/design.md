# Proxy Kernel Control Plane 设计基线

本文档定义当前项目在“基于 `GinNextTemplate` 演进业务产品”阶段的有效产品定位、系统边界、核心能力与架构约束。

当前前提是：项目不再以“纯模板工程维护”为唯一目标，而是以模板工程为基础，持续交付代理池控制端业务。后续开发默认围绕 Proxy Kernel Control Plane 展开，同时保留模板提供的通用基础设施能力。

如果新增需求超出本文档范围，必须先更新本文档，再进入实现。

## 1. 产品定位

当前项目定位为面向代理池场景的单实例 Web 控制台，首期以 Mihomo 为运行内核，为用户提供节点导入、节点测试、节点池管理、工作台配置、运行控制与状态查看能力。

当前阶段的定位原则：

* 以 `GinNextTemplate` 的用户、鉴权、上传、设置、日志和升级能力作为产品底座
* 以“统一控制平面”而非“前端直连内核”作为核心交互模式
* 以 Mihomo MVP 闭环为当前首要交付目标，同时为 sing-box / Xray 保留扩展抽象
* 以后台管理产品的结构清晰、可维护、可扩展为默认工程目标

## 2. 范围边界

当前明确纳入主线：

* 用户注册、登录、登出、会话、Token、自助资料维护
* 用户管理、角色与状态控制
* 邮箱验证码、密码重置、邮箱绑定
* 文件上传、删除、检索、下载与元数据管理
* 系统设置、系统公告、关于页与公共状态信息
* 应用日志记录与管理端查看
* 服务端版本检查、手动上传升级包和服务端升级流程
* 配置源上传、解析、节点标准化、双层去重与导入确认
* 节点池管理、搜索筛选、启用禁用、批量测试、测试结果展示
* 工作台配置，包括监听入口、节点集合、策略、测试 URL、间隔与预览
* 运行配置渲染、快照、校验和比较、启动、停止、热重载、运行状态与日志
* 内核能力发现、能力协商与按能力降级的前端体验

当前明确不纳入 MVP：

* 多租户平台化
* 多用户隔离工作区
* 多实例并行运行
* 复杂规则编辑器
* 订阅源拉取与定时同步
* 对外暴露内核私有控制 API
* 云端分布式节点编排、Agent 集群或边缘分发体系

当前明确不做：

* 把不同内核的私有配置格式直接暴露为前端主模型
* 用户直接编辑唯一运行中配置文件并绕过数据库
* 通过拼接 shell 的方式执行内核控制命令

## 3. 技术基线

### 3.1 Server

服务端当前采用：

* Go 1.24+
* Gin
* GORM
* SQLite / PostgreSQL
* Session 登录体系
* 基于 `os/exec` 的本地进程控制

### 3.2 Frontend

前端当前采用：

* Next.js App Router
* React 19
* TypeScript
* Tailwind CSS
* TanStack Query
* React Hook Form + Zod

## 4. 整体架构

```text
Proxy Kernel Control Plane
├── Server
│   ├── Admin API
│   ├── Auth / Mail / File / Security / Settings / Logs / Upgrade
│   ├── Source Import / Node Pool / Workspace / Runtime Control
│   └── Static Web Host
└── Web Admin
    ├── Public Pages
    ├── Dashboard
    ├── Source Import / Node Pool / Workspace / Runtime
    └── Settings / About
```

职责划分：

* Server：提供管理端 API、数据库访问、认证、上传、日志、升级、配置渲染与运行控制能力
* Web Admin：提供统一管理端界面、配置入口、运行视图与反馈体验
* Kernel Adapter：屏蔽不同内核在渲染、启动、重载、状态读取上的实现差异

## 5. 当前工程组织

当前项目仍以 `server/` 为服务端工作目录，并保持既有工程基线：

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
└── web/
```

当前约束：

* `cmd/server/` 作为服务启动入口
* `internal/handler/` 负责接口层
* `internal/service/` 负责业务逻辑编排
* `internal/model/` 负责数据模型与迁移
* `internal/middleware/` 负责通用中间件
* `internal/router/` 负责路由注册
* `internal/pkg/` 负责项目内公共能力与可复用基础设施
* `web/` 作为管理端前端工程与构建产物目录

`docs/dev/Tech.md` 中提出的 Domain / Application / Adapter 思路应被吸收为当前工程的设计原则，而不是直接替换现有目录基线。除非出现明确收益，不引入整套平行目录体系。

## 6. 架构原则

### 6.1 服务端分层

服务端继续遵循 MVC 扩展分层：

* Handler：请求解析、权限入口、响应封装
* Service：业务规则、流程编排、事务边界
* Model：数据模型、迁移与持久化表达
* Middleware：横切关注点
* Router：路由组织与挂载
* Pkg：解析器、渲染器、进程控制、日志流等跨模块可复用基础能力

### 6.2 统一产品模型

业务层统一使用以下产品概念：

* 配置源 `source_configs`
* 代理节点 `proxy_nodes`
* 节点测试结果 `node_test_results`
* 监听入口 / 端口配置 `port_profiles`
* 端口与节点绑定 `port_profile_nodes`
* 运行配置快照 `runtime_configs`
* 内核实例 `kernel_instances`

前后端、接口层和持久化层都应优先围绕统一概念展开，避免直接使用 Mihomo 私有字段作为主模型。

### 6.3 Runtime IR 与 Adapter

渲染运行配置时，必须先从数据库实体构建统一 Runtime IR，再由具体内核适配器完成目标配置格式输出。

要求：

* 不直接从前端 DTO 或数据库实体拼接 Mihomo YAML
* 渲染、启动、重载、状态读取通过内核适配器统一抽象
* 能力协商结果必须能驱动前端按钮、策略选项和降级提示

### 6.4 前后端同步

能力变更时，必须同步处理：

* 后端接口
* 前端页面
* 导航入口
* API client
* 类型定义
* Swagger
* 部署文档
* 配置项说明

### 6.5 数据库迁移显式化

数据库 schema 变更必须显式维护版本与迁移逻辑，迁移代码统一收敛在 `server/internal/model/migrate/`。

`AutoMigrate` 可以作为建表与补齐的底层手段，但不能替代版本化迁移方案本身。

迁移实现约束：

* `server/internal/model/migrate/` 下的公共迁移基础设施应集中维护，不按单个小方法零散拆文件
* 每个版本文件负责自身版本的升级步骤、执行顺序和校验逻辑
* `server/internal/model/main.go` 只负责迁移接入，不承载具体版本升级实现
* `server/internal/model/migrate/scheduler.go` 只负责通用调度，不承载业务化迁移步骤
* 升级过程按版本顺序逐级推进，遵循类似 Android 数据库升级的演进方式

### 6.6 安全默认

运行控制相关能力必须满足：

* 内核控制接口默认仅监听 `127.0.0.1`
* Secret 等敏感配置不得在日志或接口响应中明文回显
* 执行进程时必须使用参数数组，禁止拼接 shell
* 上传内容、渲染内容和运行目录必须受白名单与权限控制

## 7. 核心对象

当前长期保留的核心对象包括：

* `users`
* `files`
* `options`
* `app_logs`
* `source_configs`
* `proxy_nodes`
* `node_test_results`
* `port_profiles`
* `port_profile_nodes`
* `runtime_configs`
* `kernel_instances`

如后续引入能力缓存、标签、健康记录等对象，必须先确认其是否服务于当前主线。

## 8. 文档维护原则

以下内容变化时，应同步更新本文档：

* 产品范围或系统边界变化
* MVP / 非 MVP 边界变化
* 核心对象变化
* 核心架构与职责划分变化
* Runtime IR / Adapter 策略变化
* 服务端工程组织基线变化
* 数据库迁移策略变化
