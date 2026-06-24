package agent

import (
	"os"
	"testing"

	"agent-envs/internal/config"
)

func newTestPathManager(t *testing.T) *config.PathManager {
	t.Helper()

	homeDir := t.TempDir()
	configDir := t.TempDir()
	t.Setenv("HOME", homeDir)
	t.Setenv("USERPROFILE", homeDir)
	t.Setenv("APPDATA", configDir)
	t.Setenv("XDG_CONFIG_HOME", configDir)

	pm, err := config.NewPathManager()
	if err != nil {
		t.Fatalf("NewPathManager() error = %v", err)
	}
	return pm
}

func testHomeDir(t *testing.T) string {
	t.Helper()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("UserHomeDir() error = %v", err)
	}
	return homeDir
}
