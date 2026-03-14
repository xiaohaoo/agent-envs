package fileutil

import (
	"fmt"
	"os"
)

const (
	// ConfigFilePermission 配置文件权限
	ConfigFilePermission = 0644
	// AuthFilePermission 认证文件权限
	AuthFilePermission = 0600
)

// AtomicWrite 原子写入文件，先写临时文件再重命名
func AtomicWrite(path string, data []byte, perm os.FileMode) error {
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, perm); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}
