# Foam Frontend

Admin SPA for the Foam full-stack scaffold. Includes auth, dashboard, and UI / page templates.

## Stack

- React 19 + TypeScript
- Vite 8 + Tailwind CSS
- shadcn/ui + Radix UI
- TanStack Query、React Hook Form、Zod

## Local development (HMR + API proxy)

No frontend build is required for day-to-day debugging. Run backend and Vite side by side:

```bash
# terminal 1 — API on :8000 (does not need frontend/dist)
cd backend && go run ./cmd/foam --config ../config.yaml

# terminal 2 — SPA + HMR on :8010, proxies API to backend
cd frontend
pnpm install
pnpm dev
```

Open `http://127.0.0.1:8010`. Vite proxies `/api`, `/healthz`, `/readyz`, and `/swagger` to `http://127.0.0.1:8000`.

API envelope (Wavelet-aligned): `{ "error_msg": "", "data": ... }`. Auth under `/api/v1/user/*`, admin under `/api/v1/admin/*`.

Other backend target:

```bash
VITE_DEV_API_TARGET=http://127.0.0.1:9000 pnpm dev
```

## Production / embed build

```bash
pnpm build
```

Output is `dist/`. Production still uses **embed mode**: the Go process serves the SPA via `frontend.staticPath` (default `./frontend/dist`) with SPA fallback. The frontend does not read YAML; public runtime values come from controlled backend endpoints / `runtime-config.js`.

## Format

```bash
pnpm format         # biome format --write .
pnpm format:check   # biome format .
```

## Layout

```text
src/app/             routes and app shell
src/features/        feature pages
src/shared/          API, auth, config, components, utils
src/components/ui/   shadcn/ui primitives
```

## Verify

```bash
pnpm lint
pnpm format:check
pnpm build
```
