<p align="right">
  <a href="./README_CN.md">中文</a>
</p>

<div align="center">

# PoolX

A Proxy Pool Control Plane built with Gin and Next.js, designed to turn Clash/Mihomo nodes into reusable proxy pools for crawlers, scraping systems, and proxy-driven network workloads.
</div>

## What This Project Is For

PoolX is built for teams that need to:

* import and manage large sets of Clash-compatible proxy nodes
* organize those nodes into reusable proxy pools
* expose stable local proxy entrypoints for crawlers and automation services
* apply load balancing, fallback, and latency-based selection strategies
* control Mihomo as the execution kernel through a web-based control plane

In short, PoolX helps convert raw proxy subscriptions and node lists into operational proxy pools that can be consumed by scraping projects, bots, data collection services, and other outbound request systems.

## Features

* Web-based admin console with authentication, settings, logs, and file upload support
* Node import pipeline for Clash-compatible configurations
* Node pool management with testing and reuse across multiple workspaces
* Workspace-based port profile management for building proxy pool entrypoints
* Support for `Mixed` or `SOCKS/HTTP` listener modes
* Runtime aggregation that combines multiple workspace profiles into a final Mihomo configuration
* Built-in strategy support: `select`, `url-test`, `fallback`, and `load-balance`
* JSON-based proxy settings extension for listener auth, UDP, latency test options, and load-balance tuning
* Mihomo binary management from the settings page
* Built-in zashboard served at `/zashboard/`, proxied through PoolX auth instead of exposing the Clash secret to the browser

## Typical Use Cases

* Proxy pools for crawlers and scraping services
* Outbound proxy gateways for automation platforms
* Rotating or fallback proxy entrypoints for data collection jobs
* Workspace-based proxy orchestration for multi-target scraping strategies
* Internal control panels for teams operating Mihomo-backed proxy infrastructure

## Quick Start

Login URL: `http://localhost:3000/`
Username: `root`
Password: `123456`

### Docker Deployment

```yaml
services:
  postgres:
    image: postgres:17-alpine
    restart: unless-stopped
    environment:
      POSTGRES_DB: poolx
      POSTGRES_USER: poolx
      POSTGRES_PASSWORD: replace-with-strong-password
    volumes:
      - ./data/postgres:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U poolx -d poolx"]
      interval: 10s
      timeout: 5s
      retries: 5

  poolx:
    image: ghcr.io/rain-kl/poolx:latest
    restart: unless-stopped
    depends_on:
      postgres:
        condition: service_healthy
    ports:
      - "3000:3000"
#     开放代理监听端口
    environment:
      SESSION_SECRET: replace-with-random-string
      SQLITE_PATH: /data/poolx.db
      DSN: postgres://poolx:replace-with-strong-password@postgres:5432/poolx?sslmode=disable
      GIN_MODE: release
      LOG_LEVEL: info

    volumes:
      - ./data/poolx:/data
```

### Local Deployment

#### Requirements

* Go 1.24+
* Node.js 20+
* pnpm

#### Build Frontend Assets

```bash
cd server/web
corepack enable
pnpm install
pnpm build
```

```bash
cd server/zashboard
corepack enable
pnpm install
pnpm build
```

#### Run the Server

```bash
cd server
export SESSION_SECRET='replace-with-random-string'
export SQLITE_PATH='./poolx.db'
export LOG_LEVEL='info'
go run ./cmd/server
```

App URL: `http://localhost:3000`

Built-in Clash dashboard: `http://localhost:3000/zashboard/`

Default account:

* Username: `root`
* Password: `123456`

## Development

Server:

```bash
cd server
go run ./cmd/server
```

Frontend:

```bash
cd server/web
pnpm install
pnpm dev
```

Zashboard:

```bash
cd server/zashboard
pnpm install
pnpm build
```

Server tests:

```bash
cd server
go test ./...
```

Frontend build:

```bash
cd server/web
pnpm build
```

## Configuration

PoolX does not expect you to hand-maintain a single static Mihomo config file.

Instead, it:

* imports nodes
* organizes them into pools
* lets you define workspace listener profiles
* renders the final Mihomo runtime config automatically

For runtime options, deployment, and application config, see:

* [docs/app-config.md](./docs/app-config.md)
* [docs/deployment.md](./docs/deployment.md)


## License

This project is released under the [Apache License 2.0](./LICENSE).
