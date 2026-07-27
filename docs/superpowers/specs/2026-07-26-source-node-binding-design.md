# 配置源与节点绑定、级联删除与同步

**日期**：2026-07-26  
**状态**：已批准设计，待实现  
**范围**：`clash` 垂直切片 — `SourceConfig` ↔ `ProxyNode` 生命周期

## 背景

当前导入链路（上传 / 订阅 → parse → confirm）会在确认时写入 `proxy_nodes.source_config_id`，但：

1. 删除配置源**只删源行**，节点成为孤儿。
2. 不存在配置源「更新」路径：再次 confirm 对已 `imported` 源是 no-op；无 refresh / reupload。
3. 节点还被 `port_profile_nodes`、`node_test_results` 引用；删节点时未清理，易悬空。

## 目标

1. **绑定**：从配置源导入的节点必须绑定该源（`source_config_id` + `source_config_name`）。
2. **级联删除**：删除配置源时，一并删除其绑定节点及关联数据。
3. **同步更新**：配置源内容更新时，对该源节点采用**全删全插**同步。

## 非目标

- 物理外键 / `ON DELETE CASCADE`（项目约定无物理 FK）。
- 软删除与可恢复回收站。
- 改绑 / 抢占其它源已拥有的 fingerprint。
- 为 `fingerprint` 加 DB 唯一约束（可后续单独做）。
- 端口 profile 运行配置自动重渲染（删绑定后由现有编辑/重渲染路径处理）。

## 决策摘要

| 项 | 决策 |
| --- | --- |
| 实现路径 | 应用层事务级联（路径 A），对齐 port-profile 的 TX 模式 |
| 绑定字段 | 沿用 `proxy_nodes.source_config_id` / `source_config_name` |
| 去重 | 全局 fingerprint 唯一语义：本源同步时**跳过**已被其它源占用的节点 |
| 同步触发 | 再次 Confirm、订阅 Refresh、上传 Reupload — 三者共用 `SyncSourceNodes` |
| 删节点连带 | `port_profile_nodes` → `node_test_results` → `proxy_nodes` |

## 数据与所有权

### 绑定

- 节点「属于」`source_config_id` 对应的 `source_configs` 行。
- 导入 / 同步插入时必须写入 `source_config_id` 与当时的 `source_config_name`（filename 快照）。
- `source_config_id = 0` 的节点视为无源绑定，**不在**本设计的级联 / 同步范围内。

### 去重（全局 fingerprint）

- Fingerprint 由 parser 对 `type` / `server` / `port` 及部分鉴权字段计算（现有逻辑不变）。
- 池内同一 fingerprint 只应保留一条业务记录（应用层保证；本期不加唯一索引）。
- 同步本源时：先删除本源全部节点，再对解析结果查**仍存在于池中的** fingerprint；命中则计为 `duplicate_nodes` 并跳过，不改绑其它源。

### 关联清理顺序（删除一批节点时）

在同一事务内，对目标 `node_ids`：

1. 删除 `port_profile_nodes` where `proxy_node_id IN (...)`
2. 删除 `node_test_results` where `node_id IN (...)`
3. 删除 `proxy_nodes` where `id IN (...)`（或 `source_config_id = ?`）

## 核心用例

### `SyncSourceNodes(ctx, sourceID, rawContent, meta)`

配置源内容已就绪（内存中的 raw）后的共享内核：

```
BEGIN
  load SourceConfig by id (FOR UPDATE if supported / 否则先 load 再写)
  parse rawContent → proxies, stats
  list node IDs where source_config_id = sourceID
  cascade-delete those nodes (bindings + tests + nodes)
  for each valid proxy:
    if fingerprint exists in pool → skip (duplicate)
    else collect for insert with SourceConfigID / Name
  CreateBatch insertables
  update SourceConfig: raw_content, content_hash, stats, status=imported, updated_at
COMMIT
return stats { total, valid, invalid, duplicate, imported }
```

说明：

- 必须先删本源节点再查全局 fingerprint，避免「自己与自己冲突」。
- 文件内重复 fingerprint 仍用内存 map 跳过（现有行为）。
- `meta` 可带 `filename` / `source_url` / `content_type` / `fetched_at` 等需回写字段。

### 删除配置源

```
BEGIN
  list node IDs by source_config_id
  cascade-delete those nodes
  delete source_configs row
COMMIT
```

源不存在 → 404（或与现有 Delete 语义一致）。

### 触发同步的入口

| 入口 | 条件 | 行为 |
| --- | --- | --- |
| `ConfirmSourceConfig` | 任意 status（含已 `imported`） | 使用库内已有 `raw_content` 调用 `SyncSourceNodes`；**取消**「已 imported 则 early return」 |
| `RefreshSourceConfig` | `source_type = subscription_url` | HTTP 拉取 `source_url` → 更新 raw 相关字段 → `SyncSourceNodes` |
| `ReuploadSourceConfig` | `source_type = upload` | 请求体 YAML → 更新 raw → `SyncSourceNodes` |

类型不匹配 → 400。拉取订阅失败 → 映射为上游/网关错误（与现有 `FetchSubscription` 一致）。

首次从 `parsed` 确认：与同步同一路径（全删本源节点通常为空集 + 插入），行为与「首次导入」等价，实现统一。

## API

前缀：`/api/v1/admin/clash`（与现有一致）。

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/source-configs/:id/confirm` | 改为始终同步（全删全插） |
| `POST` | `/source-configs/:id/refresh` | **新增**；仅 subscription |
| `POST` | `/source-configs/:id/reupload` | **新增**；仅 upload；body 为 YAML（multipart 或 raw，对齐现有 upload） |
| `DELETE` | `/source-configs/:id` | 级联删节点 + 关联 |

同步类成功响应 `data` 建议：

```json
{
  "id": 1,
  "total_nodes": 100,
  "valid_nodes": 98,
  "invalid_nodes": 2,
  "duplicate_nodes": 5,
  "imported_nodes": 93
}
```

字段名 snake_case，信封 `{ "error_msg": "", "data": ... }`。

## 分层改动

### Domain

- 无新实体；可补充同步结果 DTO 若现有 confirm 结果类型不够用。

### Repository（接口 + relational）

新增能力（命名可微调，语义固定）：

- 按 `source_config_id` 列出 node IDs
- 按 `source_config_id` 或 `node_ids` 删除节点
- 按 `node_ids` 删除 `port_profile_nodes`
- 按 `node_ids` 删除 `node_test_results`
- 事务：在 repository 层提供 `WithTx` / 或由现有 DB 句柄在 service 侧开启事务并传入实现（**对齐 port-profile Delete / SetProfileNodes 既有模式**）

不引入物理外键迁移。

### Application (`application/clash`)

- 抽取 `SyncSourceNodes`
- 改 `ConfirmSourceConfig`、`DeleteSourceConfig`
- 新 `RefreshSourceConfig`、`ReuploadSourceConfig`
- 禁止在 handler 内拼 SQL 或开事务绕过 service

### Transport

- Handler 注册 refresh / reupload
- Delete 文案与错误码沿用 `response` 信封

### Frontend

- 删除配置源：确认文案说明将同时删除该源下全部节点（及端口绑定中的这些节点）
- 订阅源：提供「刷新」
- 上传源：提供「重新上传」
- 节点池按源筛选：API 已支持 `source_config_id`；本期建议加上，非阻塞后端

## 错误处理

| 情况 | 行为 |
| --- | --- |
| 源不存在 | 404 |
| refresh 用于 upload / reupload 用于 subscription | 400 |
| YAML 解析失败 | 400（与现有 upload 一致） |
| 订阅 HTTP 失败 | 与现有 FetchSubscription 相同映射 |
| 事务中途失败 | 整单回滚，不出现半删半插 |

## 测试计划

1. **级联删除**：导入源 A 若干节点 → 节点加入 port profile 并产生 test result → 删源 A → 断言节点 / 绑定 / 测试记录均无残留，源行不存在。
2. **Confirm 全删全插**：导入后改库内 raw（或 reupload）再 confirm → 本源节点集合与新内容一致；旧节点 ID 不应残留。
3. **跨源 duplicate**：源 A 已导入 fingerprint F → 源 B 同步含 F → B 的 `duplicate_nodes` 增加且不插入第二条 F；A 的节点仍在。
4. **Refresh 类型守卫**：upload 源调 refresh → 400。
5. **Reupload 类型守卫**：subscription 源调 reupload → 400。
6. **自冲突避免**：已 imported 源再次 confirm 相同内容 → 全删后重插，`imported_nodes` 合理（不为 0 仅因「自己挡住自己」）。

## 风险与缓解

| 风险 | 缓解 |
| --- | --- |
| 全删全插导致 port profile 丢节点 | 产品预期；删除确认文案明确；后续可做「同步后提示受影响 profile」 |
| 测试结果历史丢失 | 同步即新节点 ID，历史测速清空可接受 |
| 大订阅 TX 过长 | 先实现正确性；若实测超时再拆批（非本期） |
| 漏清关联表 | 测试强制覆盖三表；集中一个 cascade helper |

## 实现顺序建议

1. Repository 批量删除 + TX 能力  
2. `SyncSourceNodes` + 改造 Confirm / Delete  
3. Refresh / Reupload API  
4. 测试  
5. 前端操作与文案  

## 验收标准

- [ ] 删除配置源后，该 `source_config_id` 下无 `proxy_nodes`，且无指向这些节点的 `port_profile_nodes` / `node_test_results`
- [ ] Confirm / Refresh / Reupload 后本源节点与最新 raw 解析结果一致（全局 duplicate 跳过除外）
- [ ] 跨源同 fingerprint 不产生第二条，也不改绑
- [ ] `cd backend && go test ./...` 通过相关 clash 包测试
