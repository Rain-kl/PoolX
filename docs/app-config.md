# PoolX 配置项说明

本文档描述 `PoolX` 当前支持的服务端与前端配置项。

当前配置说明面向 PoolX 当前保留能力，不再把 Agent、OpenResty、观测分析等历史链路配置视为正式配置基线。

## 1. Server 配置

Server 支持三类配置来源：

1. 命令行参数
2. 环境变量
3. 数据库 `options` 表中的运行时配置

### 1.1 命令行参数

推荐通过 `server/cmd/server` 启动：

```bash
cd server
go run ./cmd/server --port 3000 --log-dir ./logs
```

如需使用嵌入式静态资源入口，也可执行：

```bash
cd server
go run . --port 3000 --log-dir ./logs
```

| 参数 | 作用 | 默认值 |
| --- | --- | --- |
| `--port` | 指定 Server 监听端口 | `3000` |
| `--log-dir` | 指定日志目录 | 空 |
| `--version` | 输出当前版本后退出 | `false` |
| `--help` | 输出帮助信息后退出 | `false` |

### 1.2 环境变量

| 环境变量 | 作用 | 默认值 |
| --- | --- | --- |
| `PORT` | Server 监听端口 | `3000` |
| `GIN_MODE` | Gin 运行模式 | 非 `debug` 时按 release |
| `LOG_LEVEL` | 日志等级 | `info` |
| `SESSION_SECRET` | Session 签名密钥 | 启动时随机生成 |
| `SQLITE_PATH` | SQLite 数据库文件路径 | `poolx.db` |
| `DSN` | PostgreSQL DSN，设置后优先于 SQLite | 空 |
| `SQL_DSN` | 兼容旧命名的 PostgreSQL DSN，优先级低于 `DSN` | 空 |
| `REDIS_CONN_STRING` | Redis 连接串 | 空 |
| `UPLOAD_PATH` | 上传目录 | `upload` |

说明：

* `DSN` 与 `SQL_DSN` 同时存在时，优先使用 `DSN`
* `SESSION_SECRET` 在生产环境必须显式配置
* `REDIS_CONN_STRING` 未配置时，相关能力退化为进程内实现
* 服务端升级默认从 `Rain-kl/PoolX` 查询发布版本，可通过运行时配置 `ServerUpdateRepo` 覆盖

### 1.3 运行时配置（Option）

以下配置由管理端维护，可热更新：

| 配置项 | 作用 | 默认值 |
| --- | --- | --- |
| `PasswordLoginEnabled` | 是否启用密码登录 | `true` |
| `PasswordRegisterEnabled` | 是否启用密码注册 | `true` |
| `EmailVerificationEnabled` | 是否启用邮箱验证码流程 | `false` |
| `RegisterEnabled` | 是否允许用户注册 | `true` |
| `ServerUpdateRepo` | 服务端版本检查与升级使用的 GitHub 仓库，格式为 `owner/repo` | `Rain-kl/PoolX` |
| `GeoIPProvider` | IP 归属解析方式 | `disabled` |
| `GitHubOAuthEnabled` | 是否启用 GitHub OAuth 登录 | `false` |
| `WeChatAuthEnabled` | 是否启用微信登录 | `false` |
| `TurnstileCheckEnabled` | 是否启用 Turnstile 校验 | `false` |
| `SMTPServer` | SMTP 服务地址 | 空 |
| `SMTPPort` | SMTP 端口 | `587` |
| `SMTPAccount` | SMTP 账号 | 空 |
| `SMTPToken` | SMTP 密钥或授权码 | 空 |
| `SystemName` | 系统名称 | `PoolX` |
| `Notice` | 系统公告 | 空 |
| `About` | 关于页内容 | 空 |
| `Footer` | 页脚文案 | 默认值 |
| `GlobalApiRateLimitNum` | API 限流次数 | `300` |
| `GlobalApiRateLimitDuration` | API 限流时间窗口 | `180` |
| `GlobalWebRateLimitNum` | Web 限流次数 | `300` |
| `GlobalWebRateLimitDuration` | Web 限流时间窗口 | `180` |
| `UploadRateLimitNum` | 上传接口限流次数 | `50` |
| `UploadRateLimitDuration` | 上传接口限流时间窗口 | `60` |
| `DownloadRateLimitNum` | 下载接口限流次数 | `50` |
| `DownloadRateLimitDuration` | 下载接口限流时间窗口 | `60` |
| `CriticalRateLimitNum` | 敏感接口限流次数 | `100` |
| `CriticalRateLimitDuration` | 敏感接口限流时间窗口 | `1200` |

说明：

* Token、Secret 一类敏感配置不会通过选项列表直接回显
* `ServerUpdateRepo` 默认值为 `Rain-kl/PoolX`，用于版本检查与自动升级；如使用自建发布仓库，可改为自己的 `owner/repo`
* 已移除业务的运行时配置项只允许在兼容清理时兜底处理，不应重新作为正式选项暴露
* 配置项如有增删，必须同步更新本文档

## 2. Frontend 构建环境变量

| 环境变量 | 作用 | 默认值 |
| --- | --- | --- |
| `NEXT_PUBLIC_API_BASE_URL` | 前端请求 API 的基础路径 | `/api` |
| `NEXT_PUBLIC_APP_VERSION` | 前端展示版本号 | `dev` |
| `NEXT_DEV_BACKEND_URL` | 本地开发代理到后端的地址 | `http://127.0.0.1:3000` |

## 3. 文档维护要求

以下内容变化时，必须同步更新本文档：

* 服务端命令行参数
* 服务端环境变量
* 运行时配置项
* 前端构建环境变量
* 任一配置项的默认值、用途或示例
