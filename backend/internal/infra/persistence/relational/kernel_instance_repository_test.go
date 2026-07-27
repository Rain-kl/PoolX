package relational_test

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/domain/clash"
	"github.com/Rain-kl/Foam/backend/internal/infra/persistence/relational"
)

func TestKernelInstanceRepository_UpsertWithPID(t *testing.T) {
	ctx := context.Background()
	db, err := relational.OpenSQLite(ctx, filepath.Join(t.TempDir(), "kernel.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.InitializeSchema(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	repo := relational.NewKernelInstanceRepository(db)
	pid := 127
	now := time.Now()
	inst := &clash.KernelInstance{
		KernelType:           "mihomo",
		Status:               clash.KernelInstanceStatusRunning,
		PID:                  &pid,
		WorkDir:              "/app/data/runtime",
		ConfigPath:           "/app/data/runtime/config.yaml",
		ControllerAddress:    "127.0.0.1:9090",
		ActiveConfigChecksum: "abc",
		ActiveProfileCount:   1,
		ActiveListenerCount:  1,
		LastAction:           "start",
		LastStartedAt:        &now,
	}

	if err := repo.Upsert(ctx, inst); err != nil {
		t.Fatalf("upsert running instance: %v", err)
	}
	if inst.ID <= 0 {
		t.Fatalf("expected positive id after upsert, got %d", inst.ID)
	}

	got, err := repo.GetByType(ctx, "mihomo")
	if err != nil {
		t.Fatalf("get by type: %v", err)
	}
	if got.Status != clash.KernelInstanceStatusRunning {
		t.Fatalf("status: got %q want running", got.Status)
	}
	if got.PID == nil || *got.PID != pid {
		t.Fatalf("pid: got %v want %d", got.PID, pid)
	}
	if got.WorkDir != "/app/data/runtime" {
		t.Fatalf("work_dir: got %q", got.WorkDir)
	}

	// Second upsert (stop) must update the same row, including clearing pid.
	stopped := &clash.KernelInstance{
		KernelType: "mihomo",
		Status:     clash.KernelInstanceStatusStopped,
		WorkDir:    "/app/data/runtime",
		ConfigPath: "/app/data/runtime/config.yaml",
		LastAction: "stop",
	}
	if err := repo.Upsert(ctx, stopped); err != nil {
		t.Fatalf("upsert stopped instance: %v", err)
	}
	if stopped.ID != inst.ID {
		t.Fatalf("expected same id on conflict update: first=%d second=%d", inst.ID, stopped.ID)
	}

	got2, err := repo.GetByType(ctx, "mihomo")
	if err != nil {
		t.Fatalf("get after stop: %v", err)
	}
	if got2.Status != clash.KernelInstanceStatusStopped {
		t.Fatalf("status after stop: got %q", got2.Status)
	}
	if got2.PID != nil {
		t.Fatalf("pid after stop: want nil, got %v", *got2.PID)
	}
}
