package config

import (
	"sort"
	"strconv"
	"strings"
)

const (
	maskPrefix    = 8
	maskSuffix    = 4
	maskMinLength = 12
)

// Profile 用 map 存储任意环境变量键值对，并保留原始 TOML 类型
type Profile map[string]any

// String 返回指定键的字符串值
func (p Profile) String(key string) (string, bool) {
	val, ok := p[key]
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok
}

// Bool 返回指定键的布尔值
func (p Profile) Bool(key string) (bool, bool) {
	val, ok := p[key]
	if !ok {
		return false, false
	}

	switch typed := val.(type) {
	case bool:
		return typed, true
	case string:
		parsed, err := strconv.ParseBool(typed)
		if err == nil {
			return parsed, true
		}
	}

	return false, false
}

// GetToken 查找 profile 中的 token 值
// 查找包含 "TOKEN" 或 "KEY" 的键（不区分大小写）
func (p Profile) GetToken() string {
	for key, val := range p {
		upperKey := strings.ToUpper(key)
		if strings.Contains(upperKey, "TOKEN") || strings.Contains(upperKey, "KEY") {
			if token, ok := val.(string); ok {
				return token
			}
		}
	}
	return ""
}

// MaskToken 遮蔽 token 值，只显示前8位和后4位
func (p Profile) MaskToken() string {
	token := p.GetToken()
	if len(token) <= maskMinLength {
		return "****"
	}
	return token[:maskPrefix] + "****" + token[len(token)-maskSuffix:]
}

// SortedKeys 返回排序后的键列表
func (p Profile) SortedKeys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
