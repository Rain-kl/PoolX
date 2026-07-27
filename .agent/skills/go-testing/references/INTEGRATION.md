# Go 测试：集成和高级模式

TestMain、验收测试和真实传输层测试的详细参考。
来源：Google Go Style Guide（最佳实践）。

---

## TestMain

> **来源**：Google Go Style Guide（最佳实践）

当 **包中的所有测试** 都需要共同的设置且需要清理时（例如，共享数据库），使用 `func TestMain(m *testing.M)`。这 **不应该是你的首选** — 尽可能优先使用作用域测试辅助函数或 `t.Cleanup`。

```go
var db *sql.DB

func TestInsert(t *testing.T) { /* 使用 db */ }
func TestSelect(t *testing.T) { /* 使用 db */ }

func runMain(ctx context.Context, m *testing.M) (code int, err error) {
    ctx, cancel := context.WithCancel(ctx)
    defer cancel()

    d, err := setupDatabase(ctx)
    if err != nil {
        return 0, err
    }
    defer d.Close()
    db = d

    return m.Run(), nil
}

func TestMain(m *testing.M) {
    code, err := runMain(context.Background(), m)
    if err != nil {
        log.Fatal(err)
    }
    // defer 语句在 os.Exit 之后不会执行
    os.Exit(code)
}
```

关键点：
- 将设置提取到辅助函数（`runMain`）中，使 `defer` 能正确工作
- 通过 `log.Fatal` 将失败信息写入 stderr
- 确保各个测试用例保持独立 — 重置它们修改的任何全局状态

---

## 验收测试

> **来源**：Google Go Style Guide（最佳实践）

验收测试验证实现是否遵循契约，将其视为黑盒。当用户实现你的接口并且你想提供可重用的验证套件时，这种模式很有用。

### 结构

1. 创建测试辅助包（例如，为 `chess` 包创建 `chesstest`）
2. 导出一个接受被测实现的验证函数：

```go
// Package chesstest 为 chess.Player 实现提供验收测试。
package chesstest

// ExercisePlayer 在单回合中测试 Player 实现。
// 如果玩家走了正确的一步，返回 nil，否则返回描述违规行为的错误。
func ExercisePlayer(b *chess.Board, p chess.Player) error {
    move := p.Move()
    if putsOwnKingIntoCheck(b, move) {
        return &IllegalMoveError{Move: move, Reason: "puts own king in check"}
    }
    return nil
}
```

3. 最终用户针对验证函数编写简单测试：

```go
func TestAcceptance(t *testing.T) {
    player := deepblue.New()
    if err := chesstest.ExerciseGame(t, chesstest.SimpleGame, player); err != nil {
        t.Errorf("Deep Blue player failed acceptance test: %v", err)
    }
}
```

仅在设置失败时使用 `t.Fatal` — 验证错误应该返回，而非 fatal。

---

## 使用真实传输层

> **来源**：Google Go Style Guide（最佳实践）

在测试基于 HTTP 或 RPC 的组件集成时，优先使用真实传输层往返而非手动实现的客户端 mock：

```go
func TestAPIIntegration(t *testing.T) {
    // 使用假后端启动测试服务器
    srv := httptest.NewServer(newFakeHandler())
    t.Cleanup(srv.Close)

    // 对测试服务器使用真实 HTTP 客户端
    client := api.NewClient(srv.URL)
    result, err := client.GetUser(context.Background(), "user-123")
    if err != nil {
        t.Fatalf("GetUser() error: %v", err)
    }
    if result.Name != "Test User" {
        t.Errorf("GetUser().Name = %q, want %q", result.Name, "Test User")
    }
}
```

使用生产客户端配合测试服务器，可以确保测试尽可能多地覆盖真实代码，避免模拟客户端行为的复杂性。

---

## 常见错误

### 在 TestMain 中直接调用 os.Exit

`os.Exit` 立即终止进程 — defer 的清理函数永远不会执行。将设置/清理提取到辅助函数中，使 `defer` 能正确工作：

```go
// 不好：defer 不会执行
func TestMain(m *testing.M) {
    setup()
    defer cleanup()
    os.Exit(m.Run()) // cleanup() 永远不会执行
}

// 好：提取到辅助函数中，使 defer 在 os.Exit 之前执行
func runTests(m *testing.M) int {
    setup()
    defer cleanup()
    return m.Run()
}

func TestMain(m *testing.M) {
    os.Exit(runTests(m))
}
```
