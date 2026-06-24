package agent

import (
	"encoding/json"
	"fmt"
	"os"

	"agent-envs/internal/config"
	"agent-envs/internal/fileutil"
)

const (
	claudeKey          = "claude"
	claudeSettingsDir  = ".claude"
	claudeSettingsFile = "settings.json"
	claudeKeyEnv       = "env"
	claudeKeyBaseURL   = "ANTHROPIC_BASE_URL"
	claudeKeyAuthToken = "ANTHROPIC_AUTH_TOKEN"
)

// Claude manages Claude Code configuration.
type Claude struct {
	pm *config.PathManager
}

func NewClaude(pm *config.PathManager) *Claude {
	return &Claude{pm: pm}
}

func (c *Claude) Key() string {
	return claudeKey
}

func (c *Claude) Name() string {
	return "Claude Code"
}

func (c *Claude) Description() string {
	return "Anthropic Claude Code"
}

func (c *Claude) LoadConfig() (*config.Config, error) {
	return config.Load(c.pm.AgentEnvsConfig(), c.Key())
}

func (c *Claude) SaveConfig(cfg *config.Config) error {
	return cfg.Save(c.pm.AgentEnvsConfig(), c.Key())
}

// ApplyProfile merges profile values into ~/.claude/settings.json env.
func (c *Claude) ApplyProfile(name string, profileMap config.Profile) error {
	path := c.pm.HomePath(claudeSettingsDir, claudeSettingsFile)

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s failed: %w", path, err)
	}

	settingsMap := make(map[string]any)
	if err := json.Unmarshal(data, &settingsMap); err != nil {
		return fmt.Errorf("parse %s failed: %w", path, err)
	}

	envMap, _ := settingsMap[claudeKeyEnv].(map[string]any)
	if envMap == nil {
		envMap = make(map[string]any)
	}
	for key, value := range profileMap {
		envMap[key] = value
	}
	settingsMap[claudeKeyEnv] = envMap

	out, err := fileutil.MarshalJSONNoTrailingNewline(settingsMap)
	if err != nil {
		return err
	}

	return fileutil.AtomicWrite(path, out, fileutil.ConfigFilePermission)
}

func (c *Claude) BuildProfile(input ProfileInput) config.Profile {
	return config.Profile{
		claudeKeyBaseURL:   input.APIURL,
		claudeKeyAuthToken: input.Token,
	}
}

func (c *Claude) SummarizeProfile(profileMap config.Profile) ProfileSummary {
	url, _ := profileMap.String(claudeKeyBaseURL)
	return ProfileSummary{URL: url, Token: profileMap.MaskToken()}
}
