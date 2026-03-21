# AGENTS.md

本文件是当前项目的 AI 接手入口。

当前项目不再以“纯模板工程维护”为唯一主线，而是以 `GinNextTemplate` 为底座，持续开发 Proxy Kernel Control Plane 业务。后续默认开发目标是在保留模板基础设施能力的前提下，交付代理池控制端的业务闭环。

本文件不承载详细设计、规范和计划。接手项目时，先按顺序阅读以下文档：

1. [docs/design.md](./docs/design.md)
   作用：理解当前产品范围、系统边界、核心对象和整体架构。

2. [docs/development-guidelines.md](./docs/development-guidelines.md)
   作用：理解当前开发规范，包括技术基线、分层约束、数据模型边界、API 约定和测试要求。

3. [docs/development-plan.md](./docs/development-plan.md)
   作用：理解当前阶段的开发重点、默认迭代顺序和准入规则。

4. [docs/frontend-development-guidelines.md](./docs/frontend-development-guidelines.md)
   作用：理解前端技术选型、目录分层、组件规范、请求层、状态管理、样式和测试约束。

5. [docs/deployment.md](./docs/deployment.md)
   作用：理解当前部署方式、启动步骤和最小验证方法。

6. [docs/app-config.md](./docs/app-config.md)
   作用：理解系统支持的环境变量、命令行参数和运行时配置项。

## 当前主线

当前项目默认主线是 Proxy Kernel Control Plane 业务开发：

* 保留并持续复用模板基础能力：用户、邮箱、文件上传、安全、系统设置、应用日志、服务端升级
* 新增并持续演进业务能力：配置导入、节点池、节点测试、工作台配置、运行控制、能力协商
* 服务端工作重心集中在 `server`
* 服务端继续按 MVC 思路演进，避免职责回流
* 前后端能力保持同步演进，避免残留失效入口
* 工程结构围绕当前仓库基线演进，不为了概念图而平移目录

当前不属于 MVP 主线的能力：

* 多租户平台化
* 多实例并行运行
* 分布式 Agent 集群
* 对外暴露内核控制 API
* 复杂规则编辑器

## 当前工程基线

当前服务端目录基线位于 `server/`，主要结构如下：

* `server/cmd/server`：服务启动入口
* `server/internal/app`：启动装配
* `server/internal/handler`：接口层
* `server/internal/service`：业务逻辑层
* `server/internal/model`：数据模型与迁移
* `server/internal/middleware`：中间件
* `server/internal/router`：路由注册
* `server/internal/pkg`：项目内公共能力与可复用基础设施
* `server/web`：前端工程与静态构建产物

如果后续确有收益，可以继续演进出 `repository/`、`dto/` 或顶层 `pkg/`，但不为了形式而迁移。

## 执行要求

* 如果实现内容超出当前产品边界，先修改 `docs/design.md`，再同步更新相关规范与计划文档后继续编码。
* 如果 `docs/design.md`、`docs/development-plan.md` 与当前实现冲突，优先以 `docs/design.md` 和当前已确认业务主线为准，并补齐相关文档。
* 如果实现方式违反 `docs/development-guidelines.md`，应优先调整方案，而不是绕过规范。
* 如果任务涉及前端改造或管理端 UI，必须同时阅读 `docs/frontend-development-guidelines.md`。
* 如果任务涉及删除或替换能力，接口与界面必须同步处理；同时检查后端路由、模型、前端入口、导航、API client、类型定义、Swagger、部署文档和配置项，避免出现残留入口。
* 新增业务能力应优先复用现有通用基础设施，不要绕开既有用户、设置、日志、上传、安全能力重复造轮子。
* 服务端开发必须遵循当前分层：`handler` 负责请求处理和响应封装，`service` 负责业务逻辑与流程编排，`model` 负责数据模型与迁移，`middleware` 负责横切能力，`router` 负责路由组织。
* Runtime IR、解析器、渲染器、启动器、状态读取器等抽象应吸收到当前工程分层中，不直接照搬 `docs/dev/Tech.md` 的推荐目录树。
* 如果任务涉及数据库 schema 变更，迁移基础设施统一收敛在 `server/internal/model/migrate/` 下的公共实现中，具体版本升级步骤、执行顺序和校验逻辑必须内聚在 `server/internal/model/migrate/v*.go` 中。
* 数据库版本升级实现应借鉴 Android 的逐级升级思想：从旧版本按 `vN -> vN+1` 顺序依次执行，而不是在入口层拼接一次性迁移。
* 新增的方法、类型、函数、结构和模块，必须放在符合其功能定义的包下；开发时应始终优先考虑代码仓库整洁度与可维护性，禁止把任意代码写进职责含糊的包里。
* 目录调整属于允许范围，但必须围绕当前基线演进，不要重新扩张旧式平铺结构。

## 文档维护要求

当以下内容发生变化时，应同步更新对应文档：

* 产品范围或系统边界变化：更新 `docs/design.md`
* 开发约束、代码规范、接口约定变化：更新 `docs/development-guidelines.md`
* 当前开发重点、默认顺序、准入规则变化：更新 `docs/development-plan.md`
* 前端目录分层、组件规范、样式体系、测试基线变化：更新 `docs/frontend-development-guidelines.md`
* 产品启动配置、部署方式或联调步骤变化：更新 `docs/deployment.md` 和 `README.md`
* 环境变量、命令行参数或运行时配置项变化：更新 `docs/app-config.md`
