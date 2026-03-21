package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"
	"testing"
	"time"

	"poolx/internal/model"
	"poolx/internal/pkg/common"
)

func TestStartRuntimeRejectsOccupiedListenerPort(t *testing.T) {
	setupServiceTestDB(t)
	resetRuntimeRegistryForTest()

	originalKernelType := common.KernelType
	originalBinaryPath := common.MihomoBinaryPath
	originalStarter := startMihomoProcess
	common.KernelType = KernelTypeMihomo
	common.MihomoBinaryPath = "/tmp/fake-mihomo"
	startCalled := false
	startMihomoProcess = func(binaryPath string, workDir string, configPath string, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
		startCalled = true
		return nil, assertiveError("should not start process when listener is occupied")
	}
	t.Cleanup(func() {
		common.KernelType = originalKernelType
		common.MihomoBinaryPath = originalBinaryPath
		startMihomoProcess = originalStarter
		resetRuntimeRegistryForTest()
	})

	node := seedRuntimeTestNode(t, "occupied-port-node")

	occupied, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("occupy local port: %v", err)
	}
	t.Cleanup(func() {
		_ = occupied.Close()
	})

	port := occupied.Addr().(*net.TCPAddr).Port
	if _, err := CreatePortProfile(PortProfilePayload{
		Name:                "occupied-port-profile",
		ListenHost:          "127.0.0.1",
		MixedPort:           port,
		StrategyType:        model.PortProfileStrategyFallback,
		StrategyGroupName:   "POOLX",
		TestURL:             "https://cp.cloudflare.com/generate_204",
		TestIntervalSeconds: 300,
		IncludeInRuntime:    true,
		NodeIDs:             []int{node.ID},
	}); err != nil {
		t.Fatalf("create port profile: %v", err)
	}

	_, err = StartRuntime(context.Background())
	if err == nil {
		t.Fatal("expected occupied listener port to fail")
	}
	if !strings.Contains(err.Error(), "监听端口已被占用") {
		t.Fatalf("expected occupied port error, got %v", err)
	}
	if !strings.Contains(err.Error(), net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port))) {
		t.Fatalf("expected occupied address in error, got %v", err)
	}
	if startCalled {
		t.Fatal("expected process start to be skipped when listener port is occupied")
	}
}

func TestStartRuntimeIncludesRecentLogsWhenControllerWaitFails(t *testing.T) {
	setupServiceTestDB(t)
	resetRuntimeRegistryForTest()

	originalKernelType := common.KernelType
	originalBinaryPath := common.MihomoBinaryPath
	originalStarter := startMihomoProcess
	originalWaiter := waitForMihomoControllerReady
	common.KernelType = KernelTypeMihomo
	common.MihomoBinaryPath = "/tmp/fake-mihomo"
	startMihomoProcess = func(binaryPath string, workDir string, configPath string, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
		cmd := exec.Command("/bin/sh", "-c", "echo boot failed 1>&2; exit 1")
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		return cmd, nil
	}
	waitForMihomoControllerReady = func(ctx context.Context, controllerAddress string, secret string) error {
		time.Sleep(150 * time.Millisecond)
		return context.DeadlineExceeded
	}
	t.Cleanup(func() {
		common.KernelType = originalKernelType
		common.MihomoBinaryPath = originalBinaryPath
		startMihomoProcess = originalStarter
		waitForMihomoControllerReady = originalWaiter
		resetRuntimeRegistryForTest()
	})

	node := seedRuntimeTestNode(t, "controller-wait-node")
	freeListener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve local port: %v", err)
	}
	freePort := freeListener.Addr().(*net.TCPAddr).Port
	_ = freeListener.Close()
	if _, err := CreatePortProfile(PortProfilePayload{
		Name:                "controller-wait-profile",
		ListenHost:          "127.0.0.1",
		MixedPort:           freePort,
		SocksPort:           0,
		HTTPPort:            0,
		StrategyType:        model.PortProfileStrategyFallback,
		StrategyGroupName:   "POOLX",
		TestURL:             "https://cp.cloudflare.com/generate_204",
		TestIntervalSeconds: 300,
		IncludeInRuntime:    true,
		NodeIDs:             []int{node.ID},
	}); err != nil {
		t.Fatalf("create port profile: %v", err)
	}

	_, err = StartRuntime(context.Background())
	if err == nil {
		t.Fatal("expected controller wait failure")
	}
	if !strings.Contains(err.Error(), "等待 Mihomo 控制接口就绪失败") {
		t.Fatalf("expected controller wait error, got %v", err)
	}
	if !strings.Contains(err.Error(), "boot failed") {
		t.Fatalf("expected recent logs in error, got %v", err)
	}
}

func seedRuntimeTestNode(t *testing.T, name string) *model.ProxyNode {
	t.Helper()

	node := &model.ProxyNode{
		SourceConfigID:   1,
		SourceConfigName: "seed.yaml",
		Name:             name,
		Type:             "vless",
		Server:           "127.0.0.1",
		Port:             443,
		Fingerprint:      "fingerprint-" + name,
		MetadataJSON:     `{"type":"vless","server":"127.0.0.1","port":443,"uuid":"63a636ac-a5b2-4463-b4f9-e983bf4680bf","tls":true,"servername":"learn.microsoft.com","client-fingerprint":"chrome"}`,
		Enabled:          true,
	}
	if err := model.DB.Create(node).Error; err != nil {
		t.Fatalf("seed proxy node: %v", err)
	}
	return node
}

func resetRuntimeRegistryForTest() {
	runtimeRegistry.mu.Lock()
	defer runtimeRegistry.mu.Unlock()
	runtimeRegistry.cmd = nil
	runtimeRegistry.logSeq = 0
	runtimeRegistry.logs = nil
	runtimeRegistry.startedAt = time.Time{}
}
