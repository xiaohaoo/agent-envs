package config

import (
	"agent-envs/internal/fileutil"
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

const (
	// KeyActive 激活的 profile 键名
	KeyActive   = "active"
	KeyProfiles = "profiles"

	agentEnvsConfigDir  = "agent-envs"
	agentEnvsConfigFile = "config.toml"

	maskPrefix    = 8
	maskSuffix    = 4
	maskMinLength = 12
)

var (
	// ErrConfigNotFound 配置文件不存在
	ErrConfigNotFound = errors.New("配置文件不存在")

	// ErrInvalidFormat 配置格式无效
	ErrInvalidFormat = errors.New("配置格式无效")

	// ErrPermissionDenied 权限不足
	ErrPermissionDenied = errors.New("权限不足")

	// ErrActiveProfileNotFound 激活的 profile 不存在
	ErrActiveProfileNotFound = errors.New("激活的 profile 不存在")
)

// Config 全部配置
type Config struct {
	Active     string             `toml:"active"`
	ProfileMap map[string]Profile `toml:"-"`
}

// Profile 用 map 存储任意环境变量键值对，并保留原始 TOML 类型
type Profile map[string]any

// PathManager 管理所有配置文件路径
type PathManager struct {
	agentEnvsConfig string
	homeDir         string
}

// NewPathManager 创建路径管理器
func NewPathManager() (*PathManager, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户配置目录失败: %w", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("获取用户主目录失败: %w", err)
	}

	return &PathManager{
		agentEnvsConfig: filepath.Join(configDir, agentEnvsConfigDir, agentEnvsConfigFile),
		homeDir:         homeDir,
	}, nil
}

// AgentEnvsConfig 返回 agent-envs 统一配置文件路径
func (pm *PathManager) AgentEnvsConfig() string {
	return pm.agentEnvsConfig
}

// HomePath returns a path under the user's home directory.
func (pm *PathManager) HomePath(pathElementList ...string) string {
	pathList := append([]string{pm.homeDir}, pathElementList...)
	return filepath.Join(pathList...)
}

// Load 从统一配置文件加载指定代理配置
func Load(path, agentKey string) (*Config, error) {
	configMap, err := loadAll(path)
	if err != nil {
		return nil, err
	}

	cfg, ok := configMap[agentKey]
	if !ok {
		return nil, fmt.Errorf("%w: %s.%s", ErrConfigNotFound, path, agentKey)
	}
	return cfg, nil
}

// Save 保存指定代理配置到统一配置文件
func (c *Config) Save(path, agentKey string) error {
	configMap, err := loadAll(path)
	if err != nil && !errors.Is(err, ErrConfigNotFound) {
		return err
	}
	if configMap == nil {
		configMap = make(map[string]*Config)
	}

	configMap[agentKey] = c
	return saveAll(path, configMap)
}

func loadAll(path string) (map[string]*Config, error) {
	rawMap := make(map[string]any)
	if _, err := toml.DecodeFile(path, &rawMap); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrConfigNotFound, path)
		}
		if os.IsPermission(err) {
			return nil, fmt.Errorf("%w: %s", ErrPermissionDenied, path)
		}
		return nil, fmt.Errorf("%w: %s", ErrInvalidFormat, err)
	}

	configMap := make(map[string]*Config)
	for agentKey, value := range rawMap {
		sectionMap, ok := value.(map[string]any)
		if !ok {
			continue
		}
		cfg, err := parseConfig(sectionMap)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", agentKey, err)
		}
		configMap[agentKey] = cfg
	}

	return configMap, nil
}

func saveAll(path string, configMap map[string]*Config) error {
	var buf bytes.Buffer

	for _, agentKey := range sortedAgentKeys(configMap) {
		cfg := configMap[agentKey]
		if cfg == nil {
			continue
		}
		if buf.Len() > 0 {
			buf.WriteString("\n")
		}
		fmt.Fprintf(&buf, "[%s]\n", agentKey)
		fmt.Fprintf(&buf, "%s = %q\n", KeyActive, cfg.Active)

		for _, name := range cfg.SortedNames() {
			profileMap := cfg.ProfileMap[name]
			buf.WriteString("\n")
			fmt.Fprintf(&buf, "[%s.%s.%q]\n", agentKey, KeyProfiles, name)
			for _, key := range profileMap.SortedKeys() {
				line, err := encodeProfileEntry(key, profileMap[key])
				if err != nil {
					return err
				}
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), fileutil.ConfigDirPermission); err != nil {
		return fmt.Errorf("创建配置目录失败: %w", err)
	}
	return fileutil.AtomicWrite(path, bytes.TrimRight(buf.Bytes(), "\n"), fileutil.ConfigFilePermission)
}

func parseConfig(sectionMap map[string]any) (*Config, error) {
	cfg := &Config{ProfileMap: make(map[string]Profile)}
	if active, ok := sectionMap[KeyActive].(string); ok {
		cfg.Active = active
	}

	profileMap, _ := sectionMap[KeyProfiles].(map[string]any)
	for name, value := range profileMap {
		if parsedProfileMap := parseProfile(value); parsedProfileMap != nil {
			cfg.ProfileMap[name] = parsedProfileMap
		}
	}

	if cfg.Active != "" {
		if _, exists := cfg.ProfileMap[cfg.Active]; !exists {
			return nil, fmt.Errorf("%w: %s", ErrActiveProfileNotFound, cfg.Active)
		}
	}

	return cfg, nil
}

// parseProfile 解析 profile section
func parseProfile(value any) Profile {
	sectionMap, ok := value.(map[string]any)
	if !ok {
		return nil
	}

	profileMap := make(Profile)
	for key, value := range sectionMap {
		profileMap[key] = value
	}
	return profileMap
}

func sortedAgentKeys(configMap map[string]*Config) []string {
	keyList := make([]string, 0, len(configMap))
	for agentKey := range configMap {
		keyList = append(keyList, agentKey)
	}
	sort.Strings(keyList)
	return keyList
}

// SortedNames 返回排序后的 profile 名称列表
func (c *Config) SortedNames() []string {
	nameList := make([]string, 0, len(c.ProfileMap))
	for name := range c.ProfileMap {
		nameList = append(nameList, name)
	}
	sort.Strings(nameList)
	return nameList
}

func encodeProfileEntry(key string, value any) (string, error) {
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(map[string]any{key: value}); err != nil {
		return "", fmt.Errorf("编码 profile 字段 %s 失败: %w", key, err)
	}
	return string(bytes.TrimRight(buf.Bytes(), "\n")), nil
}

// String 返回指定键的字符串值
func (p Profile) String(key string) (string, bool) {
	value, ok := p[key]
	if !ok {
		return "", false
	}
	text, ok := value.(string)
	return text, ok
}

// Bool 返回指定键的布尔值
func (p Profile) Bool(key string) (bool, bool) {
	value, ok := p[key]
	if !ok {
		return false, false
	}

	switch typedValue := value.(type) {
	case bool:
		return typedValue, true
	case string:
		parsedValue, err := strconv.ParseBool(typedValue)
		if err == nil {
			return parsedValue, true
		}
	}

	return false, false
}

// GetToken 查找 profile 中的 token 值
// 查找包含 "TOKEN" 或 "KEY" 的键（不区分大小写）
func (p Profile) GetToken() string {
	for key, value := range p {
		upperKey := strings.ToUpper(key)
		if strings.Contains(upperKey, "TOKEN") || strings.Contains(upperKey, "KEY") {
			if token, ok := value.(string); ok {
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
	keyList := make([]string, 0, len(p))
	for key := range p {
		keyList = append(keyList, key)
	}
	sort.Strings(keyList)
	return keyList
}
