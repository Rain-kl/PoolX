# 测试辅助函数、断言和比较

编写测试辅助函数、避免断言库以及在 t.Error 和 t.Fatal 之间选择的详细参考。
来源：Google Go Style Guide、Uber Go Style Guide。

---

## 测试辅助函数模式

测试辅助函数必须首先调用 `t.Helper()`，使失败指向调用者。
对设置失败使用 `t.Fatal`，对清理使用 `t.Cleanup`。

```go
func mustLoadTestData(t *testing.T, filename string) []byte {
    t.Helper()
    data, err := os.ReadFile(filename)
    if err != nil {
        t.Fatalf("Setup failed: could not read %s: %v", filename, err)
    }
    return data
}

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

**关键规则：**
- 将 `t.Helper()` 作为第一条语句调用，将失败归因于调用者
- 对设置失败使用 `t.Fatal`（不要从辅助函数返回错误）
- 使用 `t.Cleanup()` 进行清理而非 defer — 即使测试调用 `t.FailNow` 它也会执行

---

## 避免断言库

> **规范**：不要创建或使用断言库。

断言库会碎片化开发者体验，并且经常产生无用的失败信息。

```go
// 不好：
assert.IsNotNil(t, "obj", obj)
assert.StringEq(t, "obj.Type", obj.Type, "blogPost")
assert.IntEq(t, "obj.Comments", obj.Comments, 2)

// 好：使用 cmp 包和标准比较
want := BlogPost{
    Type:     "blogPost",
    Comments: 2,
    Body:     "Hello, world!",
}
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("GetPost() mismatch (-want +got):\n%s", diff)
}
```

### 领域特定比较

对于领域特定比较，返回值或错误而非调用 `t.Error`：

```go
func postLength(p BlogPost) int { return len(p.Body) }

func TestBlogPost(t *testing.T) {
    post := BlogPost{Body: "Hello"}
    if got, want := postLength(post), 5; got != want {
        t.Errorf("postLength(post) = %v, want %v", got, want)
    }
}
```

---

## 比较和 Diff

对于复杂类型，优先使用 `cmp.Equal` 和 `cmp.Diff`。始终在 diff 信息中包含方向键 `(-want +got)`。

```go
// struct 比较
want := &Doc{Type: "blogPost", Authors: []string{"isaac", "albert"}}
if diff := cmp.Diff(want, got); diff != "" {
    t.Errorf("AddPost() mismatch (-want +got):\n%s", diff)
}

// Protocol buffers
if diff := cmp.Diff(want, got, protocmp.Transform()); diff != "" {
    t.Errorf("Foo() mismatch (-want +got):\n%s", diff)
}
```

**避免不稳定的比较** — 不要比较可能变化的 JSON/序列化输出。改为语义比较。

---

## t.Error vs t.Fatal：详细指南

使用 `t.Error` 保持测试继续运行，在一次运行中报告所有失败：

```go
// 好：报告所有不匹配
if diff := cmp.Diff(wantMean, gotMean); diff != "" {
    t.Errorf("Mean mismatch (-want +got):\n%s", diff)
}
if diff := cmp.Diff(wantVariance, gotVariance); diff != "" {
    t.Errorf("Variance mismatch (-want +got):\n%s", diff)
}
```

当后续检查无意义时使用 `t.Fatal`：

```go
gotEncoded := Encode(input)
if gotEncoded != wantEncoded {
    t.Fatalf("Encode(%q) = %q, want %q", input, gotEncoded, wantEncoded)
}
gotDecoded, err := Decode(gotEncoded)
if err != nil {
    t.Fatalf("Decode(%q) error: %v", gotEncoded, err)
}
```

### 不要从 Goroutine 中调用 t.Fatal

> **规范**：绝不在测试 goroutine 以外的 goroutine 中调用 `t.Fatal`、`t.Fatalf` 或 `t.FailNow`。改为使用 `t.Error` 并让 goroutine 自然返回。
