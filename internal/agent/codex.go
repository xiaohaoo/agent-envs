package agent

import (
	"agent-envs/internal/config"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

const authFilePermission = 0600

// Codex 实现 Codex CLI 代理
type Codex struct {
	pm *config.PathManager
}

// NewCodex 创建 Codex 代理实例
func NewCodex(pm *config.PathManager) *Codex {
	return &Codex{pm: pm}
}

// Name 返回代理名称
func (c *Codex) Name() string {
	return "Codex"
}

// LoadConfig 加载 Codex 配置
func (c *Codex) LoadConfig() (*config.Config, error) {
	return config.Load(c.pm.CodexConfig())
}

// SaveConfig 保存 Codex 配置
func (c *Codex) SaveConfig(cfg *config.Config) error {
	return cfg.Save(c.pm.CodexConfig())
}

// ApplyProfile 将 profile 应用到 Codex 配置文件
// 写入 ~/.codex/config.toml 和 ~/.codex/auth.json
func (c *Codex) ApplyProfile(profile config.Profile) error {
	// 写入 config.toml
	if err := c.writeConfigToml(profile); err != nil {
		return err
	}

	// 写入 auth.json
	if err := c.writeAuthJson(profile); err != nil {
		return err
	}

	return nil
}

// writeConfigToml 写入 ~/.codex/config.toml
func (c *Codex) writeConfigToml(profile config.Profile) error {
	path := c.pm.CodexSettings()

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("model_provider = %q\n", profile["model_provider"]))
	buf.WriteString(fmt.Sprintf("model = %q\n", profile["model"]))
	if effort, ok := profile["model_reasoning_effort"]; ok {
		buf.WriteString(fmt.Sprintf("model_reasoning_effort = %q\n", effort))
	}
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("[model_providers.%s]\n", profile["name"]))
	buf.WriteString(fmt.Sprintf("name = %q\n", profile["name"]))
	buf.WriteString(fmt.Sprintf("base_url = %q\n", profile["base_url"]))
	buf.WriteString(fmt.Sprintf("wire_api = %q\n", profile["wire_api"]))
	if auth, ok := profile["requires_openai_auth"]; ok {
		buf.WriteString(fmt.Sprintf("requires_openai_auth = %s\n", auth))
	}

	// 原子写入
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), filePermission); err != nil {
		return fmt.Errorf("写入 config.toml 临时文件失败: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("重命名 config.toml 失败: %w", err)
	}

	return nil
}

// writeAuthJson 写入 ~/.codex/auth.json
func (c *Codex) writeAuthJson(profile config.Profile) error {
	path := c.pm.CodexAuth()

	auth := map[string]string{
		"OPENAI_API_KEY": profile["OPENAI_API_KEY"],
	}

	authData, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 auth.json 失败: %w", err)
	}
	authData = append(authData, '\n')

	// 原子写入
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, authData, authFilePermission); err != nil {
		return fmt.Errorf("写入 auth.json 临时文件失败: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("重命名 auth.json 失败: %w", err)
	}

	return nil
}
