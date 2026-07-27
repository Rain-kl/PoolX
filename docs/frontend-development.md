# 前端开发

包名（目标）：`foam-frontend`  
栈：React 19、Vite、TypeScript、Tailwind、shadcn/ui、React Router、TanStack Query、i18next

## 产品形态

管理端是 **PoolX 代理池控制面**：

- 节点导入与管理（配置文件/订阅）
- 端口配置与代理池组织（工作台）
- 内核运行时控制（启动/停止/热重载/日志）
- 系统设置

## 路由

| 路径 | 内容 |
| --- | --- |
| `/login` | 登录 |
| `/clash/nodes` | 节点管理 |
| `/clash/source-configs` | 配置来源管理 |
| `/clash/port-profiles` | 端口配置（工作台） |
| `/clash/runtime` | 内核运行时控制 |
| `/clash/dashboard` | 运行时仪表板 |
| `/settings` | 系统设置（真实 API） |

- 定义文件：`frontend/src/app/router.tsx`

### 导航

App shell（`app/app-shell.tsx`）：

- 代理池相关功能入口
- 侧栏底部：用户名 · 更多（外观/语言/改密/退出）· **设置图标**（`/settings`，在「更多」右侧）

## 代码布局

```text
frontend/src/
  main.tsx
  app/
    providers.tsx          # Query / theme / auth 等
    auth-boundary.tsx      # 登录态门禁
    app-shell.tsx          # 侧栏 + 顶栏 + Outlet
    router.tsx
    deferred-pages.tsx     # 懒加载页面出口
  features/
    auth/login-page.tsx
    clash/
      nodes/               # 节点管理
      source-configs/      # 配置来源
      port-profiles/       # 端口配置工作台
      runtime/             # 内核运行时控制页
      dashboard/           # 运行时仪表板
    settings/
  components/ui/           # 原子 shadcn 组件（无业务文案）
  shared/
    api/                   # fetch client、decoder、ApiError
    auth/                  # AuthProvider / useAuth
    components/            # PageHeader, DataTableShell, Pagination, …
    hooks/ lib/ config/ i18n/
```

## 组件复用规则

1. **原子 UI** 放 `components/ui`（Button、Dialog、Table…）
2. **跨页模式** 放 `shared/components`（PageHeader、DataTableShell、Pagination、DataState…）
3. **域内逻辑** 放 `features/<area>`
4. 新增可视化组件：先抽到 `ui` 或 `shared`，再在业务页使用
5. 避免复制粘贴第二套 Table/Dialog；合并后再展示

## 数据与 API

- HTTP 客户端：`shared/api/client.ts`  
  - 统一解析 `{ data }` / `{ error }` envelope  
  - access token + refresh 锁（session）
- 运行时配置：`shared/config/runtime-config.ts`（API base 等）
- 鉴权：`shared/auth` + `features/auth`

## 文案与 i18n

- 字符串走 `shared/i18n`（`zh-CN` + `en`）
- 产品名：**PoolX**
- 禁止恢复已删除业务域的文案键与导航项

## 本地开发

**日常调试（推荐）**：不必先 `pnpm build`。后端与 Vite 并行启动：

```bash
# 终端 1 — API :8000（可不依赖 frontend/dist）
cd backend && go run ./cmd/foam --config ../config.yaml

# 终端 2 — SPA + HMR :8010，反代 API 到后端
cd frontend
pnpm install
pnpm dev          # http://127.0.0.1:8010
```

Vite 将 `/api`、`/v1`、`/healthz`、`/readyz`、`/swagger` 代理到 `http://127.0.0.1:8000`（可用 `VITE_DEV_API_TARGET` 覆盖）。

```bash
pnpm lint
pnpm format        # biome format --write .
pnpm format:check
pnpm build         # tsc + vite build → dist/
```

**出包 / 嵌入模式**：`pnpm build` 后由 `foam` 通过 `frontend.staticPath`（默认 `./frontend/dist`）托管 SPA；访问 `server.listen`（默认 `:8000`）。开发代理与嵌入托管互不影响。

## 新增页面 checklist

1. 在 `features/<area>/` 写页面组件
2. 需要懒加载则挂到 `deferred-pages.tsx`
3. 在 `router.tsx` 注册路径（鉴权树下用 `AuthBoundary`）
4. 在 `app-shell` 导航增加入口（若需要）
5. 文案进 i18n
6. 能复用的抽到 `components/ui` 或 `shared/components`
7. `pnpm build` 通过后再说完成

## 反模式

- 在 `components/ui` 写业务 API 调用
- 为每个页面复制一套 Table 壳
- 不更新 router/nav 只加文件
