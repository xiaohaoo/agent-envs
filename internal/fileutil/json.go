package fileutil

import (
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
