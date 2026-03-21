package service

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"poolx/internal/pkg/common"
)

func TestValidateKernelTypeRejectsReservedKernel(t *testing.T) {
	err := validateKernelType(KernelTypeXray)
	if err == nil {
		t.Fatal("expected xray to be rejected for now")
	}
}

func TestSelectMihomoReleaseAssetPrefersCurrentPlatform(t *testing.T) {
	assets := []githubAsset{
		{Name: "mihomo-linux-arm64-v1.0.0.gz", BrowserDownloadURL: "https://example.com/arm64"},
		{Name: fmt.Sprintf("mihomo-%s-%s-v1.0.0.gz", runtime.GOOS, preferredMihomoArchKeywords()[0]), BrowserDownloadURL: "https://example.com/current"},
	}

	asset, err := selectMihomoReleaseAsset(assets)
	if err != nil {
		t.Fatalf("expected asset selection to succeed: %v", err)
	}
	if asset.BrowserDownloadURL != "https://example.com/current" {
		t.Fatalf("unexpected asset selected: %s", asset.BrowserDownloadURL)
	}
}

func TestSelectMihomoReleaseAssetSkipsPackageFormats(t *testing.T) {
	assets := []githubAsset{
		{Name: fmt.Sprintf("mihomo-%s-%s-v1.0.0.deb", runtime.GOOS, preferredMihomoArchKeywords()[0]), BrowserDownloadURL: "https://example.com/package"},
		{Name: fmt.Sprintf("mihomo-%s-%s-v1.0.0.gz", runtime.GOOS, preferredMihomoArchKeywords()[0]), BrowserDownloadURL: "https://example.com/binary"},
	}

	asset, err := selectMihomoReleaseAsset(assets)
	if err != nil {
		t.Fatalf("expected asset selection to succeed: %v", err)
	}
	if asset.BrowserDownloadURL != "https://example.com/binary" {
		t.Fatalf("expected gzip binary asset, got %s", asset.BrowserDownloadURL)
	}
}

func TestInstallUploadedMihomoBinary(t *testing.T) {
	setupServiceTestDB(t)

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mihomo")
	fileName, content := testExecutableFile("Mihomo Meta v1.19.21")

	originalKernelType := common.KernelType
	originalPath := common.MihomoBinaryPath
	originalVersion := common.MihomoBinaryVersion
	originalSource := common.MihomoBinarySource
	t.Cleanup(func() {
		common.KernelType = originalKernelType
		common.MihomoBinaryPath = originalPath
		common.MihomoBinaryVersion = originalVersion
		common.MihomoBinarySource = originalSource
	})

	result, err := InstallUploadedMihomoBinary(context.Background(), fileName, targetPath, bytes.NewReader(content))
	if err != nil {
		t.Fatalf("expected Mihomo install to succeed: %v", err)
	}
	if result.KernelType != KernelTypeMihomo {
		t.Fatalf("unexpected kernel type: %s", result.KernelType)
	}
	if !strings.Contains(result.DetectedVersion, "Mihomo") {
		t.Fatalf("unexpected version: %s", result.DetectedVersion)
	}
	if common.MihomoBinaryPath == "" || common.MihomoBinarySource != MihomoBinarySourceUpload {
		t.Fatalf("expected Mihomo options to be persisted")
	}
	if _, err := os.Stat(result.InstallPath); err != nil {
		t.Fatalf("expected installed binary at %s: %v", result.InstallPath, err)
	}
}

func TestInspectMihomoBinary(t *testing.T) {
	setupServiceTestDB(t)

	targetDir := t.TempDir()
	targetPath := filepath.Join(targetDir, "mihomo-existing")
	fileName, content := testExecutableFile("Mihomo Meta v1.19.22")
	if err := os.WriteFile(targetPath, content, 0o755); err != nil {
		t.Fatalf("write existing binary: %v", err)
	}
	if runtime.GOOS == "windows" && !strings.HasSuffix(targetPath, ".exe") {
		renamedPath := targetPath + ".exe"
		if err := os.Rename(targetPath, renamedPath); err != nil {
			t.Fatalf("rename existing binary: %v", err)
		}
		targetPath = renamedPath
	}
	fileName = filepath.Base(targetPath)

	result, err := InspectMihomoBinary(context.Background(), targetPath)
	if err != nil {
		t.Fatalf("expected inspect to succeed: %v", err)
	}
	if result.BinarySource != MihomoBinarySourceExisting {
		t.Fatalf("unexpected source: %s", result.BinarySource)
	}
	if result.FileName != fileName {
		t.Fatalf("unexpected file name: %s", result.FileName)
	}
}

func TestResolveExecutableInstallPathUsesWorkingDirectoryDefault(t *testing.T) {
	resolved, err := resolveExecutableInstallPath("", true)
	if err != nil {
		t.Fatalf("resolve executable install path: %v", err)
	}

	expectedPath, err := filepath.Abs(filepath.Join(".", "mihomo"))
	if err != nil {
		t.Fatalf("resolve expected path: %v", err)
	}
	if runtime.GOOS == "windows" {
		expectedPath += ".exe"
	}
	if resolved != expectedPath {
		t.Fatalf("unexpected default install path: got %s want %s", resolved, expectedPath)
	}
}

func testExecutableFile(version string) (string, []byte) {
	if runtime.GOOS == "windows" {
		return "mihomo-test.cmd", []byte("@echo off\r\necho " + version + "\r\n")
	}
	return "mihomo-test.sh", []byte("#!/bin/sh\nif [ \"$1\" = \"-v\" ]; then\n  echo " + version + "\n  exit 0\nfi\nif [ \"$1\" = \"version\" ]; then\n  echo " + version + "\n  exit 0\nfi\necho " + version + "\n")
}
