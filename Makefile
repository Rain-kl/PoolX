.PHONY: run swagger format code-check \
	build-backend build-frontend build-embedded build-test cross-build \
	dev dev-f dev-b

# ---------------------------------------------------------------------------
# Project knobs
# ---------------------------------------------------------------------------

CONFIG     ?= $(CURDIR)/config.yaml
MODULE     := github.com/Rain-kl/Foam/backend
BIN_NAME   := foam
BIN_DIR    := $(CURDIR)/bin
BACKEND    := $(CURDIR)/backend
FRONTEND   := $(CURDIR)/frontend
GOCACHE    ?= $(CURDIR)/.gocache

# Version for ldflags: make VERSION=v1.2.3 … overrides.
# Else: git tag at HEAD (v*) → VERSION file → canary-<sha> → dev.
GIT_TAG      := $(shell git tag --points-at HEAD --list 'v*' 2>/dev/null | sort -V | tail -n1)
GIT_SHA      := $(shell git rev-parse --short=7 HEAD 2>/dev/null)
FILE_VERSION := $(shell tr -d '[:space:]' < VERSION 2>/dev/null)

ifeq ($(origin VERSION),command line)
  # keep user-provided VERSION
else ifneq ($(strip $(GIT_TAG)),)
  VERSION := $(GIT_TAG)
else ifneq ($(strip $(FILE_VERSION)),)
  VERSION := $(FILE_VERSION)
else ifneq ($(strip $(GIT_SHA)),)
  VERSION := canary-$(GIT_SHA)
else
  VERSION := dev
endif

BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -s -w \
	-X '$(MODULE)/internal/buildinfo.Version=$(VERSION)' \
	-X '$(MODULE)/internal/buildinfo.BuildTime=$(BUILD_DATE)'

export GOCACHE

# ---------------------------------------------------------------------------
# Day-to-day
# ---------------------------------------------------------------------------

run:
	cd "$(BACKEND)" && go run ./cmd/foam --config "$(abspath $(CONFIG))" $(RUN_ARGS)

swagger:
	cd "$(BACKEND)" && go run github.com/swaggo/swag/cmd/swag@v1.16.6 init \
		-g main.go \
		-d cmd/foam,internal/transport/http \
		--parseInternal \
		--output docs \
		--outputTypes go,json,yaml

# Frontend HMR :8010 (proxies API to :8000) + backend API.
dev:
	@echo "==> Starting frontend (:8010) and backend (:8000) in parallel..."
	@PIDS=""; \
	STATUS=0; \
	trap 'kill $$PIDS 2>/dev/null; wait 2>/dev/null' INT TERM EXIT; \
	( cd "$(FRONTEND)" && pnpm dev 2>&1 | sed 's/^/[frontend] /' ) & PIDS="$$PIDS $$!"; \
	( cd "$(BACKEND)" && go run ./cmd/foam --config "$(abspath $(CONFIG))" $(RUN_ARGS) 2>&1 | sed 's/^/[backend]  /' ) & PIDS="$$PIDS $$!"; \
	for PID in $$PIDS; do \
		wait $$PID || STATUS=1; \
	done; \
	trap - INT TERM EXIT; \
	if [ $$STATUS -eq 0 ]; then \
		echo "==> All development servers exited successfully."; \
	else \
		echo "==> Development servers exited with errors." >&2; \
		exit 1; \
	fi

dev-f:
	@echo "==> Starting frontend development server (http://127.0.0.1:8010)..."
	cd "$(FRONTEND)" && pnpm dev

dev-b:
	@echo "==> Starting backend development server..."
	cd "$(BACKEND)" && go run ./cmd/foam --config "$(abspath $(CONFIG))" $(RUN_ARGS)

# ---------------------------------------------------------------------------
# Format / lint
# ---------------------------------------------------------------------------

format:
	@echo "==> Formatting backend Go source..."
	gofmt -w $$(find "$(BACKEND)" -type f -name '*.go' -not -path '*/.git/*')
	@echo "==> Formatting frontend source..."
	cd "$(FRONTEND)" && pnpm format

code-check:
	@echo "==> golangci-lint (backend)..."
	cd "$(BACKEND)" && golangci-lint run
	@echo "==> Typecheck + eslint (frontend)..."
	cd "$(FRONTEND)" && pnpm exec tsc -b && pnpm exec eslint . --max-warnings 0

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------

# Backend binary only → bin/foam
build-backend:
	@echo "==> Building backend version=$(VERSION) build_date=$(BUILD_DATE)..."
	@mkdir -p "$(BIN_DIR)"
	cd "$(BACKEND)" && go build \
		-buildvcs=false -trimpath \
		-ldflags "$(LDFLAGS)" \
		-o "$(BIN_DIR)/$(BIN_NAME)" \
		./cmd/foam
	@echo "==> Wrote $(BIN_DIR)/$(BIN_NAME)"

# SPA production assets → frontend/dist (served via frontend.staticPath)
build-frontend:
	@echo "==> Building frontend version=$(VERSION) build_date=$(BUILD_DATE)..."
	cd "$(FRONTEND)" && pnpm build
	@echo "==> Wrote $(FRONTEND)/dist"

# Release-style local embed: build SPA then backend.
# Foam hosts dist via config frontend.staticPath (default ./frontend/dist), not go:embed.
build-embedded: build-frontend build-backend
	@echo "==> Embedded layout ready:"
	@echo "    binary:  $(BIN_DIR)/$(BIN_NAME)"
	@echo "    static:  $(FRONTEND)/dist  (frontend.staticPath)"
	@echo "    run e.g. $(BIN_DIR)/$(BIN_NAME) --config $(CONFIG)"

# Parallel smoke: frontend production build + backend test/build.
build-test:
	@echo "==> Running frontend and backend build tests in parallel..."
	@PIDS=""; \
	STATUS=0; \
	( cd "$(FRONTEND)" && pnpm build 2>&1 | sed 's/^/[frontend] /' ) & PIDS="$$PIDS $$!"; \
	( cd "$(BACKEND)" && { go test ./... && go build -buildvcs=false -o /dev/null ./cmd/foam; } 2>&1 | sed 's/^/[backend]  /' ) & PIDS="$$PIDS $$!"; \
	for PID in $$PIDS; do \
		wait $$PID || STATUS=1; \
	done; \
	if [ $$STATUS -eq 0 ]; then \
		echo "==> All build tests passed."; \
	else \
		echo "==> Build test FAILED." >&2; \
		exit 1; \
	fi

# Pure-Go cross compile (no Docker). Override with GOOS=linux GOARCH=amd64 etc.
# Examples:
#   make cross-build
#   make cross-build GOOS=linux GOARCH=amd64
#   make cross-build GOOS="linux darwin" GOARCH="amd64 arm64"
cross-build:
	@echo "==> Cross-compiling \
	$(if $(GOOS),$(GOOS),linux/darwin/windows) × \
	$(if $(GOARCH),$(GOARCH),amd64/arm64) \
	(version=$(VERSION))..."
	@mkdir -p "$(BIN_DIR)"
	@oses="$(if $(GOOS),$(GOOS),linux darwin windows)"; \
	arches="$(if $(GOARCH),$(GOARCH),amd64 arm64)"; \
	STATUS=0; \
	for os in $$oses; do \
		for arch in $$arches; do \
			ext=""; \
			if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
			out="$(BIN_DIR)/$(BIN_NAME)-$${os}-$${arch}$${ext}"; \
			echo "    → $$out"; \
			( cd "$(BACKEND)" && \
				CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
				go build -buildvcs=false -trimpath \
					-ldflags "$(LDFLAGS)" \
					-o "$$out" \
					./cmd/foam ) || STATUS=1; \
		done; \
	done; \
	if [ $$STATUS -ne 0 ]; then \
		echo "==> Cross-build FAILED." >&2; \
		exit 1; \
	fi
	@echo "==> Done. Binaries written to $(BIN_DIR)/"
	@ls -lh "$(BIN_DIR)/"
