# 级别与上下文

关于日志级别语义、基于 context 的日志模式、性能考虑以及哪些内容不应出现在日志中的详细指导。

## 级别语义

### Debug

仅开发人员的诊断。生产中默认禁用。用于跟踪在开发或故障排查期间有帮助的内部状态：

```go
slog.Debug("cache lookup", "key", key, "hit", hit)
slog.Debug("parsed config", "fields", len(cfg.Fields))
slog.Debug("SQL query", "query", q, "args", args)
```

**何时使用**：内部状态转换、缓存行为、开发期间的详细请求/响应数据。

### Info

确认系统按预期运行的重要事件。这些应在生产中对理解系统行为有用：

```go
slog.Info("server started", "addr", addr, "version", version)
slog.Info("config loaded", "path", cfgPath, "env", env)
slog.Info("migration completed", "version", v, "elapsed_ms", elapsed)
slog.Info("user registered", "user_id", uid)
```

**何时使用**：启动/关闭、配置变更、重要业务事件、周期性健康摘要。

### Warn

发生了意外的事情，但系统已恢复或优雅降级。运维人员可能想要调查但不需要立即行动：

```go
slog.Warn("retry succeeded", "attempt", n, "endpoint", url)
slog.Warn("deprecated endpoint called", "path", r.URL.Path, "user_id", uid)
slog.Warn("rate limit approaching", "current", rate, "limit", max)
slog.Warn("fallback to default config", "err", err)
```

**何时使用**：最终成功的重试、弃用的代码路径、接近资源限制、回退行为。

### Error

操作失败并需要运维人员关注。系统无法完成请求或任务：

```go
slog.Error("payment failed", "err", err, "order_id", id, "amount", amt)
slog.Error("database connection lost", "err", err, "host", dbHost)
slog.Error("message processing failed", "err", err, "msg_id", msgID)
```

**何时使用**：影响用户的失败操作、丢失的连接、数据完整性问题、未恢复的外部服务故障。

**始终包含错误**：`slog.Error` 调用应始终带有包含实际错误值的 `"err"` 属性。

### 在 Warn 和 Error 之间选择

```
操作最终是否成功？
├─ 是（经过重试/回退后）→ Warn
└─ 否（调用者收到错误）→ Error
    ├─ 需要立即关注 → Error
    └─ 可以等到下次审查 → Warn
```

---

## 自定义详细级别

slog 级别是整数。在标准级别之间定义自定义子级别以实现细粒度控制：

```go
const (
    LevelTrace = slog.Level(-8)  // 低于 Debug
    LevelNotice = slog.Level(2)  // 在 Info 和 Warn 之间
)

slog.Log(ctx, LevelTrace, "detailed trace", "span_id", spanID)
```

使用 `HandlerOptions.Level` 配合 `slog.LevelVar` 在运行时控制最低级别。

---

## 基于 Context 的日志

### 模式 1：Context 中的日志器

在 context 中存储增强后的 `*slog.Logger`。每个中间件层添加自己的字段：

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := authenticate(r)
        logger := loggerFromCtx(r.Context()).With("user_id", userID)
        ctx := context.WithValue(r.Context(), loggerKey, logger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

**优点**：简单，与任何 handler 链配合使用。
**缺点**：需要纪律来始终使用 `loggerFromCtx`。

### 模式 2：显式日志器参数

将 `*slog.Logger` 作为函数参数与 context 一起传递：

```go
func processOrder(ctx context.Context, logger *slog.Logger, order *Order) error {
    logger.Info("processing order", "order_id", order.ID)
    // ...
}
```

**优点**：显式依赖，更易测试，无需 context 键。
**缺点**：每个函数签名中都有额外参数。

### 何时使用哪种

| 场景 | 推荐 |
|------|------|
| HTTP 处理器 / 中间件链 | Context 中的日志器 |
| 无 HTTP 依赖的库代码 | 显式参数 |
| 后台工作器 / 批处理任务 | 显式参数 |
| 深层调用链（5 层以上） | Context 中的日志器 |

---

## 性能考虑

### 使用 Enabled() 预检查

当日志级别被禁用时避免分配日志参数：

```go
// 开销大：参数始终被求值，即使 Debug 被禁用
slog.Debug("request details",
    "headers", fmt.Sprintf("%v", r.Header),
    "body", string(bodyBytes),
)

// 更好：禁用时完全跳过
if slog.Default().Enabled(ctx, slog.LevelDebug) {
    slog.Debug("request details",
        "headers", fmt.Sprintf("%v", r.Header),
        "body", string(bodyBytes),
    )
}
```

当参数构造开销大（格式化、序列化或读取数据）时，这很重要。对于简单属性（`slog.String`、`slog.Int`），开销可以忽略不计。

### 在热路径上使用 LogAttrs

`slog.LogAttrs` 避免了便捷方法（`slog.Info` 等）产生的 `[]any` 分配：

```go
// 标准——为键值对分配一个 []any
slog.Info("request handled", "method", r.Method, "status", code)

// 更快——类型化属性，无 []any 分配
slog.LogAttrs(ctx, slog.LevelInfo, "request handled",
    slog.String("method", r.Method),
    slog.Int("status", code),
)
```

### 避免在紧凑循环中记录日志

如果循环处理数千个项目，记录摘要而不是每次迭代：

```go
// 不好：10k 项目批次中每个项目一条日志
for _, item := range items {
    slog.Debug("processing item", "id", item.ID)
    process(item)
}

// 好：记录摘要
slog.Info("batch started", "count", len(items))
processed, failed := processBatch(items)
slog.Info("batch completed", "processed", processed, "failed", failed)
```

---

## 不应记录的内容

### 密钥和凭证

永远不要记录：
- 密码、API 密钥、令牌（OAuth、JWT、会话）
- 私钥、证书
- 包含凭证的数据库连接字符串

```go
// 不好
slog.Info("connecting", "dsn", dsn) // 可能包含密码

// 好
slog.Info("connecting", "host", dbHost, "database", dbName)
```

### 个人身份信息（PII）

除非调试所需且你的保留策略允许，否则避免记录：
- 电子邮件地址、电话号码
- 完整姓名、物理地址
- IP 地址（在某些司法管辖区）
- 信用卡号、社会安全号

如果必须记录用户标识符，使用不透明 ID 而非 PII。

### 高基数无界数据

不要记录完整的请求体、Info 级别的完整栈跟踪或无界集合：

```go
// 不好：无界数据
slog.Info("received", "body", string(requestBody))
slog.Info("users loaded", "users", users) // 可能有 10 万条记录

// 好：有界摘要
slog.Info("received", "content_length", len(requestBody), "content_type", ct)
slog.Info("users loaded", "count", len(users))
```

### 决策表

| 数据类型 | 记录吗？ | 替代方案 |
|----------|----------|----------|
| 请求 ID / 跟踪 ID | 是 | — |
| 用户 ID（不透明的） | 是 | — |
| HTTP 方法、路径、状态 | 是 | — |
| 错误消息 | 是 | — |
| 密码 / 令牌 | **永不** | 记录令牌前缀或 "已脱敏" |
| 完整请求体 | **否** | 记录内容长度和类型 |
| PII（邮箱、姓名） | **避免** | 记录不透明用户 ID |
| 大型集合 | **否** | 记录数量或摘要 |
| 栈跟踪 | 仅 Debug | 使用 `slog.Debug` |
