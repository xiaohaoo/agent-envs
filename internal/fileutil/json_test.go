package fileutil

import (
	"bytes"
	"testing"
)

func TestMarshalJSONNoTrailingNewline(t *testing.T) {
	data, err := MarshalJSONNoTrailingNewline(map[string]string{"name": "demo"})
	if err != nil {
		t.Fatalf("MarshalJSONNoTrailingNewline() error = %v", err)
	}
	if bytes.HasSuffix(data, []byte("\n")) {
		t.Fatalf("JSON unexpectedly ended with newline: %q", data)
	}
	if !bytes.Contains(data, []byte(`"name": "demo"`)) {
		t.Fatalf("JSON missing field: %q", data)
	}
}

func TestEnsureSingleTrailingNewline(t *testing.T) {
	got := EnsureSingleTrailingNewline([]byte("hello\n\n"))
	if string(got) != "hello\n" {
		t.Fatalf("EnsureSingleTrailingNewline() = %q", got)
	}
}
