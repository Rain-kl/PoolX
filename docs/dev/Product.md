

产品文档
Clash / Mihomo 代理池控制端产品方案（兼容多内核）
项目主题：Clash / Mihomo 代理池控制端，兼容多内核扩展
版本：v1.0（整合稿）
定位：Proxy Kernel Control Plane

说明：本稿已将“Clash 控制端”与“多内核统一控制平台”两份内容合并。当前版本以 Mihomo 为首个可落地内核，同时保留 sing-box / Xray 的扩展抽象。


1. 产品概述
本产品是一个面向代理池场景的 Web 控制台。首期以 Mihomo 为运行内核，为用户提供导入节点、测试节点、配置端口、生成配置、控制进程和查看状态的一站式体验。产品层统一使用“节点、监听入口、策略、运行配置”等概念，不直接暴露内核专有术语。
2. 产品原则
• 以代理池为中心：面向全局代理与工作台配置，不做复杂规则编辑。
• 配置可视化：普通用户无需直接编辑 YAML/JSON，仅在高级预览中查看。
• 控制与内核解耦：前端不直连内核 API，所有操作走后端统一接口。
• 能力协商驱动 UI：界面根据目标内核能力自动启用、禁用或降级功能。
• 安全默认：控制接口只在本机开放，secret 与权限隔离默认开启。
3. 典型用户旅程
阶段	用户动作	系统反馈
导入	上传配置文件	展示解析摘要、节点列表、无效项和去重结果
验证	执行批量测试	实时返回进度、延迟、成功率与错误原因
建池	确认导入节点池	节点进入统一池，可筛选和批量操作
配置	创建端口并选择节点、策略	形成工作台配置卡片并可即时预览
运行	保存并启动核心	展示端口状态、进程状态、API 健康与日志
优化	调整节点或策略并重载	生成新快照，哈希变化时执行热更新

4. 信息架构
页面	核心内容	关键组件
Dashboard	运行状态、在线端口数、节点数、成功率、最近日志	CoreStatusCard / MetricCard / LogSnippet
配置导入页	上传、解析表格、测试、确认导入	NodeSourceImporter / NodeImportTable / NodeTestPanel
节点池页	节点列表、筛选、测试结果、批量操作	NodePoolTable / FilterBar / BulkActionBar
工作台页	节点池、端口配置画布、详情与预览	PortProfileEditor / StrategySelector / RuntimeConfigPreview
运行状态页	进程状态、监听端口、API 健康、日志流	CoreStatusCard / PortStatusTable / LogStreamPanel
系统设置页	二进制路径、默认测试参数、日志级别、API secret	SettingsForm / CapabilityBadge

5. 核心模块方案
5.1 配置导入
支持上传 YAML，系统解析 proxies 后做标准化和双层去重。导入前可先测试节点并预览结果，导入后形成统一节点池。
5.2 节点池
节点池是整个产品的基础资源层，支持搜索、筛选、启用/禁用、查看最近测试结果与来源。未来可增加标签和智能分组。
5.3 工作台
工作台是核心页面。每个端口配置卡片包含名称、监听地址、端口、入站类型、节点集合、策略、测试 URL、间隔和启用状态。
5.4 配置预览
普通用户默认不看原始配置；高级用户可打开 YAML/JSON 预览抽屉，用于排障和对比。
5.5 运行控制
统一展示启动、停止、重载、运行状态、日志与异常。热重载是否可用由目标内核能力决定。
6. 策略产品语义
产品术语	用户理解	Mihomo 首期映射	备注
随机	请求在节点池中轮转分配	load-balance + round-robin	不单独造轮子
负载均衡	尽量均匀使用节点	load-balance + consistent-hashing / round-robin	根据配置选择算法
故障转移	主节点失败时自动切换	fallback	需配置 test URL 与间隔
延迟优先	优先选择低延迟节点	url-test 或等价自动选优方案	不同内核可能近似实现
手动	由用户显式选择出口	select（后续）	MVP 可预留

7. 能力协商与多内核体验
• 前端顶部展示当前目标内核、配置格式、支持的策略、是否支持热重载/外部 API/健康检查。
• 当某策略在当前内核中只有近似实现时，界面提示“近似实现”或“部分支持”。
• 当前端切换内核时，工作台保留统一概念，不直接出现 Mihomo、sing-box、Xray 的低层字段。
• 导入源与运行内核解耦：可导入 Clash/Mihomo YAML，但仍可输出为未来其他内核配置。
8. 关键交互细节
• 导入页解析完成后先展示摘要，再展示节点表，避免用户面对长列表无上下文。
• 节点测试应提供总体进度、成功/失败数量、平均延迟与失败原因分布。
• 工作台左侧节点池支持搜索和多选，中间端口卡片支持顺序调整，右侧展示明细与预览。
• 保存按钮与启动按钮分离：保存生成快照，启动/重载属于运行操作。
• 日志流默认展示最近窗口，支持按关键字过滤错误与警告。
9. API 视角的产品接口
接口分组	主要接口
导入	POST /source-configs/upload；POST /source-configs/{id}/confirm
节点	GET /nodes；POST /nodes/test；GET /nodes/{id}/test-results
工作台	GET/POST/PUT/DELETE /port-profiles；POST /port-profiles/{id}/nodes
渲染	POST /workspaces/{id}/render；POST /workspaces/{id}/export
运行	POST /runtime/start；POST /runtime/stop；POST /runtime/reload；GET /runtime/status
内核能力	GET /kernels；PUT /workspaces/{id}/kernel

10. 发布策略
• MVP 先完成单实例单配置的 Mihomo 版本，保证功能闭环与排障能力。
• 第二阶段完善能力协商，让前端不写死 Mihomo 特性。
• 第三阶段新增 sing-box 适配器，验证 IR 与 Adapter 的扩展价值。
• 后续再评估 Xray 与多工作区/多实例隔离。
产品结论：产品层坚持“统一概念、按能力协商、MVP 先 Mihomo”，既保证首期可交付，也避免未来扩多内核时推倒重来。