package service

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"poolx/internal/model"
	"poolx/internal/pkg/common"
	"runtime"
	"strings"
	"time"
)

const (
	KernelTypeMihomo  = "mihomo"
	KernelTypeXray    = "xray"
	KernelTypeSingbox = "singbox"

	MihomoBinarySourceExisting = "existing"
	MihomoBinarySourceUpload   = "upload"
	MihomoBinarySourceDownload = "download"
)

var kernelHTTPClient = &http.Client{
	Timeout: 2 * time.Minute,
}

var execCommandContext = exec.CommandContext

type InstalledKernelBinary struct {
	KernelType      string    `json:"kernel_type"`
	InstallPath     string    `json:"install_path"`
	BinarySource    string    `json:"binary_source"`
	DetectedVersion string    `json:"detected_version"`
	FileName        string    `json:"file_name"`
	ReleaseTag      string    `json:"release_tag,omitempty"`
	InstalledAt     time.Time `json:"installed_at"`
}

func InspectMihomoBinary(ctx context.Context, installPath string) (*InstalledKernelBinary, error) {
	resolvedPath, err := resolveExecutableInstallPath(installPath, false)
	if err != nil {
		return nil, err
	}
	if _, err = os.Stat(resolvedPath); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("未找到 Mihomo 二进制文件: %s", resolvedPath)
		}
		return nil, fmt.Errorf("检查 Mihomo 二进制文件失败: %v", err)
	}

	detectedVersion, err := detectMihomoBinaryVersion(ctx, resolvedPath)
	if err != nil {
		return nil, err
	}

	return persistInstalledMihomoBinary(resolvedPath, filepath.Base(resolvedPath), MihomoBinarySourceExisting, detectedVersion, "")
}

func InstallUploadedMihomoBinary(ctx context.Context, fileName string, installPath string, reader io.Reader) (*InstalledKernelBinary, error) {
	resolvedPath, err := resolveExecutableInstallPath(installPath, true)
	if err != nil {
		return nil, err
	}
	tempPath, err := persistExecutableTempFile(filepath.Dir(resolvedPath), "poolx-mihomo-upload-", fileName, reader)
	if err != nil {
		return nil, err
	}
	return installPreparedMihomoBinary(ctx, tempPath, resolvedPath, strings.TrimSpace(fileName), MihomoBinarySourceUpload, "")
}

func DownloadAndInstallMihomoBinary(ctx context.Context, installPath string) (*InstalledKernelBinary, error) {
	resolvedPath, err := resolveExecutableInstallPath(installPath, true)
	if err != nil {
		return nil, err
	}
	release, err := fetchLatestMihomoRelease(ctx)
	if err != nil {
		return nil, err
	}
	asset, err := selectMihomoReleaseAsset(release.Assets)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(defaultContext(ctx), http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Mihomo 下载请求失败: %v", err)
	}
	req.Header.Set("Accept", "application/octet-stream")
	req.Header.Set("User-Agent", "PoolX-Server")

	resp, err := kernelHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("下载 Mihomo 二进制失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("下载 Mihomo 二进制失败: %s", resp.Status)
	}

	reader := io.Reader(resp.Body)
	if strings.HasSuffix(strings.ToLower(asset.Name), ".gz") {
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("解压 Mihomo 发行包失败: %v", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	tempPath, err := persistExecutableTempFile(filepath.Dir(resolvedPath), "poolx-mihomo-download-", asset.Name, reader)
	if err != nil {
		return nil, err
	}
	return installPreparedMihomoBinary(ctx, tempPath, resolvedPath, asset.Name, MihomoBinarySourceDownload, release.TagName)
}

func fetchLatestMihomoRelease(ctx context.Context) (*githubReleaseResponse, error) {
	url := fmt.Sprintf(githubReleasesAPIBase+"/latest", common.DefaultMihomoReleaseRepo)
	req, err := http.NewRequestWithContext(defaultContext(ctx), http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建 Mihomo 版本查询请求失败: %v", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "PoolX-Server")

	resp, err := kernelHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("查询 Mihomo 官方版本失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询 Mihomo 官方版本失败: %s", resp.Status)
	}

	var release githubReleaseResponse
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析 Mihomo 版本响应失败: %v", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return nil, fmt.Errorf("未获取到有效的 Mihomo 版本标签")
	}
	return &release, nil
}

func selectMihomoReleaseAsset(assets []githubAsset) (*githubAsset, error) {
	if len(assets) == 0 {
		return nil, fmt.Errorf("Mihomo 官方发布未包含可下载资产")
	}

	normalizedAssets := make([]githubAsset, 0, len(assets))
	for _, asset := range assets {
		name := strings.ToLower(strings.TrimSpace(asset.Name))
		if name == "" || strings.TrimSpace(asset.BrowserDownloadURL) == "" {
			continue
		}
		if strings.Contains(name, "sha256") || strings.HasSuffix(name, ".txt") {
			continue
		}
		normalizedAssets = append(normalizedAssets, asset)
	}

	goos := strings.ToLower(runtime.GOOS)
	for _, archCandidate := range preferredMihomoArchKeywords() {
		for _, asset := range normalizedAssets {
			name := strings.ToLower(strings.TrimSpace(asset.Name))
			if !strings.Contains(name, "mihomo-") || !strings.Contains(name, goos) {
				continue
			}
			if strings.Contains(name, archCandidate) {
				candidate := asset
				return &candidate, nil
			}
		}
	}

	return nil, fmt.Errorf("未找到适用于当前平台 %s/%s 的 Mihomo 发行包", runtime.GOOS, runtime.GOARCH)
}

func preferredMihomoArchKeywords() []string {
	switch runtime.GOARCH {
	case "amd64":
		return []string{"amd64-v3", "amd64-v2", "amd64-v1", "amd64-compatible", "amd64"}
	case "arm64":
		return []string{"arm64"}
	case "386":
		return []string{"386"}
	case "arm":
		goarm := strings.TrimSpace(os.Getenv("GOARM"))
		keywords := make([]string, 0, 4)
		if goarm != "" {
			keywords = append(keywords, "armv"+goarm)
		}
		return append(keywords, "armv7", "armv6", "arm")
	default:
		return []string{runtime.GOARCH}
	}
}

func installPreparedMihomoBinary(ctx context.Context, tempPath string, installPath string, fileName string, source string, releaseTag string) (*InstalledKernelBinary, error) {
	detectedVersion, err := detectMihomoBinaryVersion(ctx, tempPath)
	if err != nil {
		_ = os.Remove(tempPath)
		return nil, err
	}

	if strings.TrimSpace(releaseTag) != "" && !mihomoVersionMatchesRelease(detectedVersion, releaseTag) {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("Mihomo 版本校验失败：release=%s，binary=%s", strings.TrimSpace(releaseTag), strings.TrimSpace(detectedVersion))
	}

	if err = os.MkdirAll(filepath.Dir(installPath), 0o755); err != nil {
		_ = os.Remove(tempPath)
		return nil, fmt.Errorf("创建 Mihomo 目标目录失败: %v", err)
	}
	if err = replaceExecutableFile(tempPath, installPath); err != nil {
		_ = os.Remove(tempPath)
		return nil, err
	}

	return persistInstalledMihomoBinary(installPath, fileName, source, detectedVersion, releaseTag)
}

func resolveExecutableInstallPath(installPath string, allowDefault bool) (string, error) {
	resolved := strings.TrimSpace(installPath)
	if resolved == "" {
		if !allowDefault {
			return "", fmt.Errorf("请先填写 Mihomo 二进制文件路径")
		}
		resolved = defaultMihomoInstallPath()
	}
	absolutePath, err := filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("解析 Mihomo 二进制路径失败: %v", err)
	}
	if runtime.GOOS == "windows" && filepath.Ext(absolutePath) == "" {
		absolutePath += ".exe"
	}
	if strings.TrimSpace(filepath.Base(absolutePath)) == "" {
		return "", fmt.Errorf("Mihomo 二进制文件路径无效")
	}
	return absolutePath, nil
}

func defaultMihomoInstallPath() string {
	fileName := "mihomo"
	if runtime.GOOS == "windows" {
		fileName += ".exe"
	}
	return filepath.Join(".", "data", fileName)
}

func persistExecutableTempFile(tempDir string, patternPrefix string, fileName string, reader io.Reader) (string, error) {
	suffix := filepath.Ext(strings.TrimSpace(fileName))
	if runtime.GOOS == "windows" && suffix == "" {
		suffix = ".exe"
	}
	if strings.TrimSpace(tempDir) == "" {
		tempDir = os.TempDir()
	}
	tempFile, err := os.CreateTemp(tempDir, patternPrefix+"*"+suffix)
	if err != nil {
		return "", fmt.Errorf("创建临时二进制文件失败: %v", err)
	}
	tempPath := tempFile.Name()
	if _, err = io.Copy(tempFile, reader); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("写入临时二进制文件失败: %v", err)
	}
	if err = tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("关闭临时二进制文件失败: %v", err)
	}
	if err = os.Chmod(tempPath, 0o755); err != nil && runtime.GOOS != "windows" {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("设置临时二进制文件权限失败: %v", err)
	}
	return tempPath, nil
}

func detectCommandVersion(ctx context.Context, filePath string, args ...string) (string, error) {
	commandCtx := defaultContext(ctx)
	cmd := execCommandContext(commandCtx, filePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(string(output)))
	}
	for _, line := range strings.Split(string(output), "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed, nil
		}
	}
	return "", fmt.Errorf("命令未返回有效版本号")
}

func detectMihomoBinaryVersion(ctx context.Context, filePath string) (string, error) {
	var lastErr error
	for _, args := range [][]string{{"-v"}, {"version"}, {"--version"}} {
		version, err := detectCommandVersion(ctx, filePath, args...)
		if err != nil {
			lastErr = err
			continue
		}
		if strings.Contains(strings.ToLower(version), "mihomo") {
			return version, nil
		}
		lastErr = fmt.Errorf("上传文件未表现为 Mihomo 二进制，版本输出为: %s", version)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("未能检测到 Mihomo 版本")
	}
	return "", fmt.Errorf("检查 Mihomo 二进制版本失败: %v", lastErr)
}

func persistInstalledMihomoBinary(installPath string, fileName string, source string, detectedVersion string, releaseTag string) (*InstalledKernelBinary, error) {
	if err := model.UpdateOption("KernelType", KernelTypeMihomo); err != nil {
		return nil, err
	}
	if err := model.UpdateOption("MihomoBinaryPath", installPath); err != nil {
		return nil, err
	}
	if err := model.UpdateOption("MihomoBinaryVersion", detectedVersion); err != nil {
		return nil, err
	}
	if err := model.UpdateOption("MihomoBinarySource", source); err != nil {
		return nil, err
	}

	return &InstalledKernelBinary{
		KernelType:      KernelTypeMihomo,
		InstallPath:     installPath,
		BinarySource:    source,
		DetectedVersion: detectedVersion,
		FileName:        strings.TrimSpace(fileName),
		ReleaseTag:      strings.TrimSpace(releaseTag),
		InstalledAt:     time.Now(),
	}, nil
}

func mihomoVersionMatchesRelease(version string, releaseTag string) bool {
	normalizedVersion := strings.ToLower(strings.TrimSpace(version))
	normalizedTag := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(releaseTag, "v")))
	if normalizedVersion == "" || normalizedTag == "" {
		return false
	}
	return strings.Contains(normalizedVersion, normalizedTag)
}

func replaceExecutableFile(tempPath string, targetPath string) error {
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("清理现有 Mihomo 二进制失败: %v", err)
	}
	if err := os.Rename(tempPath, targetPath); err != nil {
		return fmt.Errorf("安装 Mihomo 二进制失败: %v", err)
	}
	if err := os.Chmod(targetPath, 0o755); err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("设置 Mihomo 二进制权限失败: %v", err)
	}
	return nil
}

func defaultContext(ctx context.Context) context.Context {
	if ctx != nil {
		return ctx
	}
	return context.Background()
}
