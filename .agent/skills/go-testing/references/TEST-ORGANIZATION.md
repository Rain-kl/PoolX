# 测试组织参考

来源：Google Go Style Guide（最佳实践、决策）。

---

## 测试替身类型

| 替身 | 用途 | 有状态？ | 验证调用？ |
|------|------|----------|-----------|
| Stub | 返回预设数据 | 否 | 否 |
| Fake | 可工作但简化的实现 | 是 | 否 |
| Spy | 记录调用以供后续检查 | 是 | 是 |

**优先使用 fake 而非 mock。** Fake 更具可读性且不需要 mock 框架。仅在验证副作用（例如，分析事件）时使用 spy。

```go
// Fake：可工作的内存实现
type FakeUserStore struct {
    users map[string]*User
}

func (f *FakeUserStore) GetUser(id string) (*User, error) {
    u, ok := f.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return u, nil
}

// Spy：记录调用以供后续断言
type SpyEmailSender struct{ Sent []string }

func (s *SpyEmailSender) Send(to, body string) error {
    s.Sent = append(s.Sent, to)
    return nil
}
```

---

## 测试替身命名约定

> **建议**：为测试替身（stub、fake、spy）遵循一致的命名。

**包命名**：在生产代码旁边创建一个 `*test` 包（例如，为 `creditcard` 包创建 `creditcardtest`，为独立的 fake 服务创建 `fakeauthservice`）。

```go
// 好：在 creditcardtest 包中

// 单个替身 — 使用简单名称
type Stub struct{}
func (Stub) Charge(*creditcard.Card, money.Money) error { return nil }

// 多种行为 — 按行为命名
type AlwaysCharges struct{}
type AlwaysDeclines struct{}

// 多种类型 — 包含类型名
type StubService struct{}
type StubStoredValue struct{}
```

**局部变量**：为测试替身变量添加替身类型前缀，使调用处更清晰：

```go
// 好：替身类型立即可见
spyCC := &creditcardtest.Spy{}
stubDB := &dbtest.Stub{Balance: 100}

// 不好：模糊 — 这是真实的还是替身？
cc := &creditcardtest.Spy{}
db := &dbtest.Stub{Balance: 100}
```

---

## 独立测试辅助包

当多个包需要相同的替身、辅助函数有足够的逻辑需要自己的测试、或者你想为接口实现者提供验收测试套件时，创建独立的测试辅助包。

| 模式 | 使用场景 | 示例 |
|------|----------|------|
| `footest` | `foo` 包的通用测试辅助 | `creditcardtest`、`usertest` |
| `fakeX` | 独立的 fake 服务包 | `fakeauthservice`、`fakestorage` |

```go
package usertest

func NewFakeStore(t *testing.T, users ...*user.User) *FakeUserStore {
    t.Helper()
    store := &FakeUserStore{users: make(map[string]*user.User)}
    for _, u := range users {
        store.users[u.ID] = u
    }
    return store
}
```

导出接受 `*testing.T` 的构造函数，以便调用 `t.Helper()` 和 `t.Cleanup()`。

---

## 测试包

| 包声明 | 使用场景 |
|--------|----------|
| `package foo` | 同包测试，可以访问非导出标识符 |
| `package foo_test` | 黑盒测试，避免循环依赖 |

两者都放在同一目录下的 `foo_test.go` 文件中。

**使用 `package foo`（白盒）** 当你需要测试非导出函数或内部状态时。

**使用 `package foo_test`（黑盒）** 当仅测试公共 API、打破导入循环或验证外部可用性时。

```go
package parser_test  // 黑盒：仅测试导出的 API

import "mymodule/parser"

func TestParse(t *testing.T) {
    got, err := parser.Parse("input")
    // ...
}
```

如果黑盒测试需要非导出符号，在 `package foo`（非 `foo_test`）中创建 `export_test.go` 来暴露它。谨慎使用。

---

## 设置作用域

> **建议**：保持设置仅限于需要它的测试。

每个测试中的显式设置更清晰，避免惩罚不相关的测试：

```go
// 好：在需要它的测试中显式设置
func TestParseData(t *testing.T) {
    data := mustLoadDataset(t)
    // ...
}

func TestUnrelated(t *testing.T) {
    // 不需要为数据集加载付出代价
}
```

**避免使用全局 `init` 进行测试设置** — 它会对文件中的每个测试运行，即使是不相关的测试。

**子测试设置**：当一组子测试共享设置时，使用带 `t.Run` 的父测试：

```go
func TestDatabase(t *testing.T) {
    db := setupTestDB(t)

    t.Run("Insert", func(t *testing.T) {
        // 使用 db
    })
    t.Run("Select", func(t *testing.T) {
        // 使用 db
    })
}
```

这将数据库的生命周期限定在需要它的子测试范围内。仅在万不得已时使用 `TestMain`（参见 [INTEGRATION.md](INTEGRATION.md)）。
