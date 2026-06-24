package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadReadsAgentSectionFromUnifiedConfig(t *testing.T) {
	path := writeTempConfig(t, `
[alpha]
active = "primary"

[alpha.profiles."primary"]
TOKEN = "sk-alpha"
BASE_URL = "https://alpha.example.com"

[beta]
active = "backup"

[beta.profiles."backup"]
base_url = "https://beta.example.com"
requires_auth = true
API_KEY = "sk-beta"
`)

	cfg, err := Load(path, "beta")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Active != "backup" {
		t.Fatalf("Active = %q, want %q", cfg.Active, "backup")
	}

	profileMap := cfg.ProfileMap["backup"]
	if profileMap == nil {
		t.Fatal("missing backup profile")
	}
	if baseURL, _ := profileMap.String("base_url"); baseURL != "https://beta.example.com" {
		t.Fatalf("base_url = %q", baseURL)
	}
	if auth, _ := profileMap.Bool("requires_auth"); !auth {
		t.Fatal("requires_auth = false, want true")
	}
}

func TestSaveUpdatesOneAgentAndPreservesOtherSection(t *testing.T) {
	path := writeTempConfig(t, `
[alpha]
active = "primary"

[alpha.profiles."primary"]
TOKEN = "sk-alpha"
BASE_URL = "https://alpha.example.com"

[beta]
active = "backup"

[beta.profiles."backup"]
base_url = "https://beta.example.com"
API_KEY = "sk-beta"
`)

	cfg, err := Load(path, "alpha")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	cfg.Active = "secondary"
	cfg.ProfileMap["secondary"] = Profile{
		"TOKEN":    "sk-alpha-secondary",
		"BASE_URL": "https://secondary.example.com",
	}

	if err := cfg.Save(path, "alpha"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	alphaCfg, err := Load(path, "alpha")
	if err != nil {
		t.Fatalf("Load(alpha) error = %v", err)
	}
	if alphaCfg.Active != "secondary" {
		t.Fatalf("alpha active = %q, want %q", alphaCfg.Active, "secondary")
	}

	betaCfg, err := Load(path, "beta")
	if err != nil {
		t.Fatalf("Load(beta) error = %v", err)
	}
	if betaCfg.Active != "backup" {
		t.Fatalf("beta active = %q, want %q", betaCfg.Active, "backup")
	}
	if _, ok := betaCfg.ProfileMap["backup"]; !ok {
		t.Fatal("beta backup profile was not preserved")
	}
}

func TestPathManagerUsesUserConfigDirAndSaveCreatesDirectory(t *testing.T) {
	configDir := t.TempDir()
	homeDir := t.TempDir()
	t.Setenv("APPDATA", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)

	pm, err := NewPathManager()
	if err != nil {
		t.Fatalf("NewPathManager() error = %v", err)
	}

	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("UserConfigDir() error = %v", err)
	}
	wantPath := filepath.Join(userConfigDir, agentEnvsConfigDir, agentEnvsConfigFile)
	if pm.AgentEnvsConfig() != wantPath {
		t.Fatalf("AgentEnvsConfig() = %q, want %q", pm.AgentEnvsConfig(), wantPath)
	}

	cfg := &Config{ProfileMap: map[string]Profile{
		"demo": {
			"base_url": "https://api.example.com",
			"api_key":  "sk-demo",
		},
	}}
	if err := cfg.Save(pm.AgentEnvsConfig(), "demo-agent"); err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if _, err := os.Stat(wantPath); err != nil {
		t.Fatalf("saved config stat error = %v", err)
	}

	wantHomePath := filepath.Join(homeDir, "nested", "settings.json")
	if got := pm.HomePath("nested", "settings.json"); got != wantHomePath {
		t.Fatalf("HomePath() = %q, want %q", got, wantHomePath)
	}
}

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "agent-envs.toml")
	if err := os.WriteFile(path, []byte(strings.TrimSpace(content)), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	return path
}
