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
