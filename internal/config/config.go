package config

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/BurntSushi/toml"
)

const filePermission = 0644

// Config 全部配置
type Config struct {
	Active   string             `toml:"active"`
	Profiles map[string]Profile `toml:"-"`
}

// Load 从指定路径加载配置
func Load(path string) (*Config, error) {
	raw := make(map[string]interface{})
	if _, err := toml.DecodeFile(path, &raw); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, path)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: %s", ErrPermissionDenied, path)
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidFormat, err)
	}

	cfg := &Config{Profiles: make(map[string]Profile)}
	if active, ok := raw["active"].(string); ok {
		cfg.Active = active
	}

	// 解析每个 section 为 profile
	for key, val := range raw {
		if key == "active" {
			continue
		}
		if profile := parseProfile(val); profile != nil {
			cfg.Profiles[key] = profile
		}
	}

	// 验证激活的 profile 存在
	if cfg.Active != "" {
		if _, exists := cfg.Profiles[cfg.Active]; !exists {
			return nil, fmt.Errorf("%w: %s", ErrActiveProfileNotFound, cfg.Active)
		}
	}

	return cfg, nil
}

// parseProfile 解析 profile section
func parseProfile(val interface{}) Profile {
	section, ok := val.(map[string]interface{})
	if !ok {
		return nil
	}

	profile := make(Profile)
	for k, v := range section {
		if s, ok := v.(string); ok {
			profile[k] = s
		}
	}
	return profile
}

// Save 保存配置到指定路径
func (c *Config) Save(path string) error {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "active = %q\n\n", c.Active)

	// 按名称排序输出
	for _, name := range c.SortedNames() {
		profile := c.Profiles[name]
		fmt.Fprintf(&buf, "[%q]\n", name)
		for _, key := range profile.SortedKeys() {
			fmt.Fprintf(&buf, "%s = %q\n", key, profile[key])
		}
		buf.WriteString("\n")
	}

	// 原子写入：先写临时文件，再重命名
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, buf.Bytes(), filePermission); err != nil {
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath) // 清理临时文件
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	return nil
}

// SortedNames 返回排序后的 profile 名称列表
func (c *Config) SortedNames() []string {
	names := make([]string, 0, len(c.Profiles))
	for name := range c.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
