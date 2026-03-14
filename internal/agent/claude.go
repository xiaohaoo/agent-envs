package agent

import (
	"agent-envs/internal/config"
	"agent-envs/internal/fileutil"
	"encoding/json"
	"fmt"
	"os"
)

// Claude 实现 Claude Code 代理
type Claude struct {
	pm *config.PathManager
}

// NewClaude 创建 Claude 代理实例
func NewClaude(pm *config.PathManager) *Claude {
	return &Claude{pm: pm}
}

// Name 返回代理名称
func (c *Claude) Name() string {
	return "Claude Code"
}

// LoadConfig 加载 Claude 配置
func (c *Claude) LoadConfig() (*config.Config, error) {
	return config.Load(c.pm.ClaudeConfig())
}

// SaveConfig 保存 Claude 配置
func (c *Claude) SaveConfig(cfg *config.Config) error {
	return cfg.Save(c.pm.ClaudeConfig())
}

// ApplyProfile 将 profile 应用到 ~/.claude/settings.json 的 env 字段
func (c *Claude) ApplyProfile(profile config.Profile) error {
	path := c.pm.ClaudeSettings()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", path, err)
	}

	var settings map[string]interface{}
	if err := json.Unmarshal(data, &settings); err != nil {
		return fmt.Errorf("解析 %s 失败: %w", path, err)
	}

	// 替换 env 字段
	env := make(map[string]interface{})
	for k, v := range profile {
		env[k] = v
	}
	settings[config.KeyEnv] = env

	// 写回，保持缩进格式
	out, err := fileutil.MarshalJSONWithNewline(settings)
	if err != nil {
		return err
	}

	return fileutil.AtomicWrite(path, out, fileutil.ConfigFilePermission)
}
