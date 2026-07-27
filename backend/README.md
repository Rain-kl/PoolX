# Foam Backend

Go backend for the Foam full-stack scaffold. Provides admin auth, example CRUD, health probes, and optional SPA hosting.

## Stack

- Go 1.26, Gin, GORM
- SQLite / PostgreSQL

## Local run

Config lives at the repository root. Create a local config and/or env file:

```bash
cp ../config.example.yaml ../config.yaml
cp ../.env.example ../.env
openssl rand -hex 32
openssl rand -base64 32
```

Priority: **env (`FOAM_*`) > `config.yaml` > defaults** (CLI flags win over env).  
Set secrets in `config.yaml` and/or `.env` (`FOAM_SECRETS_JWT_SECRET`, …), then:

```bash
cd backend
go run ./cmd/foam --config ../config.yaml
```

Default listen: `http://127.0.0.1:8000`. Override config or address as needed:

```bash
go run ./cmd/foam --config /path/to/config.yaml --listen 0.0.0.0:8000
```

From the repository root:

```bash
make run
```

## Endpoints (Wavelet-aligned)

- `GET /api/health` — health (envelope `{error_msg,data}`)
- `/api/v1/user/*` — login / refresh / logout / self / change-password
- `/api/v1/admin/*` — status, settings, examples (admin JWT)
- `GET /api/v1/config/public` — public runtime map
- `/healthz`, `/readyz` — probe aliases (Docker / k8s)
- `frontend.staticPath` — optional SPA static hosting (default `./frontend/dist`)

Response envelope:

```json
{ "error_msg": "", "data": {} }
```

## Layout

```text
cmd/foam/              process entry
internal/domain/       domain models
internal/application/  application services
internal/infra/        config, persistence, security
internal/transport/    HTTP routes and middleware
internal/repository/   persistence interfaces
```

## Verify

```bash
go test ./...
go vet ./...
go build ./cmd/foam
```
