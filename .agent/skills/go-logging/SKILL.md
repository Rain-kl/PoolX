---
name: go-logging
description: 在选择日志方案、配置 slog、编写结构化日志语句或决定日志级别时使用。也适用于设置生产日志、为日志添加请求作用域上下文或从 log 迁移到 slog 的场景，即使用户未明确提及日志。不涵盖错误处理策略（参见 go-error-handling）。
license: Apache-2.0
compatibility: slog requires Go 1.21+; slog/slogtest requires Go 1.22+
metadata:
  sources: "Google Style Guide, Uber Style Guide"
---

# Go 日志

## 核心原则

日志是给**运维人员**看的，不是给开发人员看的。每一行日志都应该帮助某人诊断生产问题。如果不能达到这个目的，就是噪音。

---

## 选择日志器

> **规范**：在新的 Go 代码中使用 `log/slog`。

`slog` 是结构化的、分级别的，并且在标准库中（Go 1.21+）。它涵盖了绝大多数生产日志需求。

```
选择哪个日志器？
├─ 新的生产代码      → log/slog
├─ 简单 CLI / 一次性  → log（标准库）
└─ 有性能瓶颈        → zerolog 或 zap（先做基准测试）
```

除非性能分析显示 `slog` 在热路径中是瓶颈，否则不要引入第三方日志库。引入时，保持相同的结构化键值风格。

> 在设置 slog handler、配置 JSON/文本输出或从 log.Printf 迁移到 slog 时，阅读 [references/LOGGING-PATTERNS.md](references/LOGGING-PATTERNS.md)。

---

## 结构化日志

> **规范**：始终使用键值对。永远不要将值插值到消息字符串中。

消息是描述发生了什么的**静态描述**。动态数据放在键值属性中：

```go
// 好：静态消息，结构化字段
slog.Info("order placed", "order_id", orderID, "total", total)

// 不好：动态数据嵌入到消息字符串中
slog.Info(fmt.Sprintf("order %d placed for $%.2f", orderID, total))
```

### 键名

> **建议**：日志属性键使用 `snake_case`。

键应为小写、下划线分隔，并在整个代码库中保持一致：`user_id`、`request_id`、`elapsed_ms`。

### 类型化属性

对于性能关键路径，使用类型化构造函数以避免分配：

```go
slog.LogAttrs(ctx, slog.LevelInfo, "request handled",
    slog.String("method", r.Method),
    slog.Int("status", code),
    slog.Duration("elapsed", elapsed),
)
```

> 在优化日志性能或使用 Enabled() 进行预检查时，阅读 [references/LEVELS-AND-CONTEXT.md](references/LEVELS-AND-CONTEXT.md)。

---

## 日志级别

> **建议**：一致地遵循这些级别语义。

| 级别 | 何时使用 | 生产默认 |
|------|----------|----------|
| Debug | 仅开发人员的诊断，跟踪内部状态 | 禁用 |
| Info  | 重要的生命周期事件：启动、关闭、配置加载 | 启用 |
| Warn  | 意外但可恢复：使用了弃用功能、重试成功 | 启用 |
| Error | 操作失败，需要运维人员关注 | 启用 |

**经验法则**：
- 如果没有人需要对其采取行动，那就不是 Error——使用 Warn 或 Info
- 如果只在连接调试器时才有用，那就是 Debug
- `slog.Error` 应始终包含 `"err"` 属性

```go
slog.Error("payment failed", "err", err, "order_id", id)
slog.Warn("retry succeeded", "attempt", n, "endpoint", url)
slog.Info("server started", "addr", addr)
slog.Debug("cache lookup", "key", key, "hit", hit)
```

> 在 Warn 和 Error 之间选择或定义自定义详细级别时，阅读 [references/LEVELS-AND-CONTEXT.md](references/LEVELS-AND-CONTEXT.md)。

---

## 请求作用域日志

> **建议**：从 context 派生日志器以携带请求作用域字段。

使用中间件为日志器添加请求 ID、用户 ID 或跟踪 ID，然后通过 context 或作为显式参数将增强后的日志器传递给下游：

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        logger := slog.With("request_id", requestID(r))
        ctx := context.WithValue(r.Context(), loggerKey, logger)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

该请求中所有后续的日志调用都会自动携带 `request_id`。

> 在实现日志中间件或通过 context 传递日志器时，阅读 [references/LOGGING-PATTERNS.md](references/LOGGING-PATTERNS.md)。

---

## 日志或返回，不要同时

> **规范**：每个错误恰好处理一次——要么记录它，要么返回它。

记录错误然后返回它会导致重复噪音，因为栈上游的调用者也会处理该错误。

```go
// 不好：在这里记录，并且栈上游的每个调用者也会记录
if err != nil {
    slog.Error("query failed", "err", err)
    return fmt.Errorf("query: %w", err)
}

// 好：包装并返回——让调用者决定
if err != nil {
    return fmt.Errorf("query: %w", err)
}
```

**例外**：HTTP 处理器和其他栈顶边界可以在服务端记录详细错误，同时向客户端返回脱敏消息：

```go
if err != nil {
    slog.Error("checkout failed", "err", err, "user_id", uid)
    http.Error(w, "internal error", http.StatusInternalServerError)
    return
}
```

参见 [go-error-handling](../go-error-handling/SKILL.md) 了解完整的处理一次模式和错误包装指导。

---

## 不应记录的内容

> **规范**：永远不要记录密钥、凭证、PII 或高基数无界数据。

- 密码、API 密钥、令牌、会话 ID
- 完整的信用卡号、社会安全号
- 可能包含用户数据的请求/响应体
- 无界大小的完整切片或映射

> 在决定哪些数据可以安全包含在日志属性中时，阅读 [references/LEVELS-AND-CONTEXT.md](references/LEVELS-AND-CONTEXT.md)。

---

## 快速参考

| 应该 | 不应该 |
|------|--------|
| `slog.Info("msg", "key", val)` | `log.Printf("msg %v", val)` |
| 静态消息 + 结构化字段 | 在消息中使用 `fmt.Sprintf` |
| `snake_case` 键 | camelCase 或不一致的键 |
| 日志或返回错误 | 同时日志和返回同一错误 |
| 从 context 派生日志器 | 每次调用创建新日志器 |
| `slog.Error` 配合 `"err"` 属性 | 用 `slog.Info` 记录错误 |
| 在热路径上预检查 `Enabled()` | 始终分配日志参数 |

---

## 相关技能

- **错误处理**：在决定是记录还是返回错误，或了解处理一次模式时，参见 [go-error-handling](../go-error-handling/SKILL.md)
- **上下文传播**：在通过 context 传递请求作用域值（包括日志器）时，参见 [go-context](../go-context/SKILL.md)
- **性能**：在优化热路径日志或减少日志调用中的分配时，参见 [go-performance](../go-performance/SKILL.md)
- **代码审查**：在审查 Go PR 中的日志实践时，参见 [go-code-review](../go-code-review/SKILL.md)
