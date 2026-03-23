package fileutil

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MarshalJSONWithNewline 序列化 JSON 并添加换行符
func MarshalJSONWithNewline(v interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}
	return append(data, '\n'), nil
}

// EnsureSingleTrailingNewline 确保字节切片以且仅以一个换行符结尾
func EnsureSingleTrailingNewline(b []byte) []byte {
	return append(bytes.TrimRight(b, "\n"), '\n')
}
