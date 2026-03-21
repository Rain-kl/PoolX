# Proxy Kernel Control Plane 前端开发规范

本文档约束 `server/web` 的正式前端工程。当前前端以后台管理端为载体，后续开发默认围绕代理池控制端业务页面、通用组件和 feature 模块展开。

## 1. 技术基线

默认技术栈：

* Next.js 15 App Router
* React 19
* TypeScript 5
* Tailwind CSS 4
* TanStack Query
* React Hook Form + Zod
* Zustand
* ESLint + Prettier
* Vitest + Testing Library + Playwright
* pnpm

要求：

* 默认使用 TypeScript
* 默认使用函数组件
* 默认使用 App Router
* 前端必须支持 `light`、`dark`、`system` 三种主题模式

禁止：

* 引入大型 UI 框架破坏现有组件基线
* 使用 jQuery 风格 DOM 操作
* 让页面直接访问内核私有控制 API

## 2. 当前目录基线

当前前端目录按以下职责组织：

```text
server/web/
├── app/
├── components/
├── features/
├── lib/
├── store/
├── styles/
├── tests/
└── types/
```

职责约束：

* `app/`：路由、布局、页面装配
* `features/`：按业务域组织页面逻辑与交互
* `components/`：跨 feature 复用组件
* `lib/`：请求客户端、环境变量、工具函数、常量、流式连接封装
* `store/`：少量跨页面 UI 状态
* `types/`：共享类型定义
* `tests/`：单元与端到端测试

## 3. 页面与导航约束

当前主线页面包括：

* 公共页面：登录、注册、密码重置、关于页
* 管理页面：Dashboard、配置导入、节点池、工作台、运行状态、系统设置、日志、用户

要求：

* 页面文件只负责获取路由参数、组织结构、挂载 feature 组件
* 导航配置必须与真实页面、权限和后端接口保持一致
* 删除或新增页面时，必须同步更新导航、API client、类型与测试
* 页面术语优先使用统一产品概念，不直接暴露 Mihomo 私有字段

## 4. 数据请求与状态管理

### 4.1 请求层

所有 API 请求必须统一经过 `lib/api/`。

要求：

* 统一处理 `success/message/data` 响应结构
* 统一处理鉴权失效、网络异常和通用错误消息
* 统一维护资源接口与请求路径
* SSE / WebSocket 连接也必须经过统一封装，不在页面组件中零散创建

禁止：

* 在页面组件中直接调用裸 `fetch('/api/...')`
* 在多个组件中重复拼接同一接口路径
* 在浏览器中直接请求本机 Mihomo external-controller

### 4.2 状态分层

* 服务端状态：TanStack Query
* 页面临时状态：组件内部 `useState`
* 跨页面 UI 状态：Zustand
* 实时任务进度：统一封装的 SSE / WebSocket hook 或 client

不推荐：

* 用 Zustand 保存服务端主数据
* 用 Context 代替完整数据层方案

### 4.3 类型

要求：

* 开启 TypeScript 严格模式
* 禁止滥用 `any`
* API 响应、表单输入、业务实体、能力协商结果必须有明确类型

## 5. 表单与交互

统一使用：

* React Hook Form
* Zod

要求：

* 表单校验规则应与接口约束一致
* 高风险操作必须提供二次确认
* 成功、失败、加载、空态都要有清晰反馈
* 上传、解析、测试、渲染、启动、重载等长链路操作必须有阶段状态提示

## 6. 样式与主题

样式原则：

* 统一使用 Tailwind CSS 与现有 token 体系
* 优先复用现有基础组件、布局组件和反馈组件
* 保持视觉层级、留白、语义颜色和信息密度一致

主题要求：

* 同时支持 `light`、`dark`、`system`
* 用户选择必须持久化
* 首屏尽量避免主题闪烁

## 7. MVP 页面交付要求

MVP 阶段页面至少覆盖：

* Dashboard：运行状态、节点统计、端口统计、最近日志
* 配置导入页：上传、解析摘要、节点表、测试面板、导入确认
* 节点池页：分页、筛选、启用禁用、最近测试结果
* 工作台页：端口配置 CRUD、节点选择、策略选择、可合并片段预览
* 运行状态页：进程状态、监听端口、健康检查、日志流
* 设置页：内核路径、默认测试参数、日志级别、secret 等业务设置

## 8. 测试与交付

每个正式页面至少具备：

* 加载态
* 空态
* 错误态
* 成功反馈

测试要求：

* 公共工具、环境变量解析、主题逻辑补单元测试
* 关键页面交互补组件测试
* 导入、节点筛选、工作台保存、运行控制等核心主链路补 Playwright 或等效联调验证

交付要求：

* `pnpm build` 能生成可被 Go Server 托管的静态产物
* 新页面默认通过亮色与暗色模式验收
* 页面与后端接口变更保持同步提交
