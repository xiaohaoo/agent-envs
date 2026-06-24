package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAtomicWriteCreatesParentDirectoryAndSetsPermission(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "config.toml")

	if err := AtomicWrite(path, []byte("hello"), AuthFilePermission); err != nil {
		t.Fatalf("AtomicWrite() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if string(data) != "hello" {
		t.Fatalf("file content = %q", data)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if got := info.Mode().Perm(); got != AuthFilePermission {
		t.Fatalf("permission = %o, want %o", got, AuthFilePermission)
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Fatalf("tmp file still exists or unexpected error: %v", err)
	}
}
