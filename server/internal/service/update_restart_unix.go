//go:build !windows

package service

import (
	"fmt"
	"io"
	"os"
	"syscall"
)

var unixRename = os.Rename

func replaceAndRestartServer(execPath string, tmpPath string) error {
	backupPath := execPath + ".bak"
	_ = os.Remove(backupPath)
	if err := unixRename(execPath, backupPath); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("备份当前服务端二进制失败: %w", err)
	}
	if err := replaceFileUnix(tmpPath, execPath); err != nil {
		_ = unixRename(backupPath, execPath)
		return fmt.Errorf("替换服务端二进制失败: %w", err)
	}
	_ = os.Remove(backupPath)
	if err := syscall.Exec(execPath, os.Args, os.Environ()); err != nil {
		return fmt.Errorf("重启服务失败: %w", err)
	}
	return fmt.Errorf("unreachable after exec")
}

func replaceFileUnix(srcPath string, dstPath string) error {
	if err := unixRename(srcPath, dstPath); err == nil {
		return nil
	} else if linkErr, ok := err.(*os.LinkError); !ok || linkErr.Err != syscall.EXDEV {
		return err
	}

	sourceFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	info, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode().Perm())
	if err != nil {
		return err
	}

	copyErr := func() error {
		defer destinationFile.Close()
		if _, err = io.Copy(destinationFile, sourceFile); err != nil {
			return err
		}
		if err = destinationFile.Sync(); err != nil {
			return err
		}
		return nil
	}()
	if copyErr != nil {
		return copyErr
	}

	if err = os.Chmod(dstPath, info.Mode().Perm()); err != nil {
		return err
	}
	return os.Remove(srcPath)
}
