# Design: Docker 构建时打包最新 mihomo 内核

**日期**：2026-07-27  
**状态**：已批准  
**范围**：Dockerfile 预装 mihomo + `FOAM_CLASH_MIHOMO_BINARY_PATH` 环境变量覆盖默认二进制路径

## 背景

当前镜像只包含 `poolx` 与前端静态资源。mihomo 需在运行时通过 API 下载到 `./data/core/mihomo`。  
`data` 目录通常挂卷，且首次启动没有内核，`AutoStartKernel` 会因二进制不存在而跳过。

目标：在 `docker build` 时自动拉取 MetaCubeX/mihomo **最新正式 release**，写入镜像固定路径，容器开箱可用。

## 目标

1. `docker build` 按目标架构下载并解压最新 mihomo，装入镜像。
2. 镜像内路径固定为 **`/opt/mihomo`**（可执行文件），**不**进入 data 卷。
3. 镜像 `ENV FOAM_CLASH_MIHOMO_BINARY_PATH=/opt/mihomo`，**compose 无需再配置**。
4. 非 Docker / 本地默认仍为 `./data/core/mihomo`。

## 非目标

- 不 pin 固定 mihomo 版本（始终 latest release）。
- 不修改 `docker-compose.yml` 增加该 env。
- 不改前端 UI。
- 不在 entrypoint 启动时下载内核。
- 不把内核放进 `/app/data` 或挂载卷路径。

## 架构

### 1. Dockerfile：mihomo-downloader stage

新增独立 stage（Alpine + curl/ca-certificates/gzip）：

1. 调用 `https://api.github.com/repos/MetaCubeX/mihomo/releases/latest` 取得 `tag_name` 与 assets。
2. 按 `TARGETARCH` 选择 `.gz` 资源（与现有 Go installer 偏好一致、偏兼容）：

| TARGETARCH | 资产优先级 |
| --- | --- |
| amd64 | `mihomo-linux-amd64-compatible-*.gz` → `mihomo-linux-amd64-v1-*.gz` → 通用 `mihomo-linux-amd64-*.gz`（排除 deb/rpm/pkg、sha256） |
| arm64 | `mihomo-linux-arm64-*.gz` |

3. 下载、`gzip -dc` 解压到 `/out/mihomo`，`chmod 0755`。
4. API / 下载 / 无匹配 arch 失败时 **stage 失败**，不产出无内核镜像。

Final stage：

```dockerfile
COPY --from=mihomo-downloader --chmod=0755 /out/mihomo /opt/mihomo
ENV FOAM_CLASH_MIHOMO_BINARY_PATH=/opt/mihomo
```

### 2. Go：环境变量覆盖默认路径

新增常量：

- `FOAM_CLASH_MIHOMO_BINARY_PATH`（`backend/internal/infra/config/env.go`）

统一默认路径解析（推荐放在 `kernel` 包或 clash application 内小 helper）：

```text
DefaultMihomoBinaryPath() =
  trim(env FOAM_CLASH_MIHOMO_BINARY_PATH) if non-empty
  else "./data/core/mihomo"
```

替换当前硬编码 `defaultBinaryPath = "./data/core/mihomo"` 的使用点：

| 位置 | 行为 |
| --- | --- |
| `application/clash` 空 `BinaryPath` / Inspect / Upload / Download 回退 | 使用 `DefaultMihomoBinaryPath()` |
| `application/clash.AutoStartKernel` | 检查与启动路径使用同一默认 |
| `application/settings` 空 `MihomoBinaryPath` | 默认填入 `DefaultMihomoBinaryPath()` |

优先级（从高到低）：

1. 请求显式 `binary_path` / DB 中已保存的 settings 路径  
2. `FOAM_CLASH_MIHOMO_BINARY_PATH`（仅当上述为空时作为代码默认）  
3. `./data/core/mihomo`

说明：DB 已保存的相对路径（如历史 `./data/core/mihomo`）仍优先生效；仅「字段为空」时走 env/代码默认。Docker 新部署无自定义设置时会落到 `/opt/mihomo`。

不强制把该字段写入 `config.yaml` schema；env 在进程级覆盖「代码默认」即可。若现有 `applyEnvOverrides` 模式更易扩展，可把路径挂到 config 结构，但 **AutoStart / StartKernel 默认必须读到同一值**。

### 3. compose / 卷

- `docker-compose.yml`：**不**增加 `FOAM_CLASH_MIHOMO_BINARY_PATH`。
- 卷 `./data/poolx-data:/app/data` 不变；`/opt/mihomo` 在镜像层，不被卷覆盖。

### 4. 缓存

- 默认不强制 cache-bust；需要强制拉最新内核时使用 `docker build --no-cache`（或仅重建 downloader stage）。
- 可选后续：`ARG MIHOMO_CACHEBUST`；本设计不强制实现。

## 数据流

```text
docker build
  → mihomo-downloader: GitHub latest → /out/mihomo
  → final: COPY → /opt/mihomo
  → ENV FOAM_CLASH_MIHOMO_BINARY_PATH=/opt/mihomo

container start
  → poolx 读 env → DefaultMihomoBinaryPath() = /opt/mihomo
  → AutoStartKernel / StartKernel 使用该路径
  → workdir 仍为 ./data/runtime（在 data 卷内）
```

## 错误处理

| 阶段 | 行为 |
| --- | --- |
| Build：GitHub API / 下载失败 | Dockerfile RUN 非零退出 |
| Build：无匹配 arch asset | 明确错误信息后失败 |
| Runtime：路径不存在 | AutoStart 打日志并跳过（现有行为） |
| Runtime：用户 UI 下载/上传 | 可写到其他路径并写入 settings（现有能力） |

## 验证

1. `docker build` 成功（至少 linux/amd64；有条件时 arm64）。
2. 容器内：`test -x /opt/mihomo && /opt/mihomo -v` 有版本输出。
3. 容器内：`printenv FOAM_CLASH_MIHOMO_BINARY_PATH` 为 `/opt/mihomo`。
4. Go 测试：设置/不设置 `FOAM_CLASH_MIHOMO_BINARY_PATH` 时默认路径正确；`go test` 相关包通过。
5. compose 文件 diff 中 **无** 该 env 行。

## 实现触点（预期文件）

- `Dockerfile` — mihomo-downloader stage + final COPY/ENV  
- `backend/internal/infra/config/env.go` — env 常量（及可选 override）  
- `backend/internal/infra/clash/kernel/` 或 `application/clash` — `DefaultMihomoBinaryPath()`  
- `backend/internal/application/clash/service.go` — 默认路径与 AutoStart  
- `backend/internal/application/settings/service.go` — 空路径默认  
- 相关单测  

## 决策摘要

| 项 | 选择 |
| --- | --- |
| 版本策略 | 始终 latest release |
| 镜像路径 | `/opt/mihomo` |
| 卷 | 不挂载内核；data 仅运行数据 |
| 路径配置 | 镜像 `ENV` 固定；compose 不写 |
| 代码默认 | 本地仍 `./data/core/mihomo`；env 可覆盖 |
| 方案 | A：Dockerfile 预装 + FOAM_ env |
