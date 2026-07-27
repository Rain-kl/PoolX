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
      -o /out/foam ./cmd/foam


FROM alpine:${ALPINE_VERSION}

ENV TZ=Asia/Shanghai \
    FOAM_CONFIG_SOURCE=/run/foam/config.yaml

RUN apk add --no-cache ca-certificates su-exec tzdata && \
    addgroup -S -g 10001 foam && \
    adduser -S -D -H -u 10001 -G foam foam && \
    mkdir -p /app/data /run/foam && \
    chown -R foam:foam /app/data /run/foam

WORKDIR /app

ARG VERSION=canary
COPY --from=backend-builder --chmod=0755 /out/foam /app/foam
COPY --from=frontend-builder /src/frontend/dist /app/frontend/dist
# Prefer ldflags-injected Version; keep VERSION file as runtime fallback.
RUN printf '%s\n' "${VERSION}" > /app/VERSION
COPY --chmod=0755 docker/entrypoint.sh /usr/local/bin/foam-entrypoint

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8000/healthz >/dev/null || exit 1

ENTRYPOINT ["/usr/local/bin/foam-entrypoint"]
CMD ["/app/foam", "--config", "/app/config.yaml", "--listen", "0.0.0.0:8000"]
