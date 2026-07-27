# 表驱动测试、子测试和并行测试

在 Go 中组织表驱动测试和子测试的详细参考。
来源：Google Go Style Guide、Uber Go Style Guide。

---

## 基本结构

```go
func TestCompare(t *testing.T) {
    tests := []struct {
        a, b string
        want int
    }{
        {"", "", 0},
        {"a", "", 1},
        {"", "a", -1},
        {"abc", "abc", 0},
    }
    for _, tt := range tests {
        got := Compare(tt.a, tt.b)
        if got != tt.want {
            t.Errorf("Compare(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
        }
    }
}
```

---

## 最佳实践

当测试用例跨越多行或有相同类型的相邻字段时，**使用字段名**：

```go
tests := []struct {
    name  string
    input string
    want  int
}{
    {name: "empty", input: "", want: 0},
    {name: "single", input: "a", want: 1},
}
```

**不要通过索引标识行** — 在失败信息中包含输入，而非使用 `Case #%d failed`。

---

## 避免表测试中的复杂性

当测试用例需要复杂设置、条件 mock 或多个分支时，优先使用单独的测试函数而非表测试。

```go
// 不好：太多条件字段使测试难以理解
tests := []struct {
    give          string
    want          string
    wantErr       error
    shouldCallX   bool
    shouldCallY   bool
    giveXResponse string
    giveXErr      error
    giveYResponse string
    giveYErr      error
}{...}

for _, tt := range tests {
    t.Run(tt.give, func(t *testing.T) {
        if tt.shouldCallX {
            xMock.EXPECT().Call().Return(tt.giveXResponse, tt.giveXErr)
        }
        if tt.shouldCallY {
            yMock.EXPECT().Call().Return(tt.giveYResponse, tt.giveYErr)
        }
        // ...
    })
}

// 好：单独的专注测试更清晰
func TestShouldCallX(t *testing.T) {
    xMock.EXPECT().Call().Return("XResponse", nil)
    got, err := DoComplexThing("inputX", xMock, yMock)
    // 断言...
}

func TestShouldCallYAndFail(t *testing.T) {
    yMock.EXPECT().Call().Return("YResponse", nil)
    _, err := DoComplexThing("inputY", xMock, yMock)
    // 断言错误...
}
```

**表测试最适合以下场景：**

- 所有用例运行相同逻辑（无条件断言）
- 所有用例的设置相同
- 没有基于测试用例字段的条件 mock
- 所有表字段在所有测试中都被使用

如果测试体短且直接，单个 `shouldErr` 字段用于成功/失败检查是可以接受的。

---

## 子测试

使用 `t.Run` 实现更好的组织、过滤和并行执行。

### 子测试命名

- 使用清晰、简洁的名称：`t.Run("empty_input", ...)`、`t.Run("hu_to_en", ...)`
- 避免冗长的描述或斜杠（斜杠会破坏测试过滤）
- 子测试必须独立 — 不共享状态或执行顺序依赖

### 带子测试的表测试

```go
func TestTranslate(t *testing.T) {
    tests := []struct {
        name, srcLang, dstLang, input, want string
    }{
        {"hu_en_basic", "hu", "en", "köszönöm", "thank you"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            if got := Translate(tt.srcLang, tt.dstLang, tt.input); got != tt.want {
                t.Errorf("Translate(%q, %q, %q) = %q, want %q",
                    tt.srcLang, tt.dstLang, tt.input, got, tt.want)
            }
        })
    }
}
```

---

## 并行测试

在表测试中使用 `t.Parallel()` 时，注意循环变量捕获：

```go
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        t.Parallel()
        // Go 1.22+：tt 在每次迭代中被正确捕获
        // Go 1.21-：在此处添加 "tt := tt" 来捕获变量
        got := Process(tt.give)
        if got != tt.want {
            t.Errorf("Process(%q) = %q, want %q", tt.give, got, tt.want)
        }
    })
}
```
