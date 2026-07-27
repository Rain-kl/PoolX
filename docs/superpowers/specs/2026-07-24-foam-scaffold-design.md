# Foam Scaffold Design

**Date:** 2026-07-24  
**Status:** Draft for approval  
**Source project:** Grok2API monorepo (Go gateway + React admin)  
**Target:** Foam — a reusable full-stack development scaffold

## 1. Goal

Transform the current Grok2API codebase into **Foam**, a skeleton-style scaffold for secondary development:

- Strip Grok-specific business logic.
- Keep a **runnable** generic framework (not empty folders only).
- Frontend becomes a **component + page example gallery**, based on the existing admin UI (not a minimal empty shell).
- Document structure and conventions in `AGENTS.md` + focused docs.
- After all cuts, rename packages/binaries/branding to Foam.

### Success criteria

1. `make run` boots backend + serves frontend.
2. Admin login works (bootstrap admin from config).
3. Frontend routes under `/example/*` render without Grok/account-pool APIs.
4. Backend exposes a thin `example` CRUD API as the extension template.
5. `AGENTS.md` explains structure and how to develop efficiently.
6. Module path is `github.com/Rain-kl/Foam/backend`; binary is `foam`.
7. No Grok/provider/gateway business code remains in the runnable path.

## 2. Approach

**In-place trim + thin example module (Approach 1).**

1. Delete business implementations; keep layering conventions.
2. Rewrite `internal/app` as a thin composition root.
3. Add one vertical slice: `example` (domain → application → repository → HTTP).
4. Reframe frontend as example gallery (reuse existing pages/components).
5. Directory READMEs for removed business areas.
6. Global rename only after the skeleton runs.

Rejected alternatives:

- Empty-directory-only scaffold (not runnable).
- Rewrite from scratch (discards reusable auth/middleware/persistence).
- Rename before trim (wasteful churn on deleted files).

## 3. Backend

### 3.1 Keep (generic)

| Path | Role |
| --- | --- |
| `backend/cmd/foam` | Process entry (renamed from `grok2api`) |
| `backend/internal/cli` | Flags, `--config`, version |
| `backend/internal/app` | DI / lifecycle (**rewrite thin**) |
| `backend/internal/infra/config` | YAML load/validate; drop Grok-only fields |
| `backend/internal/infra/security` | JWT, password hash, credential cipher helpers |
| `backend/internal/infra/persistence` | SQLite / Postgres + GORM |
| `backend/internal/infra/runtime` | memory/redis store abstraction (simplify if unused by example) |
| `backend/internal/infra/observability` | Logging helpers worth keeping |
| `backend/internal/shared/response` | Unified API envelope |
| `backend/internal/transport/http/middleware` | Auth, timeouts, request ID, etc. |
| `backend/internal/transport/http/adminauth` | Login / refresh / me |
| `backend/internal/transport/http/system` | Health / version |
| `backend/internal/application/adminauth` | Admin auth use cases |
| `backend/internal/domain/admin` | Admin entity |
| `backend/internal/buildinfo` | Version injection |
| `backend/internal/pkg/*` | Keep only non-business utilities still referenced |

### 3.2 Add

Full vertical slice **`example`**:

- `domain/example`
- `application/example`
- `repository` interface + relational implementation
- `transport/http/example`

Minimal fields for demo: `id`, `name`, `description`, `created_at`, `updated_at`.  
CRUD: list / get / create / update / delete under admin JWT.

### 3.3 Delete (business)

Remove implementations under (non-exhaustive; anything Grok/gateway-specific):

- `application/{account,accountsync,audit,clientkey,dashboard,egress,gateway,media,model,quotarecovery,updatecheck,invalidation,settings}`
- Matching `domain/*` (except `admin` + `example`)
- `infra/provider/**`, `infra/egress/**`, `infra/media/**`
- Matching `transport/http/*` handlers
- Matching repository files
- Swagger business annotations; Grok routing/media/provider config sections

### 3.4 Directory placeholders

Where a layer convention should remain visible after deletion, keep a short `README.md` describing:

- Layer responsibility
- What belongs here
- How to copy the `example` slice

Do not leave broken imports or stub packages that fail `go build`.

### 3.5 Runnable API surface

```
CLI → config → DB migrate/bootstrap → HTTP server
  POST /api/admin/login
  POST /api/admin/refresh   (if already present)
  GET  /api/admin/me
  GET  /api/system/health
  GET  /api/system/version
  CRUD /api/examples        (admin JWT)
  Static: frontend/dist
```

### 3.6 Config skeleton

Retain: `server`, `auth`, `secrets`, `bootstrapAdmin`, `frontend`, `database`, `runtimeStore` (if still used).  
Remove or neutralize: Grok routing, provider pools, media pipeline, cluster fields only needed for multi-replica gateway, sponsor-related anything.

Default names/prefixes become `foam` after rename phase.

## 4. Frontend

### 4.1 Product shape

Not a minimal CRUD shell. Rebuild the admin as an **example gallery**:

- Preserve layout shell, shadcn UI, table/dialog/chart patterns from the original console.
- Neutralize copy (no Grok / account pool / API gateway product language).
- Merge duplicate components before showcasing them.
- Prefer **mock/static data** for gallery pages; only auth + health (+ optional example CRUD) hit the real backend.

### 4.2 Routes (final)

| Path | Content |
| --- | --- |
| `/login` | Login (neutral copy) |
| `/example/component` | **Single page** sections: buttons, cards, table primitives, dialogs/sheets, tabs, gallery, forms, charts |
| `/example/page/dashboard` | Former `/dashboard` as ops overview page template |
| `/example/page/table` | Full list page template (filters + table + pagination + bulk actions) |
| `/example/page/chat` | Former `/creative-console` as chat/creative page template |

Default post-login route: `/example/page/dashboard`.  
Old business paths removed or redirected into the above.

### 4.3 Code layout

```
frontend/src/
  components/ui/              # Atomic shadcn components (no business copy)
  shared/
    api/ auth/ hooks/ lib/ config/
    components/               # PageHeader, DataTableShell, Pagination, ...
  features/
    auth/                     # Login
    example/
      component-page.tsx      # /example/component
      pages/
        dashboard-page.tsx
        table-page.tsx
        chat-page.tsx
  app/                        # providers, auth-boundary, app-shell, router
```

Navigation:

- Example → Component
- Example → Page → Dashboard / Table / Chat

### 4.4 Transformation rules

1. Base work on existing feature pages; do not redesign from zero.
2. Merge duplicated UI into `components/ui` or `shared/components`, then display on `/example/component`.
3. Replace business labels with generic demo labels.
4. Strip live dependency on accounts/models/gateway APIs; mock where needed.
5. `features/example/pages/table` may optionally call backend `/api/examples` as a live CRUD demo.

### 4.5 Delete / stop shipping

- Product features no longer routed: accounts, audits, client-keys, models, media galleries as product pages, API docs product, settings product surface, version update checks tied to Grok releases, sponsors assets.
- Branding assets (`grok2api` images, sponsor folder) replaced with Foam placeholders.

## 5. AGENTS.md and docs

Use progressive disclosure.

### `AGENTS.md` (always-on router)

- What Foam is
- Always-on rules: layering, extend via `example` copy, neutral copy, evidence before “done”
- Module/binary: `github.com/Rain-kl/Foam/backend`, `foam`
- Index table pointing to docs

### On-demand docs

| Doc | When |
| --- | --- |
| `docs/architecture.md` | Overall layout, request flow, how to add a module |
| `docs/backend-development.md` | Backend layers, example slice, tests |
| `docs/frontend-development.md` | Routes, example pages, component reuse |
| `docs/directory-map.md` | Per-directory purpose and expected contents |

Optional short `README.md` files inside important empty/convention directories.

## 6. Rename phase (after cuts compile and smoke)

| Item | From | To |
| --- | --- | --- |
| Go module | `github.com/chenyme/grok2api/backend` | `github.com/Rain-kl/Foam/backend` |
| Binary / cmd | `cmd/grok2api` | `cmd/foam` |
| Frontend package | `grok2api-frontend` | `foam-frontend` |
| Docker / compose names | grok2api | foam |
| Config defaults (prefix, cluster) | grok2api | foam |
| README / product name | Grok2API | Foam |

Rename is a dedicated last pass (search-replace imports, Makefile, Dockerfile, compose, docs).

## 7. Implementation order

1. **Backend trim** — remove business packages; thin `app` + config; keep admin auth/system; add `example` CRUD; `go build` / smoke login.
2. **Frontend gallery** — rewire router/shell; build `/example/component` and three page templates from existing UI; mock data; neutralize copy.
3. **Placeholders** — directory READMEs; slim `config.example.yaml`; Docker/compose still name-agnostic until rename.
4. **Docs** — `AGENTS.md` + architecture / backend / frontend / directory-map.
5. **Rename** — module, binary, package names, image names, defaults.
6. **Verify** — build backend, build frontend, login, open all four example routes, optional example CRUD.

## 8. Out of scope

- New design system beyond existing shadcn/tokens
- Re-implementing Grok gateway / multi-provider routing
- Publishing as a separate npm/Go library (scaffold stays monorepo app)
- Multi-tenant product features

## 9. Risks and mitigations

| Risk | Mitigation |
| --- | --- |
| `app` wiring too entangled to trim safely | Rewrite composition root rather than comment-out; delete dead imports aggressively |
| Frontend pages tightly coupled to APIs | Introduce mock adapters per example page first |
| Over-deletion of shared UI | Inventory `shared/components` and `components/ui` before deleting features |
| Rename misses strings | Final `rg grok2api|chenyme|Grok` gate before claiming done |

## 10. Open decisions (resolved)

- Scaffold depth: **A skeleton (runnable)**
- Frontend: **Admin gallery, not pure component library only**
- Component demos: **single route `/example/component`**
- Page demos: **dashboard / table / chat under `/example/page/*`**
- Approach: **in-place trim + example slice**
- Go module: **`github.com/Rain-kl/Foam/backend`**
