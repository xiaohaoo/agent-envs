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
	// 缓存的路径
	claudeConfig   string
	claudeSettings string
	codexConfig    string
	codexSettings  string
	codexAuth      string
}

// NewPathManager 创建路径管理器
func NewPathManager() (*PathManager, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}
	pm := &PathManager{homeDir: home}
	// 预先计算所有路径
	pm.claudeConfig = filepath.Join(home, claudeConfigFile)
	pm.claudeSettings = filepath.Join(home, claudeSettingsFile)
	pm.codexConfig = filepath.Join(home, codexConfigFile)
	pm.codexSettings = filepath.Join(home, codexSettingsFile)
	pm.codexAuth = filepath.Join(home, codexAuthFile)
	return pm, nil
}

// ClaudeConfig 返回 Claude 配置文件路径
func (pm *PathManager) ClaudeConfig() string {
	return pm.claudeConfig
}

// ClaudeSettings 返回 Claude settings.json 路径
func (pm *PathManager) ClaudeSettings() string {
	return pm.claudeSettings
}

// CodexConfig 返回 Codex 配置文件路径
func (pm *PathManager) CodexConfig() string {
	return pm.codexConfig
}

// CodexSettings 返回 Codex config.toml 路径
func (pm *PathManager) CodexSettings() string {
	return pm.codexSettings
}

// CodexAuth 返回 Codex auth.json 路径
func (pm *PathManager) CodexAuth() string {
	return pm.codexAuth
}
