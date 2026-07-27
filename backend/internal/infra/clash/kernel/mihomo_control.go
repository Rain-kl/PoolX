package kernel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"time"
)

var MihomoHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}

// StartMihomoProcess launches a long-running mihomo kernel.
//
// IMPORTANT: the process must NOT be bound to the caller's context via
// exec.CommandContext. HTTP handlers pass a request-scoped context that is
// cancelled when the Timeout middleware returns; binding the kernel to that
// context would kill mihomo immediately after "start" succeeds.
//
// ctx is only checked before Start (cancel-before-launch). After Start returns,
// process lifetime is owned by the caller (StopKernel / restart / OS).
//
//nolint:contextcheck
func StartMihomoProcess(ctx context.Context, binaryPath string, workDir string, configPath string, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	resolvedWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		return nil, fmt.Errorf("resolve mihomo workdir: %w", err)
	}
	resolvedConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("resolve mihomo config path: %w", err)
	}
	// Use exec.Command (not CommandContext) so the kernel outlives the request.
	cmd := exec.Command(binaryPath, "-d", resolvedWorkDir, "-f", resolvedConfigPath) //nolint:gosec
	cmd.Dir = resolvedWorkDir
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return cmd, nil
}

func WaitForControllerReady(ctx context.Context, controllerAddress string, secret string) error {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()

	for {
		if err := ProbeController(ctx, controllerAddress, secret); err == nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}

func ProbeController(ctx context.Context, controllerAddress string, secret string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/version", controllerAddress), nil)
	if err != nil {
		return err
	}
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := MihomoHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("controller status: %s", resp.Status)
	}
	return nil
}

func ReloadMihomoConfig(ctx context.Context, controllerAddress string, secret string, path string) error {
	payload, err := json.Marshal(map[string]string{
		"path":    path,
		"payload": "",
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, fmt.Sprintf("http://%s/configs?force=true", controllerAddress), bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := MihomoHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("reload controller config failed: %s", resp.Status)
	}
	return nil
}

func GetMihomoVersion(ctx context.Context, controllerAddress string, secret string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/version", controllerAddress), nil)
	if err != nil {
		return "", err
	}
	if secret != "" {
		req.Header.Set("Authorization", "Bearer "+secret)
	}
	resp, err := MihomoHTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("controller status: %s", resp.Status)
	}
	var payload struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}
	return payload.Version, nil
}
