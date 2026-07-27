---
name: go-testing
description: Use when writing, reviewing, or improving Go test code — including table-driven tests, subtests, parallel tests, test helpers, test doubles, and assertions with cmp.Diff. Also use when a user asks to write a test for a Go function, even if they don't mention specific patterns like table-driven tests or subtests. Does not cover benchmark performance testing (see go-performance).
license: Apache-2.0
compatibility: Uses github.com/google/go-cmp for cmp.Diff comparisons
metadata:
  sources: "Google Style Guide, Uber Style Guide"
allowed-tools: Bash(bash:*)
---

# Go 测试

## 快速参考

| 模式 | 使用场景 |
|------|----------|
| `t.Error` | 默认 — 报告失败，继续运行 |
| `t.Fatal` | 设置失败或继续运行没有意义 |
| `cmp.Diff` | 比较 struct、slice、map、proto |
| 表驱动 | 多个用例共享相同逻辑 |
| 子测试 | 需要过滤、并行执行或命名 |
| `t.Helper()` | 任何测试辅助函数（作为第一条语句调用） |
| `t.Cleanup()` | 在辅助函数中进行清理，替代 defer |

---

## 有用的测试失败信息

> **规范**：测试失败必须在不阅读测试源码的情况下可诊断。

每条失败信息必须包含：函数名、输入、实际值（got）和期望值（want）。使用格式 `YourFunc(%v) = %v, want %v`。

```go
// 好：
t.Errorf("Add(2, 3) = %d, want %d", got, 5)

// 不好：缺少函数名和输入
t.Errorf("got %d, want %d", got, 5)
```

始终先打印 got 再打印 want：`got %v, want %v` — 绝不反转。

---

## 不使用断言库

> **规范**：不要使用断言库。对于复杂比较使用 `cmp.Diff`。

```go
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("GetPost() mismatch (-want +got):\n%s", diff)
}
```

对于 protocol buffers，添加 `protocmp.Transform()` 作为 cmp 选项。始终在 diff 信息中包含方向键 `(-want +got)`。避免比较 JSON/序列化输出 — 改为语义比较。

> 在编写自定义比较辅助函数或领域特定测试工具时，请阅读 [references/TEST-HELPERS.md](references/TEST-HELPERS.md)。

---

## t.Error vs t.Fatal

> **规范**：默认使用 `t.Error` 以在一次运行中报告所有失败。仅在无法继续时使用 `t.Fatal`。

**选择 `t.Fatal` 的场景：**
- 设置失败（数据库连接、文件加载）
- 下一个断言依赖于上一个断言成功（例如，编码后的解码）

**绝不在测试 goroutine 以外的 goroutine 中调用 `t.Fatal`/`t.FailNow`** — 改为使用 `t.Error`。

> 在编写需要在 t.Error 和 t.Fatal 之间选择的辅助函数时，或需要两者的详细示例时，请阅读 [references/TEST-HELPERS.md](references/TEST-HELPERS.md)。

---

## 表驱动测试

> 在搭建新的表驱动测试并需要标准的 struct、循环和子测试布局时，请参阅 `assets/table-test-template.go`。

> **建议**：当多个用例共享相同逻辑时使用表驱动测试。

**使用表测试的场景：** 所有用例运行相同的代码路径，没有条件设置、mock 或断言。单个 `shouldErr` bool 是可以接受的。

**不使用表测试的场景：** 用例需要复杂设置、条件 mock 或多个分支 — 改为编写单独的测试函数。

**关键规则：**
- 当用例跨越多行或有相同类型的相邻字段时，使用字段名
- 在失败信息中包含输入 — 绝不通过索引标识行

> 在编写表驱动测试、子测试或并行测试时，请阅读 [references/TABLE-DRIVEN-TESTS.md](references/TABLE-DRIVEN-TESTS.md)。

> **验证**：在生成或修改测试后，运行 `go test -run TestXxx -v` 验证测试能编译并通过。在继续之前修复任何编译错误。

---

## 测试辅助函数

> **规范**：测试辅助函数必须首先调用 `t.Helper()` 并使用 `t.Cleanup()` 进行清理。

```go
func setupTestDB(t *testing.T) *sql.DB {
    t.Helper()
    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("Could not open database: %v", err)
    }
    t.Cleanup(func() { db.Close() })
    return db
}
```

> 在编写测试辅助函数、清理函数或自定义比较工具时，请阅读 [references/TEST-HELPERS.md](references/TEST-HELPERS.md)。

---

## 测试错误语义

> **建议**：测试错误语义，而非错误消息字符串。

```go
// 不好：脆弱的字符串比较
if err.Error() != "invalid input" { ... }

// 好：语义检查
if !errors.Is(err, ErrInvalidInput) { ... }
```

对于不需要特定语义的简单存在性检查：

```go
if gotErr := err != nil; gotErr != tt.wantErr {
    t.Errorf("f(%v) error = %v, want error presence = %t", tt.input, err, tt.wantErr)
}
```

---

## 测试组织

> 在使用测试替身、选择测试包位置或规划测试设置范围时，请阅读 [references/TEST-ORGANIZATION.md](references/TEST-ORGANIZATION.md)。

> 在设计可重用的测试验证函数时，请阅读 [references/VALIDATION-APIS.md](references/VALIDATION-APIS.md)。

---

## 集成测试

> 在编写 TestMain、验收测试或需要真实 HTTP/RPC 传输层的测试时，请阅读 [references/INTEGRATION.md](references/INTEGRATION.md)。

---

## 可用脚本

- **`scripts/gen-table-test.sh`** — 生成表驱动测试脚手架

```bash
bash scripts/gen-table-test.sh ParseConfig config > config/parse_config_test.go
bash scripts/gen-table-test.sh --parallel ParseConfig config      # 带 t.Parallel()
bash scripts/gen-table-test.sh --output config/parse_config_test.go ParseConfig config
```

---

## 相关 Skill

- **错误测试**：在使用 `errors.Is`/`errors.As` 或哨兵错误测试错误语义时，请参阅 [go-error-handling](../go-error-handling/SKILL.md)
- **接口 mock**：在消费端通过实现接口创建测试替身时，请参阅 [go-interfaces](../go-interfaces/SKILL.md)
- **测试函数命名**：在命名测试函数、子测试或测试辅助工具时，请参阅 [go-naming](../go-naming/SKILL.md)
- **Linter 集成**：在 CI 或 pre-commit hooks 中与测试一起运行 linter 时，请参阅 [go-linting](../go-linting/SKILL.md)
