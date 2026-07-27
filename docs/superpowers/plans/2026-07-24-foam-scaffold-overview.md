# Foam Scaffold — Overall Implementation Plan

> **For agentic workers:** Implement in order: Backend plan → Frontend plan → this overview’s final docs/rename/verify tasks.  
> Use `superpowers:subagent-driven-development` (recommended) or `superpowers:executing-plans`.  
> Steps use checkbox (`- [ ]`) syntax.

**Goal:** Turn Grok2API into **Foam**, a runnable full-stack scaffold for secondary development.

**Architecture:** In-place trim. Backend keeps layered skeleton + admin auth + `example` CRUD. Frontend becomes an example gallery (`/example/component` + page templates). Rename packages/binaries only after both sides build and smoke.

**Tech Stack:** Go 1.26, Gin, GORM (SQLite/Postgres), React 19, Vite, Tailwind 4, shadcn/Radix, pnpm, Docker.

**Spec:** `docs/superpowers/specs/2026-07-24-foam-scaffold-design.md`

## Companion Plans

| Plan | Path | Responsibility |
| --- | --- | --- |
| Backend | `docs/superpowers/plans/2026-07-24-foam-scaffold-backend.md` | Delete business, thin app/config/http, example CRUD, build/smoke |
| Frontend | `docs/superpowers/plans/2026-07-24-foam-scaffold-frontend.md` | Example gallery routes, mock data, neutral copy, shell/nav |
| Overview (this) | `docs/superpowers/plans/2026-07-24-foam-scaffold-overview.md` | Docs, directory READMEs, global rename, final verification |

## Global Constraints

- Go module (final): `github.com/Rain-kl/Foam/backend`
- Binary / cmd: `foam` / `backend/cmd/foam`
- Frontend package: `foam-frontend`
- Component demos: **single route** `/example/component` (in-page sections, no child paths)
- Page demos: `/example/page/dashboard`, `/example/page/table`, `/example/page/chat`
- Default post-login: `/example/page/dashboard`
- Gallery pages use mock/static data; only auth/health/(optional) example CRUD hit backend
- No Grok/provider/gateway business left on the runnable path
- **Rename last** — do not mass-rename until backend + frontend smoke pass under temporary old module path if needed; prefer completing trim first, then one rename pass
- Neutral product copy only (Foam / Example)
- Evidence before “done”: build commands + route smoke

## Execution Order

```text
1. Backend plan (all tasks)     → go build + login + /api/examples CRUD
2. Frontend plan (all tasks)    → pnpm build + four example routes
3. Overview Task O1             → directory READMEs + config/docker slim (pre-rename names OK)
4. Overview Task O2             → AGENTS.md + docs/*
5. Overview Task O3             → global rename to Foam
6. Overview Task O4             → final verification gate
```

During backend/frontend work, **imports may still use** `github.com/chenyme/grok2api/backend` until Task O3. That is intentional to avoid double churn.

---

### Task O1: Scaffold docs in directories + slim deploy artifacts (pre-rename)

**Files:**
- Create: short `README.md` under key backend/frontend convention dirs (see list below)
- Modify: `config.example.yaml` (skeleton only)
- Modify: `Dockerfile`, `docker-compose.yml`, root `Makefile` (remove business-only stages/comments; names can stay until O3)
- Modify: root `README.md` / `README.zh-CN.md` → temporary Foam scaffold description (full brand polish in O3)

**Directory README targets (create if directory kept as convention):**

```text
backend/internal/domain/README.md
backend/internal/application/README.md
backend/internal/repository/README.md
backend/internal/transport/http/README.md
backend/internal/infra/README.md
backend/internal/pkg/README.md
frontend/src/features/README.md
frontend/src/entities/README.md
```

Each README must answer:
1. What this layer is for
2. What code belongs here
3. How to extend by copying `example`

- [ ] **Step 1: Write directory READMEs** using the template:

```md
# <layer>

## Role
...

## Put here
- ...

## Extend
Copy the `example` vertical slice (backend) or `features/example` (frontend), then rename.
```

- [ ] **Step 2: Slim `config.example.yaml`** to only:

```yaml
server:
  listen: "127.0.0.1:8000"
  maxBodyBytes: 33554432
  readTimeout: 15m
  requestTimeout: 2h
  swaggerEnabled: false

auth:
  accessTokenTTL: 15m
  refreshTokenTTL: 720h
  secureCookies: false

secrets:
  jwtSecret: "replace-with-at-least-32-characters"
  credentialEncryptionKey: "replace-with-base64-key"

bootstrapAdmin:
  username: "admin"
  password: "replace-with-a-strong-password"

frontend:
  staticPath: "./frontend/dist"

database:
  driver: sqlite
  sqlite:
    path: "./data/backend.db"
  postgres:
    dsn: "postgres://user:password@127.0.0.1:5432/foam?sslmode=disable"
    maxOpenConns: 50
    maxIdleConns: 10
```

Drop provider/routing/media/audit/clientKeyDefaults/runtimeStore unless backend plan still requires runtimeStore for admin rate limit — if required, keep minimal `runtimeStore.driver: memory`.

- [ ] **Step 3: Slim Docker/Make** — remove FlareSolverr/WARP comments if unused; keep multi-stage build working for frontend+backend.

- [ ] **Step 4: Commit**

```bash
git add config.example.yaml Dockerfile docker-compose.yml Makefile README.md README.zh-CN.md \
  backend/internal/**/README.md frontend/src/**/README.md
git commit -m "docs: add directory maps and slim deploy skeleton for Foam"
```

---

### Task O2: AGENTS.md + focused docs

**Files:**
- Create: `AGENTS.md`
- Create: `docs/architecture.md`
- Create: `docs/backend-development.md`
- Create: `docs/frontend-development.md`
- Create: `docs/directory-map.md`

Note: `.gitignore` ignores `/docs/*`. Force-add docs that must live in repo:

```bash
git add -f AGENTS.md docs/architecture.md docs/backend-development.md \
  docs/frontend-development.md docs/directory-map.md
```

- [ ] **Step 1: Write `AGENTS.md` as a router** (keep short):

```md
# Foam — Agent Instructions

> Scope: Always-on rules only. Task details live in docs/.

## What this is
Foam is a Go + React full-stack scaffold for secondary development.
Not a Grok gateway product.

## Core principles
- Layered backend: domain → application → repository → transport → infra
- Extend by copying the `example` slice, never by reintroducing deleted business packages
- Frontend gallery is the UI pattern library; new product features copy `features/example`
- Neutral copy only (Foam / Example)
- Evidence before claiming done (build + smoke)

## Package / binary
- Module: `github.com/Rain-kl/Foam/backend`
- Binary: `foam` (`backend/cmd/foam`)
- Frontend package: `foam-frontend`

## Read-on-demand index

| Task | Read first | Trigger |
| --- | --- | --- |
| Architecture / layering | `docs/architecture.md` | new module, request flow |
| Backend work | `docs/backend-development.md` | Go services, tests, example CRUD |
| Frontend work | `docs/frontend-development.md` | routes, components, pages |
| Directory purpose | `docs/directory-map.md` | where to put files |

## Always-on
- Do not resurrect Grok provider/gateway code
- Keep `/example/component` as a single page (sections inside)
- Prefer mock data on gallery pages; real API only for auth/health/example CRUD
- User instruction > nearest AGENTS/docs > this file
```

- [ ] **Step 2: Write the four docs** covering:
  - architecture: mermaid request flow (login + example CRUD + SPA)
  - backend-dev: how to add a resource by cloning example
  - frontend-dev: route table + component reuse rules
  - directory-map: table of major paths

- [ ] **Step 3: Commit with `git add -f` as needed**

```bash
git commit -m "docs: add AGENTS.md and Foam development guides"
```

---

### Task O3: Global rename to Foam

**Do only after backend + frontend plans smoke-pass.**

**Files (mechanical pass):**
- `backend/go.mod` module path
- All Go imports `github.com/chenyme/grok2api/backend` → `github.com/Rain-kl/Foam/backend`
- `backend/cmd/grok2api` → `backend/cmd/foam` (move `main.go`)
- Root/backend `Makefile` targets
- `Dockerfile`, `docker-compose.yml`, `.github/**` if present
- `frontend/package.json` name → `foam-frontend`
- Cookie/lock names: `grok2api_admin_refresh` → `foam_admin_refresh`
- Refresh lock: `grok2api:admin-session-refresh` → `foam:admin-session-refresh`
- i18n `appName`, README, LICENSE branding as appropriate
- `VERSION` keep or reset to `0.1.0` (scaffold baseline)

- [ ] **Step 1: Move cmd**

```bash
mkdir -p backend/cmd/foam
git mv backend/cmd/grok2api/main.go backend/cmd/foam/main.go
rmdir backend/cmd/grok2api 2>/dev/null || true
```

- [ ] **Step 2: Rewrite module path**

```bash
# from backend/
# edit go.mod:
# module github.com/Rain-kl/Foam/backend
rg -l 'github.com/chenyme/grok2api/backend' | while read f; do
  sed -i '' 's|github.com/chenyme/grok2api/backend|github.com/Rain-kl/Foam/backend|g' "$f"
done
```

On Linux drop the `''` after `-i`. Prefer a single controlled replace, then `go mod tidy`.

- [ ] **Step 3: Update Makefile / Docker / compose / frontend package name / cookies**

- [ ] **Step 4: Gate residual brand strings**

```bash
rg -n 'grok2api|chenyme/grok|Grok2API|Grok Build|Grok Web|Grok Console' \
  --glob '!**/.git/**' --glob '!**/pnpm-lock.yaml' --glob '!**/go.sum' || true
```

Expected: zero hits in source (docs history under superpowers may mention migration — OK if clearly historical).

- [ ] **Step 5: Build both sides**

```bash
cd backend && go build -o /tmp/foam ./cmd/foam
cd frontend && pnpm build
```

- [ ] **Step 6: Commit**

```bash
git commit -m "refactor: rename project to Foam (module, binary, branding)"
```

---

### Task O4: Final verification gate

- [ ] **Step 1: Config for local smoke**

```bash
cp config.example.yaml config.yaml
# set jwtSecret (>=32 chars), credentialEncryptionKey (base64 32 bytes), bootstrapAdmin.password (>=8)
```

- [ ] **Step 2: Run backend**

```bash
make run
# or: cd backend && go run ./cmd/foam --config ../config.yaml
```

- [ ] **Step 3: API smoke**

```bash
curl -sS http://127.0.0.1:8000/healthz
# login
curl -sS -c /tmp/foam.ck -X POST http://127.0.0.1:8000/api/admin/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"<password>"}'
# create example (use accessToken from response)
curl -sS -X POST http://127.0.0.1:8000/api/admin/v1/examples \
  -H "Authorization: Bearer <token>" \
  -H 'Content-Type: application/json' \
  -d '{"name":"demo","description":"hello"}'
curl -sS http://127.0.0.1:8000/api/admin/v1/examples \
  -H "Authorization: Bearer <token>"
```

- [ ] **Step 4: Frontend smoke** (dev or static)

Open and confirm no console crash:
- `/login`
- `/example/component`
- `/example/page/dashboard`
- `/example/page/table`
- `/example/page/chat`

- [ ] **Step 5: Residual check + commit fixups if any**

```bash
rg -n 'grok2api|chenyme/grok' --glob '!**/.git/**' --glob '!**/docs/superpowers/**' || true
```

- [ ] **Step 6: Final commit if needed**

```bash
git commit -m "chore: Foam scaffold verification fixups"
```

---

## Done definition

- [ ] Backend plan complete
- [ ] Frontend plan complete
- [ ] O1–O4 complete
- [ ] Module is `github.com/Rain-kl/Foam/backend`, binary `foam`
- [ ] Example gallery routes work
- [ ] `AGENTS.md` present and accurate
