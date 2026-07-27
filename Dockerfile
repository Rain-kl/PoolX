ARG NODE_VERSION=22
ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.23

FROM --platform=$BUILDPLATFORM node:${NODE_VERSION}-alpine AS frontend-builder

WORKDIR /src/frontend
RUN corepack enable

COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN --mount=type=cache,id=foam-pnpm,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm fetch --frozen-lockfile

RUN --mount=type=cache,id=foam-pnpm,target=/pnpm/store \
    pnpm config set store-dir /pnpm/store && \
    pnpm install --offline --frozen-lockfile

COPY frontend/index.html frontend/vite.config.ts frontend/tsconfig.json frontend/tsconfig.app.json frontend/tsconfig.node.json ./
COPY frontend/public ./public
COPY frontend/src ./src
RUN --mount=type=cache,id=foam-tsc,target=/src/frontend/.cache,sharing=locked \
    pnpm build


FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS backend-builder

ARG TARGETOS
ARG TARGETARCH
# Release metadata: tag (vX.Y.Z[-ext]) or canary when untagged.
ARG VERSION=canary
ARG BUILD_DATE=

WORKDIR /src/backend
RUN apk add --no-cache ca-certificates git

COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,id=foam-go-mod,target=/go/pkg/mod,sharing=locked \
    go mod download

COPY backend/cmd ./cmd
COPY backend/internal ./internal
COPY backend/zashboard ./zashboard
COPY backend/docs/docs.go ./docs/docs.go
RUN --mount=type=cache,id=foam-go-mod,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,id=foam-go-build,target=/root/.cache/go-build,sharing=locked \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -buildvcs=false -trimpath \
      -ldflags="-s -w -X 'github.com/Rain-kl/Foam/backend/internal/buildinfo.Version=${VERSION}' -X 'github.com/Rain-kl/Foam/backend/internal/buildinfo.BuildTime=${BUILD_DATE}'" \
      -o /out/poolx ./cmd/foam


# Download latest mihomo for TARGETARCH (runs on build platform; asset selected by arch).
FROM --platform=$BUILDPLATFORM alpine:${ALPINE_VERSION} AS mihomo-downloader
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
    chmod 0755 /out/mihomo


FROM alpine:${ALPINE_VERSION}

ENV TZ=Asia/Shanghai \
    FOAM_CONFIG_SOURCE=/run/foam/config.yaml \
    FOAM_CLASH_MIHOMO_BINARY_PATH=/opt/mihomo

RUN apk add --no-cache ca-certificates su-exec tzdata && \
    addgroup -S -g 10001 foam && \
    adduser -S -D -H -u 10001 -G foam foam && \
    mkdir -p /app/data /run/foam && \
    chown -R foam:foam /app/data /run/foam

WORKDIR /app

ARG VERSION=canary
COPY --from=backend-builder --chmod=0755 /out/poolx /app/poolx
COPY --from=frontend-builder /src/frontend/dist /app/frontend/dist
COPY --from=mihomo-downloader --chmod=0755 /out/mihomo /opt/mihomo
# Prefer ldflags-injected Version; keep VERSION file as runtime fallback.
RUN printf '%s\n' "${VERSION}" > /app/VERSION
COPY --chmod=0755 docker/entrypoint.sh /usr/local/bin/foam-entrypoint

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8000/healthz >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/foam-entrypoint"]
# config.yaml is optional; when absent, all settings come from FOAM_* env vars.
CMD ["/app/poolx", "--listen", "0.0.0.0:8000"]
