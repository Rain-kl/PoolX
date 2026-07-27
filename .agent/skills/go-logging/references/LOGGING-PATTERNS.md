# 日志模式

关于 slog 设置、handler 配置、测试、HTTP 中间件以及从旧版 `log` 包迁移的详细模式。

## 设置 slog

### 基本配置

```go
package main

import (
    "log/slog"
    "os"
)

func main() {
    // JSON handler 用于生产（机器可解析）
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    slog.Info("server started", "addr", ":8080")
    // 输出：{"time":"...","level":"INFO","msg":"server started","addr":":8080"}
}
```

### 用于开发的 Text Handler

```go
// 本地开发的人类可读输出
logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
slog.SetDefault(logger)
// 输出：time=... level=DEBUG msg="cache lookup" key=user:42 hit=true
```

### 动态级别控制

使用 `slog.LevelVar` 在运行时更改最低级别（例如通过管理端点或信号处理器）：

```go
var programLevel = new(slog.LevelVar) // 默认 Info

func init() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: programLevel,
    }))
    slog.SetDefault(logger)
}

// 从管理端点或信号处理器调用
func enableDebug() {
    programLevel.Set(slog.LevelDebug)
}
```

---

## 自定义 Handler 模式

### 添加源位置

```go
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    AddSource: true,
    Level:     slog.LevelInfo,
}))
// 输出包含："source":{"function":"main.handleRequest","file":"server.go","line":42}
```

### 使用默认属性包装 Handler

使用 `slog.Handler` 中间件向每条日志记录注入字段：

```go
type contextHandler struct {
    inner   slog.Handler
    attrs   []slog.Attr
}

func (h *contextHandler) Enabled(ctx context.Context, level slog.Level) bool {
    return h.inner.Enabled(ctx, level)
}

func (h *contextHandler) Handle(ctx context.Context, r slog.Record) error {
    r.AddAttrs(h.attrs...)
    return h.inner.Handle(ctx, r)
}

func (h *contextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    return &contextHandler{inner: h.inner.WithAttrs(attrs), attrs: h.attrs}
}

func (h *contextHandler) WithGroup(name string) slog.Handler {
    return &contextHandler{inner: h.inner.WithGroup(name), attrs: h.attrs}
}
```

### 多 Handler（扇出）

写入多个目标（例如 stdout + 文件）：

```go
type multiHandler struct {
    handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
    for _, h := range m.handlers {
        if h.Enabled(ctx, level) {
            return true
        }
    }
    return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
    var errs []error
    for _, h := range m.handlers {
        if h.Enabled(ctx, r.Level) {
            if err := h.Handle(ctx, r); err != nil {
                errs = append(errs, err)
            }
        }
    }
    return errors.Join(errs...)
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
    handlers := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers {
        handlers[i] = h.WithAttrs(attrs)
    }
    return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
    handlers := make([]slog.Handler, len(m.handlers))
    for i, h := range m.handlers {
        handlers[i] = h.WithGroup(name)
    }
    return &multiHandler{handlers: handlers}
}
```

---

## 使用 slogtest 测试

Go 1.22+ 提供了 `testing/slogtest` 来验证 handler 实现：

```go
package myhandler_test

import (
    "testing"
    "testing/slogtest"
)

func TestHandler(t *testing.T) {
    // newHandler 返回你的自定义 slog.Handler 和一个
    // 将输出解析为 []map[string]any 的函数用于验证。
    results := func(t *testing.T) map[string]any {
        // 在此解析你的 handler 输出
    }

    h := NewMyHandler(buf, nil)
    slogtest.Run(t, func(t *testing.T) slog.Handler { return h }, results)
}
```

### 在测试中捕获日志

对于断言日志输出的单元测试，写入 buffer：

```go
func TestOrderProcessing(t *testing.T) {
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))

    processOrder(logger, order)

    if !strings.Contains(buf.String(), `"order_id"`) {
        t.Error("expected order_id in log output")
    }
}
```

---

## HTTP 请求日志中间件

一个完整的中间件，记录每个请求的计时、状态和请求作用域字段：

```go
func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        reqID := r.Header.Get("X-Request-ID")
        if reqID == "" {
            reqID = uuid.NewString()
        }

        logger := slog.With(
            "request_id", reqID,
            "method", r.Method,
            "path", r.URL.Path,
        )

        // 包装 response writer 以捕获状态码
        rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

        // 将日志器存入 context 供下游 handler 使用
        ctx := context.WithValue(r.Context(), loggerKey, logger)
        next.ServeHTTP(rw, r.WithContext(ctx))

        logger.Info("request completed",
            "status", rw.status,
            "elapsed_ms", time.Since(start).Milliseconds(),
        )
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.status = code
    rw.ResponseWriter.WriteHeader(code)
}
```

### 从 Context 获取日志器

```go
type ctxKey struct{}

var loggerKey = ctxKey{}

func loggerFromCtx(ctx context.Context) *slog.Logger {
    if l, ok := ctx.Value(loggerKey).(*slog.Logger); ok {
        return l
    }
    return slog.Default()
}
```

---

## 从 log.Printf 迁移到 slog

### 第 1 步：替换直接调用

```go
// 迁移前
log.Printf("user %s logged in from %s", userID, ip)

// 迁移后
slog.Info("user logged in", "user_id", userID, "ip", ip)
```

### 第 2 步：替换 main() 中的 log.Fatalf

```go
// 迁移前
log.Fatalf("failed to connect: %v", err)

// 迁移后——slog 没有 Fatal；在 main 中使用 slog + os.Exit
slog.Error("failed to connect", "err", err)
os.Exit(1)
```

### 第 3 步：桥接旧代码

如果逐步迁移，将标准 `log` 包的输出通过 slog 重定向：

```go
// 在 main() 中，设置 slog 之后：
slog.SetDefault(logger)

// 标准 log 包现在通过 slog 的默认 handler 写入。
// 这是因为 slog.SetDefault 也会更新 log.Default()。
```

### 第 4 步：替换日志器参数

```go
// 迁移前：传递 *log.Logger
func NewServer(addr string, logger *log.Logger) *Server

// 迁移后：显式传递 *slog.Logger
func NewServer(addr string, logger *slog.Logger) *Server

// 或从 handler 中的 context 派生
func (s *Server) handleRequest(ctx context.Context) {
    logger := loggerFromCtx(ctx)
    logger.Info("handling request")
}
```

### 迁移清单

| 步骤 | 更改什么 | 验证 |
|------|----------|------|
| 1 | `log.Printf` → `slog.Info/Warn/Error` | `rg 'log\.Printf'` 返回 0 个匹配 |
| 2 | `log.Fatalf` → `slog.Error` + `os.Exit(1)` 在 main 中 | 仅在 `main()` 中 |
| 3 | 在 main 中尽早设置 `slog.SetDefault` | 旧版 `log` 调用通过 slog 路由 |
| 4 | `*log.Logger` 参数 → `*slog.Logger` | 所有构造函数已更新 |
| 5 | 移除已替换处的 `"log"` 导入 | `goimports` 会自动处理 |
