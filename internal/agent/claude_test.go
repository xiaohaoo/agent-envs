package agent

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"agent-envs/internal/config"
)

func TestClaudeApplyProfileCreatesSettingsAndDirectory(t *testing.T) {
	pm := newTestPathManager(t)
	claude := NewClaude(pm)

	profileMap := config.Profile{
		claudeKeyBaseURL:   "https://api.example.com",
		claudeKeyAuthToken: "sk-ant-test",
	}
	if err := claude.ApplyProfile("demo", profileMap); err != nil {
		t.Fatalf("ApplyProfile() error = %v", err)
	}

	path := filepath.Join(testHomeDir(t), claudeSettingsDir, claudeSettingsFile)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var settingsMap map[string]any
	if err := json.Unmarshal(data, &settingsMap); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	envMap, ok := settingsMap[claudeKeyEnv].(map[string]any)
	if !ok {
		t.Fatalf("env map missing in %#v", settingsMap)
	}
	if got := envMap[claudeKeyBaseURL]; got != "https://api.example.com" {
		t.Fatalf("base url = %#v", got)
	}
	if got := envMap[claudeKeyAuthToken]; got != "sk-ant-test" {
		t.Fatalf("token = %#v", got)
	}
}

func TestClaudeApplyProfilePreservesExistingSettings(t *testing.T) {
	pm := newTestPathManager(t)
	path := filepath.Join(testHomeDir(t), claudeSettingsDir, claudeSettingsFile)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}
	if err := os.WriteFile(path, []byte(`{"theme":"dark","env":{"KEEP":"yes"}}`), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	claude := NewClaude(pm)
	if err := claude.ApplyProfile("demo", config.Profile{claudeKeyAuthToken: "sk-ant-test"}); err != nil {
		t.Fatalf("ApplyProfile() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	var settingsMap map[string]any
	if err := json.Unmarshal(data, &settingsMap); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if got := settingsMap["theme"]; got != "dark" {
		t.Fatalf("theme = %#v", got)
	}
	envMap := settingsMap[claudeKeyEnv].(map[string]any)
	if got := envMap["KEEP"]; got != "yes" {
		t.Fatalf("KEEP = %#v", got)
	}
	if got := envMap[claudeKeyAuthToken]; got != "sk-ant-test" {
		t.Fatalf("token = %#v", got)
	}
}
