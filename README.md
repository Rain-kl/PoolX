<p align="right">
  <a href="./README.zh-CN.md">中文</a>
</p>

<div align="center">

# PoolX

A proxy pool control panel built with Gin and React. Its core goal is to organize Clash/Mihomo nodes into reusable proxy pools for web scraping, data collection, proxy requests, and automated outbound network scenarios.

</div>

> [!NOTE]
> The node-to-proxy-pool approach based on glider is deprecated. The legacy version is archived in the `glider` branch. This project now provides a full graphical interface system for building proxy pools, currently supporting only the Mihomo core.

> [!WARNING]
> This project is only a proxy pool control panel and does NOT provide nodes or methods to obtain them. It is intended for learning, research, and technical exchange only. Any illegal use is strictly prohibited.

> [!WARNING]
> After logging in with the bootstrap admin account for the first time, be sure to change the default password immediately!

---

## Introduction

When making multiple requests to target websites using crawlers, anti-scraping mechanisms are often triggered, leading to access bans. These bans are typically based on IP addresses, so switching IPs is an effective way to bypass restrictions and maintain stable scraping.

A common solution is to use a proxy pool. However, existing solutions have drawbacks:

- Free proxy pools often have poor quality, with low stability and availability
- Paid proxy services can be expensive and exceed practical needs

A more cost-effective approach is to utilize nodes provided by proxy service providers ("airports"), combined with open-source proxy cores, to build reusable proxy pools.

This project provides a clean and user-friendly UI for efficiently managing and organizing proxy nodes, eliminating the need to manually write or maintain complex configuration files. It also offers unified core control for centralized management of the proxy engine.

### Core Features

- Import and manage a large number of nodes via configuration files or subscription URLs
- Organize nodes into reusable proxy pools
- Provide stable local proxy endpoints for crawlers, automation tasks, and proxy systems
- Support load balancing, automatic fallback, and latency-based node selection
- Visual dashboard for managing core runtime status
- Auto-start the kernel on application launch when configuration is available

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

### Docker

```bash
cp config.example.yaml config.yaml
# set secrets.jwtSecret, secrets.credentialEncryptionKey, bootstrapAdmin.password
docker compose up -d
docker compose logs -f foam
```

Default backend listen address: `http://127.0.0.1:8000`

### Local Development

```bash
cp config.example.yaml config.yaml
# set secrets, then:
make dev              # frontend :8010 + backend :8000
```

Or run separately:

```bash
cd backend && go run ./cmd/foam --config ../config.yaml
cd frontend && pnpm install && pnpm dev
```

Dev UI: `http://127.0.0.1:8010` → API proxy → `http://127.0.0.1:8000`

---

## Configuration

PoolX does not require maintaining a fixed configuration file manually.

Its core workflow is:
* Import nodes
* Organize proxy pools
* Define workspace listening configurations
* Automatically render the final runtime configuration for the core

Copy the template and fill in required secrets:

```bash
cp config.example.yaml config.yaml
```

Key fields:
- `secrets.jwtSecret` — at least 32 characters
- `secrets.credentialEncryptionKey` — base64-encoded 32-byte key
- `bootstrapAdmin.password` — initial admin password

For runtime parameters, deployment methods, and system configuration, refer to:
* [docs/architecture.md](./docs/architecture.md)
* [docs/backend-development.md](./docs/backend-development.md)

---

## License

MIT — see [LICENSE](./LICENSE).
