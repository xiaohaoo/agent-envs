package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"agent-envs/internal/config"
)

func TestCodexApplyProfileCreatesNativeFilesAndDirectory(t *testing.T) {
	pm := newTestPathManager(t)
	codex := NewCodex(pm)

	profileMap := config.Profile{
		codexKeyBaseURL:            "https://api.example.com",
		codexKeyWireAPI:            "responses",
		codexKeyRequiresOpenAIAuth: true,
		codexKeyOpenAIAPIKey:       "sk-test",
	}
	if err := codex.ApplyProfile("demo", profileMap); err != nil {
		t.Fatalf("ApplyProfile() error = %v", err)
	}

	configPath := filepath.Join(testHomeDir(t), codexConfigDir, codexSettingsFile)
	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	configText := string(configData)
	for _, want := range []string{
		`model_provider = "demo"`,
		`[model_providers."demo"]`,
		`base_url = "https://api.example.com"`,
		`name = "demo"`,
		`requires_openai_auth = true`,
		`wire_api = "responses"`,
	} {
		if !strings.Contains(configText, want) {
			t.Fatalf("config missing %q:\n%s", want, configText)
		}
	}

	authPath := filepath.Join(testHomeDir(t), codexConfigDir, codexAuthFile)
	authData, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("ReadFile(auth) error = %v", err)
	}
	var authMap map[string]any
	if err := json.Unmarshal(authData, &authMap); err != nil {
		t.Fatalf("Unmarshal(auth) error = %v", err)
	}
	if got := authMap[codexKeyOpenAIAPIKey]; got != "sk-test" {
		t.Fatalf("OPENAI_API_KEY = %#v", got)
	}
	authInfo, err := os.Stat(authPath)
	if err != nil {
		t.Fatalf("Stat(auth) error = %v", err)
	}
	if got := authInfo.Mode().Perm(); got != 0o600 {
		t.Fatalf("auth permission = %o, want 600", got)
	}
}

func TestCodexApplyProfilePreservesExistingConfigAndAuth(t *testing.T) {
	pm := newTestPathManager(t)
	codexDir := filepath.Join(testHomeDir(t), codexConfigDir)
	if err := os.MkdirAll(codexDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	configPath := filepath.Join(codexDir, codexSettingsFile)
	originalConfig := `
approval_policy = "on-request"
model_provider = "old"

[model_providers."old"]
base_url = "https://old.example.com"
name = "old"
wire_api = "responses"

[model_providers."other"]
base_url = "https://other.example.com"
name = "other"
wire_api = "chat"
`
	if err := os.WriteFile(configPath, []byte(strings.TrimSpace(originalConfig)), 0o644); err != nil {
		t.Fatalf("WriteFile(config) error = %v", err)
	}

	authPath := filepath.Join(codexDir, codexAuthFile)
	if err := os.WriteFile(authPath, []byte(`{"OTHER":"keep","OPENAI_API_KEY":"old-key"}`), 0o600); err != nil {
		t.Fatalf("WriteFile(auth) error = %v", err)
	}

	if err := NewCodex(pm).ApplyProfile("new", config.Profile{
		codexKeyBaseURL:            "https://new.example.com",
		codexKeyWireAPI:            "responses",
		codexKeyRequiresOpenAIAuth: false,
	}); err != nil {
		t.Fatalf("ApplyProfile() error = %v", err)
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("ReadFile(config) error = %v", err)
	}
	configText := string(configData)
	for _, want := range []string{
		`approval_policy = "on-request"`,
		`model_provider = "new"`,
		`[model_providers."other"]`,
		`base_url = "https://other.example.com"`,
		`[model_providers."new"]`,
		`requires_openai_auth = false`,
	} {
		if !strings.Contains(configText, want) {
			t.Fatalf("config missing %q:\n%s", want, configText)
		}
	}

	authData, err := os.ReadFile(authPath)
	if err != nil {
		t.Fatalf("ReadFile(auth) error = %v", err)
	}
	var authMap map[string]any
	if err := json.Unmarshal(authData, &authMap); err != nil {
		t.Fatalf("Unmarshal(auth) error = %v", err)
	}
	if got := authMap["OTHER"]; got != "keep" {
		t.Fatalf("OTHER = %#v", got)
	}
	if got := authMap[codexKeyOpenAIAPIKey]; got != "old-key" {
		t.Fatalf("OPENAI_API_KEY = %#v", got)
	}
}
