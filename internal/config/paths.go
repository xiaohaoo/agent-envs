package config

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	claudeConfigFile   = ".claude/agent-envs.toml"
	claudeSettingsFile = ".claude/settings.json"
	codexConfigFile    = ".codex/agent-envs.toml"
	codexSettingsFile  = ".codex/config.toml"
	codexAuthFile      = ".codex/auth.json"
)

// PathManager 管理所有配置文件路径
type PathManager struct {
	homeDir string
}

// NewPathManager 创建路径管理器
func NewPathManager() (*PathManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}
	return &PathManager{homeDir: home}, nil
}

// ClaudeConfig 返回 Claude 配置文件路径
func (pm *PathManager) ClaudeConfig() string {
	return filepath.Join(pm.homeDir, claudeConfigFile)
}

// ClaudeSettings 返回 Claude settings.json 路径
func (pm *PathManager) ClaudeSettings() string {
	return filepath.Join(pm.homeDir, claudeSettingsFile)
}

// CodexConfig 返回 Codex 配置文件路径
func (pm *PathManager) CodexConfig() string {
	return filepath.Join(pm.homeDir, codexConfigFile)
}

// CodexSettings 返回 Codex config.toml 路径
func (pm *PathManager) CodexSettings() string {
	return filepath.Join(pm.homeDir, codexSettingsFile)
}

// CodexAuth 返回 Codex auth.json 路径
func (pm *PathManager) CodexAuth() string {
	return filepath.Join(pm.homeDir, codexAuthFile)
}
