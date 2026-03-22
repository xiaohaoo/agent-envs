package agent

import (
	"agent-envs/internal/config"
	"agent-envs/internal/fileutil"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/BurntSushi/toml"
)

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
// 并行写入 ~/.codex/config.toml 和 ~/.codex/auth.json
func (c *Codex) ApplyProfile(profile config.Profile) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// 并行写入两个文件
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := c.writeConfigToml(profile); err != nil {
			errChan <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := c.writeAuthJson(profile); err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// writeConfigToml 写入 ~/.codex/config.toml
// 保留已有的配置字段，只覆盖 profile 中定义的 key
func (c *Codex) writeConfigToml(profile config.Profile) error {
	path := c.pm.CodexSettings()

	// 读取已有配置
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		_ = toml.Unmarshal(data, &existing)
	}

	// 顶层字段由用户自行控制，只合并 model_providers 子表
	providers, _ := existing["model_providers"].(map[string]interface{})
	if providers == nil {
		providers = make(map[string]interface{})
	}
	providers[profile[config.KeyName]] = map[string]interface{}{
		"name":     profile[config.KeyName],
		"base_url": profile[config.KeyBaseURL],
		"wire_api": profile[config.KeyWireAPI],
	}
	if auth, ok := profile[config.KeyRequiresOpenAIAuth]; ok {
		providerMap := providers[profile[config.KeyName]].(map[string]interface{})
		providerMap["requires_openai_auth"] = auth == "true"
	}
	existing["model_providers"] = providers

	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(existing); err != nil {
		return fmt.Errorf("编码 TOML 失败: %w", err)
	}

	return fileutil.AtomicWrite(path, buf.Bytes(), fileutil.ConfigFilePermission)
}

// writeAuthJson 写入 ~/.codex/auth.json
// 保留已有的认证字段，只覆盖 profile 中定义的 key
func (c *Codex) writeAuthJson(profile config.Profile) error {
	path := c.pm.CodexAuth()

	// 读取已有认证配置
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	// 只覆盖 OPENAI_API_KEY
	existing[config.KeyOpenAIAPIKey] = profile[config.KeyOpenAIAPIKey]

	authData, err := fileutil.MarshalJSONWithNewline(existing)
	if err != nil {
		return err
	}

	return fileutil.AtomicWrite(path, authData, fileutil.AuthFilePermission)
}
