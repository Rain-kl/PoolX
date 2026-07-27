# Mihomo Docker Packaging Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** `docker build` 时下载最新 mihomo 到 `/opt/mihomo`，镜像固定 `ENV FOAM_CLASH_MIHOMO_BINARY_PATH=/opt/mihomo`；代码默认路径可由该 env 覆盖。

**Architecture:** kernel 包提供 `DefaultMihomoBinaryPath()` 读 `FOAM_CLASH_MIHOMO_BINARY_PATH`，否则 `./data/core/mihomo`；clash/settings 所有空路径回退走该 helper；Dockerfile 新增 mihomo-downloader stage。

**Tech Stack:** Go 1.26、Docker multi-stage、MetaCubeX/mihomo GitHub Releases API

## Global Constraints

- 始终 latest release，不 pin 版本
- 镜像路径 `/opt/mihomo`；不进 data 卷
- compose **不**写该 env（镜像 ENV 固定）
- 本地默认仍 `./data/core/mihomo`
- Spec: `docs/superpowers/specs/2026-07-27-mihomo-docker-design.md`

---

### Task 1: DefaultMihomoBinaryPath helper + tests

**Files:**
- Create: `backend/internal/infra/clash/kernel/default_path_test.go`
- Modify: `backend/internal/infra/clash/kernel/installer.go`
- Modify: `backend/internal/infra/config/env.go`

**Interfaces:**
- Produces: `func DefaultMihomoBinaryPath() string`
- Produces: `config.EnvClashMihomoBinaryPath = "FOAM_CLASH_MIHOMO_BINARY_PATH"`
- Consumes: `os.Getenv`

- [ ] **Step 1: Write failing tests**

```go
package kernel_test

import (
	"testing"

	"github.com/Rain-kl/Foam/backend/internal/infra/clash/kernel"
	"github.com/Rain-kl/Foam/backend/internal/infra/config"
)

func TestDefaultMihomoBinaryPath_FallsBackToDataCore(t *testing.T) {
	t.Setenv(config.EnvClashMihomoBinaryPath, "")
	if got := kernel.DefaultMihomoBinaryPath(); got != "./data/core/mihomo" {
		t.Fatalf("got %q, want ./data/core/mihomo", got)
	}
}

func TestDefaultMihomoBinaryPath_UsesEnv(t *testing.T) {
	t.Setenv(config.EnvClashMihomoBinaryPath, "/opt/mihomo")
	if got := kernel.DefaultMihomoBinaryPath(); got != "/opt/mihomo" {
		t.Fatalf("got %q, want /opt/mihomo", got)
	}
}

func TestDefaultMihomoBinaryPath_TrimsEnv(t *testing.T) {
	t.Setenv(config.EnvClashMihomoBinaryPath, "  /opt/mihomo  ")
	if got := kernel.DefaultMihomoBinaryPath(); got != "/opt/mihomo" {
		t.Fatalf("got %q, want /opt/mihomo", got)
	}
}
```

- [ ] **Step 2: Run tests — expect FAIL**

```bash
cd backend && go test ./internal/infra/clash/kernel/ -run DefaultMihomo -v
```

Expected: undefined `DefaultMihomoBinaryPath` / missing const

- [ ] **Step 3: Implement**

In `env.go` add:

```go
EnvClashMihomoBinaryPath = "FOAM_CLASH_MIHOMO_BINARY_PATH"
```

In `installer.go` add:

```go
func DefaultMihomoBinaryPath() string {
	if v := strings.TrimSpace(os.Getenv(config.EnvClashMihomoBinaryPath)); v != "" {
		return v
	}
	return "./data/core/mihomo"
}
```

Also change `ResolveProjectDataPath` empty fallback to call `DefaultMihomoBinaryPath()`.

Import `config` package in kernel (or avoid cycle: hardcode env name string `"FOAM_CLASH_MIHOMO_BINARY_PATH"` in kernel to avoid infra→config coupling if needed). **Prefer hardcoding the env key string in kernel** matching the const value, and export const only from config for documentation/tests — tests can use the string literal or config const. If kernel importing config is clean (config has no kernel dep), use config const.

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd backend && go test ./internal/infra/clash/kernel/ -run DefaultMihomo -v
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/infra/clash/kernel/ backend/internal/infra/config/env.go
git commit -m "feat(clash): add DefaultMihomoBinaryPath from FOAM_ env"
```

---

### Task 2: Wire clash + settings defaults

**Files:**
- Modify: `backend/internal/application/clash/service.go`
- Modify: `backend/internal/application/settings/service.go`

**Interfaces:**
- Consumes: `kernel.DefaultMihomoBinaryPath()`

- [ ] **Step 1: Replace hardcoded defaults**

In `service.go` (clash):
- Remove `const defaultBinaryPath = "./data/core/mihomo"`
- All former uses → `kernelctrl.DefaultMihomoBinaryPath()`

In `settings/service.go`:
- Empty `MihomoBinaryPath` → `kernel.DefaultMihomoBinaryPath()` (add import)

- [ ] **Step 2: Run tests**

```bash
cd backend && go test ./internal/application/clash/... ./internal/application/settings/... ./internal/infra/clash/kernel/ -count=1
```

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add backend/internal/application/clash/service.go backend/internal/application/settings/service.go
git commit -m "feat(clash): use DefaultMihomoBinaryPath for empty path fallbacks"
```

---

### Task 3: Dockerfile mihomo stage + ENV

**Files:**
- Modify: `Dockerfile`
- Do **not** modify `docker-compose.yml` for this env

- [ ] **Step 1: Add mihomo-downloader stage** (before final stage)

```dockerfile
FROM alpine:${ALPINE_VERSION} AS mihomo-downloader
ARG TARGETARCH
RUN apk add --no-cache ca-certificates curl gzip jq
RUN set -eu; \
    ARCH="${TARGETARCH:-amd64}"; \
    case "$ARCH" in \
      amd64|arm64) ;; \
      *) echo "unsupported TARGETARCH=$ARCH" >&2; exit 1 ;; \
    esac; \
    API="https://api.github.com/repos/MetaCubeX/mihomo/releases/latest"; \
    JSON=$(curl -fsSL -H "Accept: application/vnd.github+json" -H "User-Agent: Foam-Docker-Build" "$API"); \
    if [ "$ARCH" = "amd64" ]; then \
      URL=$(echo "$JSON" | jq -r '[.assets[] | select(.name|test("^mihomo-linux-amd64-compatible-.*\\.gz$"))][0].browser_download_url // empty'); \
      if [ -z "$URL" ]; then \
        URL=$(echo "$JSON" | jq -r '[.assets[] | select(.name|test("^mihomo-linux-amd64-v1-.*\\.gz$"))][0].browser_download_url // empty'); \
      fi; \
      if [ -z "$URL" ]; then \
        URL=$(echo "$JSON" | jq -r '[.assets[] | select(.name|test("^mihomo-linux-amd64-v[0-9.]+\\.gz$"))][0].browser_download_url // empty'); \
      fi; \
    else \
      URL=$(echo "$JSON" | jq -r '[.assets[] | select(.name|test("^mihomo-linux-arm64-.*\\.gz$"))][0].browser_download_url // empty'); \
    fi; \
    if [ -z "$URL" ] || [ "$URL" = "null" ]; then echo "no mihomo asset for $ARCH" >&2; exit 1; fi; \
    echo "Downloading $URL"; \
    mkdir -p /out; \
    curl -fsSL -o /tmp/mihomo.gz "$URL"; \
    gzip -dc /tmp/mihomo.gz > /out/mihomo; \
    chmod 0755 /out/mihomo; \
    /out/mihomo -v || true
```

Final stage additions after WORKDIR/COPY poolx:

```dockerfile
COPY --from=mihomo-downloader --chmod=0755 /out/mihomo /opt/mihomo
ENV FOAM_CLASH_MIHOMO_BINARY_PATH=/opt/mihomo
```

Also add to existing `ENV TZ=...` block or separate ENV line.

- [ ] **Step 2: Verify compose unchanged for this env**

```bash
grep -n FOAM_CLASH_MIHOMO_BINARY_PATH docker-compose.yml || true
```

Expected: no matches

- [ ] **Step 3: Build and smoke (if docker available)**

```bash
docker build -t poolx:mihomo-test .
docker run --rm --entrypoint /bin/sh poolx:mihomo-test -c 'test -x /opt/mihomo && /opt/mihomo -v && printenv FOAM_CLASH_MIHOMO_BINARY_PATH'
```

Expected: version output and `/opt/mihomo`

- [ ] **Step 4: Commit**

```bash
git add Dockerfile
git commit -m "feat(docker): bundle latest mihomo at /opt/mihomo"
```

---

### Task 4: Full verification

- [ ] **Step 1:** `cd backend && go test ./...`
- [ ] **Step 2:** Confirm Dockerfile has COPY + ENV; compose has no FOAM_CLASH_MIHOMO_BINARY_PATH
- [ ] **Step 3:** Done

## Spec coverage checklist

| Spec item | Task |
| --- | --- |
| Build download latest | Task 3 |
| `/opt/mihomo` | Task 3 |
| Image ENV fixed | Task 3 |
| compose no env | Task 3 |
| DefaultMihomoBinaryPath | Task 1 |
| clash/settings wire | Task 2 |
| data volume untouched | Task 3 (no volume change) |
