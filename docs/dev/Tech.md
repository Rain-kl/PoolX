

设计文档
Clash / Mihomo 代理池控制端架构设计（多内核扩展版）
项目主题：Clash / Mihomo 代理池控制端，兼容多内核扩展
版本：v1.0（整合稿）
定位：Proxy Kernel Control Plane

说明：本稿已将“Clash 控制端”与“多内核统一控制平台”两份内容合并。当前版本以 Mihomo 为首个可落地内核，同时保留 sing-box / Xray 的扩展抽象。


1. 架构目标
• 首期完成 Mihomo 运行控制闭环：解析、测试、渲染、启动、停止、重载、状态查询。
• 通过统一领域模型、Runtime IR 与 Kernel Adapter 支撑 sing-box / Xray 扩展。
• 保证前端、应用层与运行内核解耦，避免业务主链路绑定某个配置格式。
2. 技术栈
层级	技术
前端	Next.js + React + TypeScript + HeroUI/NextUI + react-hook-form + zod + Zustand/TanStack Query
后端	Go + Gin/Fiber + GORM + PostgreSQL + Redis + Viper/koanf + yaml.v3
运行控制	os/exec + WebSocket/SSE + 文件系统 + 进程健康检查
内核	Mihomo（MVP），后续 sing-box / Xray
预览	Monaco Editor 用于 YAML/JSON 高级预览

3. 分层架构
层	职责
Domain	定义 Node、ListenerProfile、PoolStrategy、WorkspaceRuntime 等统一模型
Application	实现 ImportConfig、TestNodes、RenderRuntimeConfig、StartKernel 等用例
Kernel Adapter	把统一模型翻译为 Mihomo / sing-box / Xray 的配置、进程和状态接口
Infrastructure	数据库、缓存、文件系统、进程管理、日志、SSE/WebSocket

4. 统一领域模型
• ProxyNode：统一节点对象，包含协议、地址、端口、原始配置与来源元数据。
• ListenerProfile：监听入口，定义 listen、port、type、enabled。
• PoolStrategy：产品语义策略，包含 random、load_balance、failover、latency_first、manual。
• PortRuntimeProfile：一个端口运行单元，绑定监听入口、节点集合与策略。
• WorkspaceRuntime：工作台运行时聚合对象，记录目标内核、模式、端口集合与节点集合。
5. Runtime IR 设计
为了避免直接从数据库实体或前端 DTO 渲染 Mihomo YAML，设计统一中间表示 Runtime IR。不同内核只关心如何把 IR 映射为自己的配置结构。
IR 对象	作用
IRListener	统一监听入口表示
IRNode	统一节点表示，保留 raw 配置
IRPool	统一节点池与策略表示
IRBinding	描述 listener 与 pool 的绑定关系
IRHealthCheck	统一健康检查参数
RuntimeIR	组合 listeners / nodes / pools / bindings / health-check

6. 设计模式与接口
模式	用途
Strategy	不同内核的配置渲染与进程启动实现
Abstract Factory	按内核类型生成 Renderer / Launcher / Reloader / StatusReader
Adapter	屏蔽 Mihomo、sing-box、Xray 私有 API 与配置差异
Template Method	统一“校验 → 构建 IR → 渲染 → 写文件 → 生成哈希”的流程
Builder	构建 Runtime IR，避免直接拼接 YAML/JSON

7. 数据模型设计
实体	关键字段
SourceConfig	user_id, raw_yaml, source_type, parse_status, parse_error
ProxyNode	source_config_id, type, server, port, raw_json, fingerprint_hash, enabled
NodeTestResult	node_id, test_type, success, latency_ms, error_message, tested_at
PortProfile	listen, port, inbound_type, strategy_type, test_url, test_interval, enabled
PortProfileNode	port_profile_id, node_id, order_index, weight, enabled
RuntimeConfig	rendered_yaml/json, version, checksum, status, applied_at
CoreInstance / KernelInstance	process_id, status, api_addr, secret, started_at, stopped_at

8. 配置渲染与内核映射
• Mihomo：IR → YAML；listeners、proxies、proxy-groups、rules；random 映射为 load-balance + round-robin。
• sing-box：IR → JSON；manual 可映射 selector，latency_first 可映射 urltest，其他策略可能部分支持。
• Xray：IR → JSON；依赖 routing、balancer、observatory 组合实现策略。
• 所有渲染器输出 RenderedConfig：kernelType、format、content、checksum。
9. 运行器与热重载设计
组件	设计要点
KernelProcessSpec	统一描述 BinaryPath、Args、Env、WorkDir、ConfigPath
ProcessLauncher	按内核实现 BuildSpec / Start / Stop
Reloader	声明 ReloadMode：API / Signal / Restart
StatusReader	读取运行状态、端口、版本、健康信息
HealthChecker	检查 API 可达性、配置应用状态与进程健康

10. 统一 API 设计
分类	接口
能力发现	GET /api/v1/kernels
工作区内核切换	PUT /api/v1/workspaces/{id}/kernel
渲染与导出	POST /api/v1/workspaces/{id}/render；POST /api/v1/workspaces/{id}/export
运行控制	POST /api/v1/runtime/start / stop / reload；GET /api/v1/runtime/status
日志	GET /api/v1/runtime/logs；SSE/WebSocket 推送
导入与节点	复用 source-configs / nodes / port-profiles 相关 REST API

11. 安全设计
• Mihomo external-controller 仅监听 127.0.0.1，使用 secret。
• 前端只访问 Go API，禁止直接暴露内核控制 API 到公网。
• 上传文件需限制大小、层级和允许字段，原始 YAML 不直接下发内核。
• 启动进程使用参数数组，严禁用户输入拼接 shell。
• 运行配置目录、日志目录与二进制执行权限需隔离。
12. 性能与可观测性
• 节点测试采用并发队列与 Redis 缓存，结果缓存 1~5 分钟。
• 运行配置保存 checksum，未变化不重载。
• 节点列表分页，测试历史单独查询或聚合缓存。
• 观测指标至少包括节点总数、可用节点数、端口绑定数、最近成功率、重载次数、进程存活状态。
14. 实施建议
• 第一阶段先定义统一模型与接口：RuntimeIR、KernelFactory、Renderer、Launcher、Reloader、CapabilityProvider。
• 第二阶段只实现 Mihomo Parser / Renderer / Launcher / Reloader / StatusReader。
• 第三阶段前端按能力协商渲染，不写死 Mihomo。
• 第四阶段逐步接入 sing-box、Xray。
设计结论：采用“统一领域模型 + Runtime IR + Kernel Adapter + Capability Discovery + Unified Runtime API”是本项目兼顾首期交付与长期扩展的最优结构。