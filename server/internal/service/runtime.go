package service

import (
	"bufio"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	kernelpkg "poolx/internal/pkg/kernel"
	"poolx/internal/pkg/runtimeconfig"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

const (
	runtimeLogLimit          = 400
	defaultRuntimeController = "127.0.0.1:19090"
	runtimeConfigFileName    = "config.yaml"
)

type RuntimeLogEntry struct {
	Seq       int64     `json:"seq"`
	Stream    string    `json:"stream"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type RuntimeLogList struct {
	Items []*RuntimeLogEntry `json:"items"`
}

type RuntimeStatus struct {
	Instance              *model.KernelInstance           `json:"instance"`
	Running               bool                            `json:"running"`
	APIHealthy            bool                            `json:"api_healthy"`
	APIVersion            string                          `json:"api_version,omitempty"`
	ProfileCount          int                             `json:"profile_count"`
	ListenerCount         int                             `json:"listener_count"`
	Listeners             []runtimeconfig.RuntimeListener `json:"listeners"`
	RenderedConfigPreview string                          `json:"rendered_config_preview,omitempty"`
}

type runtimeProcessRegistry struct {
	mu        sync.Mutex
	cmd       *exec.Cmd
	logSeq    int64
	logs      []*RuntimeLogEntry
	startedAt time.Time
}

var runtimeRegistry = &runtimeProcessRegistry{}
var startMihomoProcess = kernelpkg.StartMihomoProcess
var waitForMihomoControllerReady = kernelpkg.WaitForControllerReady
var reloadMihomoConfig = kernelpkg.ReloadMihomoConfig
var getMihomoVersion = kernelpkg.GetMihomoVersion

func StartRuntime(ctx context.Context) (*RuntimeStatus, error) {
	if strings.TrimSpace(common.KernelType) != KernelTypeMihomo {
		return nil, fmt.Errorf("当前仅支持启动 Mihomo")
	}
	if strings.TrimSpace(common.MihomoBinaryPath) == "" {
		return nil, fmt.Errorf("请先在系统设置中完成 Mihomo 二进制安装或路径校验")
	}

	runtimeRegistry.mu.Lock()
	if runtimeRegistry.cmd != nil && runtimeRegistry.cmd.Process != nil && runtimeRegistry.cmd.ProcessState == nil {
		runtimeRegistry.mu.Unlock()
		return nil, fmt.Errorf("Mihomo 已在运行中")
	}
	runtimeRegistry.mu.Unlock()

	instance, err := ensureKernelInstance()
	if err != nil {
		return nil, err
	}
	secret := strings.TrimSpace(instance.ControllerSecret)
	if secret == "" {
		secret, err = randomSecret()
		if err != nil {
			return nil, err
		}
	}
	rendered, workDir, configPath, secret, err := buildFinalRuntimeConfig(secret, true)
	if err != nil {
		_ = persistRuntimeState(instance, model.KernelInstanceStatusError, "start", err.Error(), nil, nil)
		return nil, err
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建运行目录失败: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(rendered.Content), 0o644); err != nil {
		return nil, fmt.Errorf("写入运行配置失败: %v", err)
	}
	if err := ensureRuntimeListenersAvailable(rendered.Listeners); err != nil {
		instance.Status = model.KernelInstanceStatusError
		instance.LastError = err.Error()
		_ = model.DB.Save(instance).Error
		return nil, err
	}

	now := time.Now()
	instance.WorkDir = workDir
	instance.ConfigPath = configPath
	instance.ControllerAddress = defaultRuntimeController
	instance.ControllerSecret = secret
	instance.ActiveConfigChecksum = rendered.Checksum
	instance.ActiveProfileCount = rendered.ProfileCount
	instance.ActiveListenerCount = rendered.ListenerCount
	instance.LastAction = "start"
	instance.LastError = ""
	instance.LastStartedAt = &now
	instance.Status = model.KernelInstanceStatusStarting
	if err := model.DB.Save(instance).Error; err != nil {
		return nil, err
	}

	stdoutWriter := runtimeRegistry.attachLogWriter("stdout")
	stderrWriter := runtimeRegistry.attachLogWriter("stderr")
	cmd, err := startMihomoProcess(common.MihomoBinaryPath, workDir, configPath, stdoutWriter, stderrWriter)
	if err != nil {
		instance.Status = model.KernelInstanceStatusError
		instance.LastError = fmt.Sprintf("启动 Mihomo 失败: %v", err)
		_ = model.DB.Save(instance).Error
		return nil, fmt.Errorf("启动 Mihomo 失败: %v", err)
	}
	pid := cmd.Process.Pid
	instance.PID = &pid
	runtimeRegistry.mu.Lock()
	runtimeRegistry.cmd = cmd
	runtimeRegistry.startedAt = now
	runtimeRegistry.mu.Unlock()
	go runtimeRegistry.waitProcessExit(cmd)

	readyCtx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	if err := waitForMihomoControllerReady(readyCtx, defaultRuntimeController, secret); err != nil {
		_ = terminateProcess(cmd)
		waitErr := formatRuntimeStartWaitError(err)
		instance.Status = model.KernelInstanceStatusError
		instance.LastError = waitErr.Error()
		_ = model.DB.Save(instance).Error
		return nil, waitErr
	}

	instance.Status = model.KernelInstanceStatusRunning
	instance.LastError = ""
	if err := model.DB.Save(instance).Error; err != nil {
		return nil, err
	}
	_ = AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, fmt.Sprintf("Mihomo 已启动，监听 %d 个入口。", rendered.ListenerCount))
	return GetRuntimeStatus(context.Background())
}

func StopRuntime(ctx context.Context) (*RuntimeStatus, error) {
	_ = ctx
	instance, err := ensureKernelInstance()
	if err != nil {
		return nil, err
	}
	instance.Status = model.KernelInstanceStatusStopping
	instance.LastAction = "stop"
	instance.LastError = ""
	if err := model.DB.Save(instance).Error; err != nil {
		return nil, err
	}

	runtimeRegistry.mu.Lock()
	cmd := runtimeRegistry.cmd
	runtimeRegistry.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		if err := terminateProcess(cmd); err != nil {
			instance.Status = model.KernelInstanceStatusError
			instance.LastError = fmt.Sprintf("停止 Mihomo 失败: %v", err)
			_ = model.DB.Save(instance).Error
			return nil, fmt.Errorf("停止 Mihomo 失败: %v", err)
		}
	} else if instance.PID != nil {
		process, findErr := os.FindProcess(*instance.PID)
		if findErr == nil {
			_ = process.Kill()
		}
	}

	now := time.Now()
	instance.Status = model.KernelInstanceStatusStopped
	instance.PID = nil
	instance.LastStoppedAt = &now
	instance.LastError = ""
	if err := model.DB.Save(instance).Error; err != nil {
		return nil, err
	}

	runtimeRegistry.mu.Lock()
	runtimeRegistry.cmd = nil
	runtimeRegistry.startedAt = time.Time{}
	runtimeRegistry.mu.Unlock()

	_ = AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, "Mihomo 已停止。")
	return GetRuntimeStatus(context.Background())
}

func ReloadRuntime(ctx context.Context) (*RuntimeStatus, error) {
	instance, err := ensureKernelInstance()
	if err != nil {
		return nil, err
	}
	if instance.PID == nil {
		return nil, fmt.Errorf("Mihomo 当前未运行，无法热重载")
	}

	rendered, workDir, configPath, secret, err := buildFinalRuntimeConfig(instance.ControllerSecret, true)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(instance.ActiveConfigChecksum) == rendered.Checksum {
		return GetRuntimeStatus(ctx)
	}
	if err := os.MkdirAll(workDir, 0o755); err != nil {
		return nil, fmt.Errorf("创建运行目录失败: %v", err)
	}
	previous, _ := os.ReadFile(configPath)
	if err := os.WriteFile(configPath, []byte(rendered.Content), 0o644); err != nil {
		return nil, fmt.Errorf("写入运行配置失败: %v", err)
	}

	instance.Status = model.KernelInstanceStatusReloading
	instance.LastAction = "reload"
	instance.LastError = ""
	instance.WorkDir = workDir
	instance.ConfigPath = configPath
	instance.ControllerAddress = defaultRuntimeController
	instance.ControllerSecret = secret
	if err := model.DB.Save(instance).Error; err != nil {
		return nil, err
	}

	reloadCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()
	if err := reloadMihomoConfig(reloadCtx, instance.ControllerAddress, instance.ControllerSecret, filepath.Base(configPath)); err != nil {
		if len(previous) > 0 {
			_ = os.WriteFile(configPath, previous, 0o644)
		}
		instance.Status = model.KernelInstanceStatusError
		instance.LastError = fmt.Sprintf("热重载失败: %v", err)
		_ = model.DB.Save(instance).Error
		return nil, fmt.Errorf("热重载失败: %v", err)
	}

	now := time.Now()
	instance.Status = model.KernelInstanceStatusRunning
	instance.ActiveConfigChecksum = rendered.Checksum
	instance.ActiveProfileCount = rendered.ProfileCount
	instance.ActiveListenerCount = rendered.ListenerCount
	instance.LastReloadedAt = &now
	instance.LastError = ""
	if err := model.DB.Save(instance).Error; err != nil {
		return nil, err
	}
	_ = AppLog.Push(model.AppLogClassificationBusiness, model.AppLogLevelInfo, fmt.Sprintf("Mihomo 已热重载，启用 %d 个入口。", rendered.ListenerCount))
	return GetRuntimeStatus(context.Background())
}

func GetRuntimeStatus(ctx context.Context) (*RuntimeStatus, error) {
	instance, err := ensureKernelInstance()
	if err != nil {
		return nil, err
	}
	rendered, _, _, _, renderErr := buildFinalRuntimeConfig(instance.ControllerSecret, false)

	apiHealthy := false
	apiVersion := ""
	if strings.TrimSpace(instance.ControllerAddress) != "" && strings.TrimSpace(instance.ControllerSecret) != "" {
		version, err := getMihomoVersion(ctx, instance.ControllerAddress, instance.ControllerSecret)
		if err == nil {
			apiHealthy = true
			apiVersion = version
		}
	}

	runtimeRegistry.mu.Lock()
	running := runtimeRegistry.cmd != nil && runtimeRegistry.cmd.Process != nil && runtimeRegistry.cmd.ProcessState == nil
	runtimeRegistry.mu.Unlock()
	if !running && instance.Status == model.KernelInstanceStatusRunning && apiHealthy {
		running = true
	}

	status := &RuntimeStatus{
		Instance:   instance,
		Running:    running,
		APIHealthy: apiHealthy,
		APIVersion: apiVersion,
	}
	if renderErr == nil {
		status.ProfileCount = rendered.ProfileCount
		status.ListenerCount = rendered.ListenerCount
		status.Listeners = rendered.Listeners
		status.RenderedConfigPreview = rendered.Content
	}
	return status, nil
}

func GetRuntimeLogs(afterSeq int64, limit int) (*RuntimeLogList, error) {
	runtimeRegistry.mu.Lock()
	defer runtimeRegistry.mu.Unlock()
	if limit <= 0 {
		limit = 100
	}
	if limit > 300 {
		limit = 300
	}
	items := make([]*RuntimeLogEntry, 0, limit)
	for _, item := range runtimeRegistry.logs {
		if item.Seq <= afterSeq {
			continue
		}
		items = append(items, item)
		if len(items) >= limit {
			break
		}
	}
	return &RuntimeLogList{Items: items}, nil
}

func ensureKernelInstance() (*model.KernelInstance, error) {
	instance, err := model.GetKernelInstanceByType(common.KernelType)
	if err == nil {
		return instance, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	workDir := filepath.Join("data", "runtime", common.KernelType)
	configPath := filepath.Join(workDir, runtimeConfigFileName)
	instance = &model.KernelInstance{
		KernelType:        common.KernelType,
		Status:            model.KernelInstanceStatusStopped,
		WorkDir:           workDir,
		ConfigPath:        configPath,
		ControllerAddress: defaultRuntimeController,
		ControllerSecret:  "",
	}
	if err := model.DB.Create(instance).Error; err != nil {
		return nil, err
	}
	return instance, nil
}

func buildFinalRuntimeConfig(existingSecret string, persistSnapshots bool) (*runtimeconfig.FinalRenderResult, string, string, string, error) {
	profiles, err := ListPortProfiles()
	if err != nil {
		return nil, "", "", "", err
	}
	enabled := make([]*model.PortProfileWithNodes, 0, len(profiles))
	for _, profile := range profiles {
		if profile.Profile.Enabled {
			enabled = append(enabled, profile)
		}
	}
	if len(enabled) == 0 {
		return nil, "", "", "", fmt.Errorf("当前没有启用的端口配置，无法启动 Mihomo")
	}
	workDir := filepath.Join("data", "runtime", common.KernelType)
	configPath := filepath.Join(workDir, runtimeConfigFileName)
	secret := strings.TrimSpace(existingSecret)
	if secret == "" {
		secret = "preview-secret"
	}
	rendered, err := runtimeconfig.RenderFinalMihomoConfig(runtimeconfig.AggregatedMihomoInput{
		Profiles:          enabled,
		ControllerAddress: defaultRuntimeController,
		ControllerSecret:  secret,
	})
	if err != nil {
		return nil, "", "", "", err
	}
	if persistSnapshots {
		for _, profile := range enabled {
			if _, err := SaveRuntimePreview(profile.Profile.ID); err != nil {
				return nil, "", "", "", err
			}
		}
	}
	return rendered, workDir, configPath, secret, nil
}

func randomSecret() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("生成控制密钥失败: %v", err)
	}
	return hex.EncodeToString(buffer), nil
}

func persistRuntimeState(instance *model.KernelInstance, status string, action string, lastError string, pid *int, rendered *runtimeconfig.FinalRenderResult) error {
	instance.Status = status
	instance.LastAction = action
	instance.LastError = lastError
	instance.PID = pid
	if rendered != nil {
		instance.ActiveConfigChecksum = rendered.Checksum
		instance.ActiveProfileCount = rendered.ProfileCount
		instance.ActiveListenerCount = rendered.ListenerCount
	}
	return model.DB.Save(instance).Error
}

func (registry *runtimeProcessRegistry) attachLogWriter(stream string) io.Writer {
	reader, writer := io.Pipe()
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			message := strings.TrimSpace(scanner.Text())
			if message == "" {
				continue
			}
			level := model.AppLogLevelInfo
			lower := strings.ToLower(message)
			if strings.Contains(lower, "error") || strings.Contains(lower, "fatal") {
				level = model.AppLogLevelError
			} else if strings.Contains(lower, "warn") {
				level = model.AppLogLevelWarn
			}
			registry.appendLog(stream, level, message)
		}
	}()
	return writer
}

func (registry *runtimeProcessRegistry) appendLog(stream string, level string, message string) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	registry.logSeq++
	entry := &RuntimeLogEntry{
		Seq:       registry.logSeq,
		Stream:    stream,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now(),
	}
	registry.logs = append(registry.logs, entry)
	if len(registry.logs) > runtimeLogLimit {
		registry.logs = registry.logs[len(registry.logs)-runtimeLogLimit:]
	}
}

func (registry *runtimeProcessRegistry) waitProcessExit(cmd *exec.Cmd) {
	err := cmd.Wait()
	registry.mu.Lock()
	defer registry.mu.Unlock()
	if registry.cmd == cmd {
		registry.cmd = nil
	}
	if err != nil {
		registry.logSeq++
		registry.logs = append(registry.logs, &RuntimeLogEntry{
			Seq:       registry.logSeq,
			Stream:    "system",
			Level:     model.AppLogLevelError,
			Message:   fmt.Sprintf("Mihomo 进程退出: %v", err),
			CreatedAt: time.Now(),
		})
		if len(registry.logs) > runtimeLogLimit {
			registry.logs = registry.logs[len(registry.logs)-runtimeLogLimit:]
		}
	}
	instance, getErr := model.GetKernelInstanceByType(common.KernelType)
	if getErr == nil {
		now := time.Now()
		instance.PID = nil
		instance.LastStoppedAt = &now
		if err != nil {
			instance.Status = model.KernelInstanceStatusError
			instance.LastError = err.Error()
		} else {
			instance.Status = model.KernelInstanceStatusStopped
			instance.LastError = ""
		}
		_ = model.DB.Save(instance).Error
	}
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	_ = cmd.Process.Signal(os.Interrupt)
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cmd.ProcessState != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return cmd.Process.Kill()
}

func ensureRuntimeListenersAvailable(listeners []runtimeconfig.RuntimeListener) error {
	for _, listener := range listeners {
		address := net.JoinHostPort(listener.Listen, fmt.Sprintf("%d", listener.Port))
		probe, err := net.Listen("tcp", address)
		if err != nil {
			return fmt.Errorf("监听端口已被占用，无法启动 Mihomo: %s (%s)", address, listener.Name)
		}
		_ = probe.Close()
	}
	return nil
}

func formatRuntimeStartWaitError(waitErr error) error {
	message := fmt.Sprintf("等待 Mihomo 控制接口就绪失败: %v", waitErr)
	logs := runtimeRegistry.recentLogSummary(8)
	if logs == "" {
		return fmt.Errorf("%s", message)
	}
	return fmt.Errorf("%s | 最近日志: %s", message, logs)
}

func (registry *runtimeProcessRegistry) recentLogSummary(limit int) string {
	registry.mu.Lock()
	defer registry.mu.Unlock()

	if limit <= 0 || len(registry.logs) == 0 {
		return ""
	}
	start := len(registry.logs) - limit
	if start < 0 {
		start = 0
	}

	messages := make([]string, 0, len(registry.logs)-start)
	for _, entry := range registry.logs[start:] {
		messages = append(messages, fmt.Sprintf("[%s] %s", entry.Stream, entry.Message))
	}
	return strings.Join(messages, " || ")
}
