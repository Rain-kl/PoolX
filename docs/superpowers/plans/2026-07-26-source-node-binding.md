# Source-Node Binding Implementation Plan

> **For agentic workers:** Implement task-by-task. Steps use checkbox syntax.

**Goal:** Bind imported nodes to source configs; cascade-delete nodes (and related rows) when a source is deleted; full delete+insert sync on confirm/refresh/reupload.

**Architecture:** Application-layer cascade (no physical FKs). Shared `SyncSourceNodes` kernel; `ProxyNodeRepository.DeleteBySourceConfigID` cleans bindings + tests + nodes in a TX. Global fingerprint skip for cross-source duplicates.

**Tech Stack:** Go, GORM, Gin, existing clash slice, React frontend.

**Spec:** `docs/superpowers/specs/2026-07-26-source-node-binding-design.md`

## Files

- Modify: `backend/internal/repository/clash.go`
- Modify: `backend/internal/infra/persistence/relational/clash_repository.go`
- Modify: `backend/internal/application/clash/service.go`
- Modify: `backend/internal/application/clash/service_test.go`
- Modify: `backend/internal/transport/http/clash/handler.go`
- Modify: `frontend/src/shared/api/clash.ts`
- Modify: `frontend/src/features/clash/source-configs/source-configs-page.tsx`

## Tasks

### Task 1: Repository cascade APIs
- Add `ListIDsBySourceConfigID`, `DeleteBySourceConfigID` on `ProxyNodeRepository`
- Implement cascade TX (port_profile_nodes → node_test_results → proxy_nodes)

### Task 2: Application SyncSourceNodes + Delete/Confirm/Refresh/Reupload
- Shared sync kernel (delete own nodes → fingerprint skip → insert → update source)
- Confirm always syncs; Delete cascades; Refresh/Reupload with type guards

### Task 3: HTTP handlers
- Routes: refresh, reupload; confirm response still `imported_nodes`

### Task 4: Tests
- Cascade delete, re-confirm full replace, cross-source skip, type guards

### Task 5: Frontend
- Delete warning; refresh / reupload actions

### Task 6: Verify
- `cd backend && go test ./internal/application/clash/... ./internal/transport/http/clash/...`
