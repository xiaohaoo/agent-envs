package agent

import (
	"agent-envs/internal/config"
	"agent-envs/internal/fileutil"
	"bytes"
	"fmt"
	"sync"
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
func (c *Codex) writeConfigToml(profile config.Profile) error {
	path := c.pm.CodexSettings()

	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("model_provider = %q\n", profile[config.KeyModelProvider]))
	buf.WriteString(fmt.Sprintf("model = %q\n", profile[config.KeyModel]))
	if effort, ok := profile[config.KeyModelReasoningEffort]; ok {
		buf.WriteString(fmt.Sprintf("model_reasoning_effort = %q\n", effort))
	}
	buf.WriteString("\n")
	buf.WriteString(fmt.Sprintf("[model_providers.%s]\n", profile[config.KeyName]))
	buf.WriteString(fmt.Sprintf("name = %q\n", profile[config.KeyName]))
	buf.WriteString(fmt.Sprintf("base_url = %q\n", profile[config.KeyBaseURL]))
	buf.WriteString(fmt.Sprintf("wire_api = %q\n", profile[config.KeyWireAPI]))
	if auth, ok := profile[config.KeyRequiresOpenAIAuth]; ok {
		buf.WriteString(fmt.Sprintf("requires_openai_auth = %s\n", auth))
	}

	return fileutil.AtomicWrite(path, buf.Bytes(), fileutil.ConfigFilePermission)
}

// writeAuthJson 写入 ~/.codex/auth.json
func (c *Codex) writeAuthJson(profile config.Profile) error {
	path := c.pm.CodexAuth()

	auth := map[string]string{
		config.KeyOpenAIAPIKey: profile[config.KeyOpenAIAPIKey],
	}

	authData, err := fileutil.MarshalJSONWithNewline(auth)
	if err != nil {
		return err
	}

	return fileutil.AtomicWrite(path, authData, fileutil.AuthFilePermission)
}
