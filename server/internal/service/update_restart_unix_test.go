//go:build !windows

package service

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

func TestReplaceFileUnixFallsBackOnCrossDeviceRename(t *testing.T) {
	tempDir := t.TempDir()
	srcPath := filepath.Join(tempDir, "source.bin")
	dstPath := filepath.Join(tempDir, "target.bin")

	if err := os.WriteFile(srcPath, []byte("new-binary"), 0o755); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}
	if err := os.WriteFile(dstPath, []byte("old-binary"), 0o755); err != nil {
		t.Fatalf("failed to write target file: %v", err)
	}

	originalRename := unixRename
	unixRename = func(oldPath string, newPath string) error {
		if oldPath == srcPath && newPath == dstPath {
			return &os.LinkError{Op: "rename", Old: oldPath, New: newPath, Err: syscall.EXDEV}
		}
		return os.Rename(oldPath, newPath)
	}
	t.Cleanup(func() {
		unixRename = originalRename
	})

	if err := replaceFileUnix(srcPath, dstPath); err != nil {
		t.Fatalf("expected cross-device fallback to succeed: %v", err)
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}
	if string(content) != "new-binary" {
		t.Fatalf("unexpected target content: %s", string(content))
	}
	if _, err = os.Stat(srcPath); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("expected source file to be removed, got err=%v", err)
	}
}
