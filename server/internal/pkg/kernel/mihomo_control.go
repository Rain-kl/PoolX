package kernel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"
)

var MihomoHTTPClient = &http.Client{
	Timeout: 5 * time.Second,
}

func StartMihomoProcess(binaryPath string, workDir string, configPath string, stdout io.Writer, stderr io.Writer) (*exec.Cmd, error) {
	cmd := exec.Command(binaryPath, "-d", workDir, "-f", configPath)
	cmd.Dir = workDir
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
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
	defer resp.Body.Close()
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
