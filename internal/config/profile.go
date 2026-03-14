package config

import (
	"sort"
	"strings"
)

const (
	maskPrefix    = 8
	maskSuffix    = 4
	maskMinLength = 12
)

// Profile 用 map 存储任意环境变量键值对
type Profile map[string]string

// GetToken 查找 profile 中的 token 值
// 查找包含 "TOKEN" 或 "KEY" 的键（不区分大小写）
func (p Profile) GetToken() string {
	for key, val := range p {
		upperKey := strings.ToUpper(key)
		if strings.Contains(upperKey, "TOKEN") || strings.Contains(upperKey, "KEY") {
			return val
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
