# 可扩展验证 API

设计可重用测试验证函数的详细参考，调用者可以将其用于验收测试。来源：Google Go Style Guide（最佳实践）。

---

## `*test` 包导出模式

当你拥有一个由他人实现的接口时，在配套的 `*test` 包中导出一个验证函数。这使实现者无需复制你的测试逻辑即可验证正确性。

```go
// Package storagetest 为 storage.Backend 提供验收测试。
package storagetest

// Verify 对任何 storage.Backend 运行验证套件。
// 返回描述第一个违规行为的错误，成功时返回 nil。
func Verify(b storage.Backend) error {
    if err := verifyRoundTrip(b); err != nil {
        return fmt.Errorf("round-trip: %w", err)
    }
    if err := verifyNotFound(b); err != nil {
        return fmt.Errorf("not-found: %w", err)
    }
    return nil
}
```

调用者编写一个薄测试来接入他们的实现：

```go
func TestMyBackend(t *testing.T) {
    b := mybackend.New(t)
    if err := storagetest.Verify(b); err != nil {
        t.Errorf("MyBackend failed acceptance: %v", err)
    }
}
```

---

## 设计可扩展的验证函数

**返回错误，而非 `*testing.T` 失败。** 这使验证函数可作为普通 Go 函数使用 — 调用者决定违规是 `t.Error` 还是 `t.Fatal`。

```go
// 好：返回错误 — 调用者控制测试流程
func ExercisePlayer(b *chess.Board, p chess.Player) error {
    move := p.Move()
    if putsOwnKingIntoCheck(b, move) {
        return &IllegalMoveError{Move: move, Reason: "puts own king in check"}
    }
    return nil
}

// 不好：调用 t.Fatal — 调用者失去控制
func ExercisePlayer(t *testing.T, b *chess.Board, p chess.Player) {
    t.Helper()
    move := p.Move()
    if putsOwnKingIntoCheck(b, move) {
        t.Fatalf("illegal move: %v puts own king in check", move)
    }
}
```

**在需要丰富诊断信息时使用自定义错误类型**：

```go
type IllegalMoveError struct {
    Move   chess.Move
    Reason string
}

func (e *IllegalMoveError) Error() string {
    return fmt.Sprintf("illegal move %v: %s", e.Move, e.Reason)
}
```

---

## 何时使用验证 API vs 简单辅助函数

| 场景 | 使用方式 |
|------|----------|
| 你拥有的接口，由他人实现 | `*test` 包中的验证 API |
| 同一包中跨测试共享设置 | 使用 `t.Helper()` 的测试辅助函数 |
| 在 2-3 个测试中重用的复杂断言 | 返回 `error` 或 `bool` 的辅助函数 |
| 一次性的设置或比较 | 内联测试代码 |

**验证 API** 在以下场景值得额外的包：
- 多个外部包将实现你的接口
- 契约有容易被忽略的非显而易见的不变量
- 你想为"正确行为"提供单一事实来源

**简单辅助函数** 在以下场景更好：
- 辅助函数是直接的设置或比较函数
- 重用是偶然的，不是已发布契约的一部分

---

## 命名约定

使用表示范围的动词命名函数：`Verify`、`Exercise`、`RunConformance`。接受被测接口作为参数 — 绝不在验证包内部构造实现。

| 包 | 函数 | 用途 |
|----|------|------|
| `storagetest` | `Verify` | 验证 `storage.Backend` |
| `chesstest` | `ExercisePlayer` | 验证 `chess.Player` |
| `cachetest` | `RunConformance` | `cache.Cache` 的完整一致性套件 |
