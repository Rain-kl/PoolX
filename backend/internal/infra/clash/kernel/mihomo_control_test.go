package kernel_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/Rain-kl/Foam/backend/internal/infra/clash/kernel"
)

// TestStartMihomoProcess_SurvivesCancelledContext ensures the kernel process is
// not bound to the caller context (HTTP request Timeout middleware cancels it).
func TestStartMihomoProcess_SurvivesCancelledContext(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("sleep script uses sh")
	}

	// Fake "mihomo": a shell that sleeps long enough to outlive request cancel.
	binDir := t.TempDir()
	binPath := filepath.Join(binDir, "mihomo")
	script := "#!/bin/sh\nexec sleep 30\n"
	if err := os.WriteFile(binPath, []byte(script), 0o755); err != nil {
		t.Fatalf("write fake binary: %v", err)
	}

	workDir := t.TempDir()
	configPath := filepath.Join(workDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte("mixed-port: 0\n"), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("open devnull: %v", err)
	}
	t.Cleanup(func() { _ = devNull.Close() })

	ctx, cancel := context.WithCancel(context.Background())
	cmd, err := kernel.StartMihomoProcess(ctx, binPath, workDir, configPath, devNull, devNull)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
			_, _ = cmd.Process.Wait()
		}
	})

	// Simulate HTTP request end / Timeout middleware cancel.
	cancel()
	time.Sleep(200 * time.Millisecond)

	// Process must still be alive after context cancel.
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		t.Fatalf("process not alive after context cancel (CommandContext regression?): %v", err)
	}
}

func TestStartMihomoProcess_RespectsCancelledContextBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := kernel.StartMihomoProcess(ctx, "/bin/true", t.TempDir(), filepath.Join(t.TempDir(), "c.yaml"), nil, nil)
	if err == nil {
		t.Fatal("expected error when ctx already cancelled")
	}
}
