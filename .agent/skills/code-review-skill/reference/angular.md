# Angular Code Review Guide

> Angular 17+ 代码审查指南，覆盖 Signals、Standalone 组件、RxJS 反模式、Zoneless 变更检测、模板最佳实践及性能优化等核心主题。

## 目录

- [Signals 与变更检测](#signals-与变更检测)
- [Standalone 组件迁移](#standalone-组件迁移)
- [RxJS 反模式](#rxjs-反模式)
- [Zoneless 变更检测](#zoneless-变更检测)
- [模板最佳实践](#模板最佳实践)
- [性能优化](#性能优化)
- [Review Checklist](#review-checklist)

---

## Signals 与变更检测

### Signal + OnPush 自动触发变更检测

```typescript
// ❌ 可变状态 + OnPush = 界面不更新
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `<p>{{ data.name }}</p>`,
})
export class UserProfile {
  data = { name: 'Alice' };
  changeName() { this.data.name = 'Bob'; } // UI 不会更新！
}

// ✅ Signal + OnPush = 自动变更检测
@Component({
  changeDetection: ChangeDetectionStrategy.OnPush,
  template: `<p>{{ name() }}</p>`,
})
export class UserProfile {
  name = signal('Alice');
  changeName() { this.name.set('Bob'); } // 自动触发 CD
}
```

### @Input() 对象变异不会被 OnPush 检测

```typescript
// ❌ 变异 Input 对象——引用不变，OnPush 不检测
@Input() config!: Config;
updateConfig() { this.config.theme = 'dark'; }

// ✅ 创建新引用
updateConfig() { this.config = { ...this.config, theme: 'dark' }; }
```

### computed() 用于派生状态

```typescript
// ❌ effect 用于同步状态——反模式，可能触发额外 CD 周期
export class CartComponent {
  total = signal(0);
  discounted = signal(0);

  constructor() {
    effect(() => this.discounted.set(this.total() * 0.9));
  }
}

// ✅ computed 用于派生状态——惰性计算，无副作用
export class CartComponent {
  total = signal(0);
  discounted = computed(() => this.total() * 0.9);
}
```

### effect() 中 Signal 读取在 await 后不会被追踪

```typescript
// ❌ await 之后读取 Signal——依赖未被追踪
effect(async () => {
  const data = await fetchUserData();
  console.log(`Theme: ${theme()}`); // theme() 未被追踪！
});

// ✅ 在 await 之前同步读取
effect(async () => {
  const currentTheme = theme(); // 同步读取，被追踪
  const data = await fetchUserData();
  console.log(`Theme: ${currentTheme}`);
});
```

### effect 只在特定场景使用

```typescript
// ❌ 用 effect 同步两个 Signal——永远用 computed
effect(() => { this.filtered.set(this.items().filter(i => i.active)); });

// ✅ effect 的合理场景：DOM 操作、分析日志、订阅外部源
effect(() => {
  const canvas = this.canvasRef.nativeElement;
  const ctx = canvas.getContext('2d');
  ctx.fillStyle = this.color();
  ctx.fillRect(0, 0, this.size(), this.size());
});

// 💡 "There are no situations where effect is good,
//    only situations where it is appropriate."
```

---

## Standalone 组件迁移

### Angular 19+ standalone 是默认值

```typescript
// ❌ Legacy NgModule 组件
@Component({
  selector: 'old-component',
  standalone: false,
})
export class OldComponent {}

// ✅ 现代 Standalone 组件（Angular 19+ standalone 是默认值）
@Component({
  selector: 'user-profile',
  imports: [ProfilePhoto, RouterLink],
  template: `<profile-photo /><a routerLink="/edit">Edit</a>`,
})
export class UserProfile {}
```

### 审查标记

```typescript
// ⚠️ 需要迁移的信号：
// 1. standalone: false
// 2. @NgModule declarations
// 3. 组件通过 NgModule 而非直接 import

// ✅ 迁移路径：
// 1. 删除 standalone: false
// 2. 将依赖添加到组件的 imports 数组
// 3. 如果不再有 declarations，删除 NgModule
```

---

## RxJS 反模式

### subscribe() 必须配 takeUntilDestroyed

```typescript
// ❌ 裸 subscribe——内存泄漏！组件销毁后仍继续接收数据
@Component({ /* ... */ })
export class UserProfile implements OnInit {
  ngOnInit() {
    this.data$.subscribe(data => this.processData(data));
  }
}

// ✅ takeUntilDestroyed——自动在组件销毁时取消（需在构造函数或注入上下文中调用）
@Component({ /* ... */ })
export class UserProfile {
  constructor() {
    this.data$.pipe(takeUntilDestroyed()).subscribe(data => {
      this.processData(data);
    });
  }
}

// ✅ 在构造函数外使用——传入 DestroyRef
@Component({ /* ... */ })
export class UserProfile {
  private destroyRef = inject(DestroyRef);

  startListening() {
    this.data$.pipe(takeUntilDestroyed(this.destroyRef)).subscribe(/* ... */);
  }
}
```

### toSignal 优于 AsyncPipe

```typescript
// ❌ AsyncPipe——需要导入，模板中有 | async
@Component({
  imports: [AsyncPipe],
  template: `{{ data$ | async }}`,
})

// ✅ toSignal——自动取消订阅，可在任何地方使用
export class UserProfile {
  data = toSignal(this.data$, { initialValue: null });
  // 模板直接用 data()
}
```

### 避免重复 toSignal 调用

```typescript
// ❌ toSignal 每次调用都创建新订阅
getData() {
  return toSignal(this.http.get('/api/data'));
}

// ✅ 存储结果
data = toSignal(this.http.get('/api/data'), { initialValue: null });
```

---

## Zoneless 变更检测

### 普通属性变异不会被检测（Angular 21+）

```typescript
// ❌ Zoneless 下普通属性赋值不触发 CD
export class UserService {
  user: User | null = null;
  loadUser() { this.user = fetchResult; } // 不触发！
}

// ✅ Signal 自动触发 CD
export class UserService {
  private _user = signal<User | null>(null);
  readonly user = this._user.asReadonly();
  loadUser() { this._user.set(fetchResult); }
}
```

### NgZone API 在 Zoneless 中失效

```typescript
// ❌ NgZone.onStable 在 zoneless 中永远不会触发
ngZone.onStable.subscribe(() => { /* 永远不触发 */ });

// ✅ 使用 afterNextRender
afterNextRender({ write: () => { /* CD 之后执行 */ } });
```

### Reactive Forms 变异需要 markForCheck

```typescript
// ❌ Reactive Forms 的 setValue/patchValue 在 zoneless 中不自动调度 CD
this.form.patchValue({ name: 'Alice' }); // UI 可能不更新

// ✅ 手动标记或通过 Signal 反映
this.form.patchValue({ name: 'Alice' });
this.cdr.markForCheck();
```

### Zoneless 下有效的 CD 触发器

| 触发器 | 说明 |
|--------|------|
| `signal.set()` / `.update()` | Signal 更新自动触发 |
| `ChangeDetectorRef.markForCheck()` | 手动标记 |
| `ComponentRef.setInput()` | 输入绑定 |
| 模板事件监听器回调 | 用户交互 |

---

## 模板最佳实践

### 复杂逻辑提取为 computed Signal

```typescript
// ❌ 模板中复杂表达式
template: `<div *ngIf="items.filter(i => i.active).length > 0 && user.role === 'admin'">`

// ✅ 提取为 computed
filteredItems = computed(() => this.items().filter(i => i.active));
shouldShow = computed(() => this.filteredItems().length > 0 && this.user().role === 'admin');
template: `@if (shouldShow()) { <div>...</div> }`
```

### 原生绑定优于 NgClass / NgStyle

```typescript
// ❌ NgClass/NgStyle——额外指令开销
template: `<div [ngClass]="{active: isActive}" [ngStyle]="{'color': textColor}">`

// ✅ 原生 class/style 绑定——性能更好
template: `<div [class.active]="isActive" [style.color]="textColor">`
```

### 模板专用成员标记 protected

```typescript
// ❂ 模板专用方法暴露为 public
export class UserProfile {
  formatName(name: string) { return name.trim(); }
}

// ✅ 模板专用成员用 protected
export class UserProfile {
  protected formatName(name: string) { return name.trim(); }
}
```

### Angular 管理的属性标记 readonly

```typescript
// ❌ input/output/model 可被意外覆盖
userId = input<string>();
userSaved = output<void>();

// ✅ readonly 防止意外赋值
readonly userId = input<string>();
readonly userSaved = output<void>();
readonly userName = model<string>();
```

### 命名规范：操作名而非事件名

```typescript
// ❌ 以事件命名
template: `<button (click)="handleClick()">Save</button>`

// ✅ 以操作命名
template: `<button (click)="saveUserData()">Save</button>`
```

---

## 性能优化

### effect 是最后手段——优先 computed

```typescript
// ❌ effect 用于状态同步——触发额外 CD，可能无限循环
effect(() => {
  this.filteredItems.set(this.items().filter(i => i.active));
});

// ✅ computed——惰性计算，无副作用，无额外 CD
filteredItems = computed(() => this.items().filter(i => i.active));
```

### afterRenderEffect 分离读写阶段

```typescript
// ❌ 无阶段指定 = mixedReadWrite = 额外 DOM 回流
afterRenderEffect(() => {
  const height = el.offsetHeight; // 读
  el.style.height = height + 10 + 'px'; // 写
});

// ✅ 分离阶段减少回流
afterRenderEffect({
  earlyRead: () => el.offsetHeight,
  write: (height) => { el.style.height = height() + 10 + 'px'; },
  read: () => verifyLayout(),
});
```

### inject() 优于构造函数注入

```typescript
// ❌ 构造函数注入——多依赖时难以阅读
export class UserService {
  constructor(
    private http: HttpClient,
    private router: Router,
    private auth: AuthService,
  ) {}
}

// ✅ inject()——更好的类型推断和可读性
export class UserService {
  private http = inject(HttpClient);
  private router = inject(Router);
  private auth = inject(AuthService);
}
```

---

## Review Checklist

### Signals 与变更检测

- [ ] Signal + OnPush 用于模板状态（非可变对象）
- [ ] `@Input()` 对象通过新引用更新（非变异）
- [ ] 派生状态用 `computed()`，不用 `effect()`
- [ ] `effect()` 中 Signal 读取在 `await` 之前
- [ ] `effect()` 只用于 DOM 操作、日志、外部源订阅

### Standalone 组件

- [ ] 无 `standalone: false`（Angular 19+）
- [ ] 组件通过 `imports` 数组导入依赖
- [ ] 无不必要的 `@NgModule`

### RxJS

- [ ] `.subscribe()` 配 `takeUntilDestroyed` 或 `async` pipe
- [ ] 优先 `toSignal` 而非 `AsyncPipe`
- [ ] 无重复 `toSignal` 调用

### Zoneless

- [ ] 模板状态通过 Signal 管理（非普通属性）
- [ ] 无 `NgZone.onStable` / `NgZone.onMicrotaskEmpty`
- [ ] Reactive Forms 变异后有 `markForCheck()`

### 模板

- [ ] 复杂逻辑提取为 `computed` Signal
- [ ] 使用原生 `[class]`/`[style]` 而非 `NgClass`/`NgStyle`
- [ ] 模板专用成员标记 `protected`
- [ ] `input`/`output`/`model` 属性标记 `readonly`
- [ ] 事件处理器以操作命名（`saveData` 而非 `handleClick`）

### 性能

- [ ] `effect()` 不用于状态同步
- [ ] `afterRenderEffect` 分离读写阶段
- [ ] `inject()` 用于依赖注入
