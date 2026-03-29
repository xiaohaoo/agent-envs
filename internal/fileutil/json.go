package fileutil

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// MarshalJSONNoTrailingNewline 序列化 JSON，不追加末尾换行
func MarshalJSONNoTrailingNewline(v interface{}) ([]byte, error) {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("序列化失败: %w", err)
	}
	return data, nil
}

// EnsureSingleTrailingNewline 确保字节切片以且仅以一个换行符结尾
func EnsureSingleTrailingNewline(b []byte) []byte {
	return append(bytes.TrimRight(b, "\n"), '\n')
}
