<p align="right">
  <a href="./README_CN.md">中文</a>
</p>

<div align="center">

# PoolX

A proxy pool control panel built with Gin and Next.js. Its core goal is to organize Clash/Mihomo nodes into reusable proxy pools for web scraping, data collection, proxy requests, and automated outbound network scenarios.

</div>

> [!NOTE]
> The node-to-proxy-pool approach based on glider is deprecated. The legacy version is archived in the `glider` branch. This project now provides a full graphical interface system for building proxy pools, currently supporting only the Mihomo core.

> [!WARNING]
> This project is only a proxy pool control panel and does NOT provide nodes or methods to obtain them. It is intended for learning, research, and technical exchange only. Any illegal use is strictly prohibited.

> [!WARNING]
> After logging in with the root account for the first time, be sure to change the default password `123456` immediately!

---

## Introduction

When making multiple requests to target websites using crawlers, anti-scraping mechanisms are often triggered, leading to access bans. These bans are typically based on IP addresses, so switching IPs is an effective way to bypass restrictions and maintain stable scraping.

A common solution is to use a proxy pool. However, existing solutions have drawbacks:

- Free proxy pools often have poor quality, with low stability and availability
- Paid proxy services can be expensive and exceed practical needs

A more cost-effective approach is to utilize nodes provided by proxy service providers (“airports”), combined with open-source proxy cores, to build reusable proxy pools.

This project provides a clean and user-friendly UI for efficiently managing and organizing proxy nodes, eliminating the need to manually write or maintain complex configuration files. It also offers unified core control for centralized management of the proxy engine.

### Core Features

- Import and manage a large number of nodes via configuration files or subscription URLs
- Organize nodes into reusable proxy pools
- Provide stable local proxy endpoints for crawlers, automation tasks, and proxy systems
- Support load balancing, automatic fallback, and latency-based node selection
- Visual dashboard for managing core runtime status

---

## Features

- Web-based management panel with authentication and system settings
- Node pool management with testing, filtering, and cross-workspace reuse
- Workspace-based port configuration for building proxy pool entry points
- Multiple port listeners, each mapped to a different proxy pool configuration
- Built-in zashboard for extended monitoring and management of the core

---

## Typical Use Cases

- Proxy pools for web scraping and crawling services
- Unified outbound proxy gateway for automation platforms
- Rotating or fallback proxy entry points for data collection tasks
- Workspace-based proxy orchestration for different target sites
- Proxy access for AI services to reduce risk and improve stability

---

## Quick Start

Login URL: `http://IP:3000/`  
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
#     Expose proxy listening ports as needed
    environment:
      SESSION_SECRET: replace-with-random-string
#     If SQL_DSN is specified, the following will be ignored
      SQLITE_PATH: /data/poolx.db
#     To use SQLite, comment out SQL_DSN and remove postgres service
      SQL_DSN: postgres://poolx:replace-with-strong-password@postgres:5432/poolx?sslmode=disable
      GIN_MODE: release
      LOG_LEVEL: info

    volumes:
      - ./data/poolx:/data
```

### Local Deployment

Download the precompiled binary from the Release page:

```bash
# Start with SQLite
SESSION_SECRET=replace-with-random-string \
SQLITE_PATH=/path/to/poolx.db \
./poolx
```

Access URL: http://localhost:3000

Default credentials:
* Username: root
* Password: 123456


## Configuration

PoolX does not require maintaining a fixed configuration file manually.

Its core workflow is:
* Import nodes
* Organize proxy pools
* Define workspace listening configurations
* Automatically render the final runtime configuration for the core

For runtime parameters, deployment methods, and system configuration, refer to:
* [docs/app-config.md](./docs/app-config.md)
* [docs/deployment.md](./docs/deployment.md)


## License

This project is licensed under the [Apache License 2.0](./LICENSE) ￼.
