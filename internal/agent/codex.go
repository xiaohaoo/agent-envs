package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"

	"agent-envs/internal/config"
	"agent-envs/internal/fileutil"
)

// Codex 实现 Codex CLI 代理
type Codex struct {
	pm *config.PathManager
}

// NewCodex 创建 Codex 代理实例
func NewCodex(pm *config.PathManager) *Codex {
	return &Codex{pm: pm}
}

// Name 返回代理名称
func (c *Codex) Name() string {
	return "Codex"
}

// LoadConfig 加载 Codex 配置
func (c *Codex) LoadConfig() (*config.Config, error) {
	return config.Load(c.pm.CodexConfig())
}

// SaveConfig 保存 Codex 配置
func (c *Codex) SaveConfig(cfg *config.Config) error {
	return cfg.Save(c.pm.CodexConfig())
}

// ApplyProfile 将 profile 应用到 Codex 配置文件
// 并行写入 ~/.codex/config.toml 和 ~/.codex/auth.json
func (c *Codex) ApplyProfile(name string, profile config.Profile) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// 并行写入两个文件
	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := c.writeConfigToml(name, profile); err != nil {
			errChan <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := c.writeAuthJson(profile); err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

// writeConfigToml 写入 ~/.codex/config.toml
// 保留已有的配置字段，只覆盖 model_provider 和 model_providers 子表
func (c *Codex) writeConfigToml(name string, profile config.Profile) error {
	path := c.pm.CodexSettings()

	// 读取已有配置
	var original string
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		original = string(data)
		if err := toml.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("解析已有配置失败: %w", err)
		}
	}

	// 更新 model_providers 子表（用 profile name 作为 key）
	providers, _ := existing["model_providers"].(map[string]interface{})
	if providers == nil {
		providers = make(map[string]interface{})
	}
	baseURL, _ := profile.String(config.KeyBaseURL)
	wireAPI, _ := profile.String(config.KeyWireAPI)
	providerEntry := map[string]interface{}{
		"base_url": baseURL,
		"name":     name,
		"wire_api": wireAPI,
	}
	if auth, ok := profile.Bool(config.KeyRequiresOpenAIAuth); ok {
		providerEntry["requires_openai_auth"] = auth
	}
	providers[name] = providerEntry

	// 设置顶层 model_provider
	existing[config.KeyModelProvider] = name

	body, providerOrder := stripManagedCodexConfig(original, name)
	rendered, err := renderCodexConfig(body, providers, providerOrder)
	if err != nil {
		return err
	}

	return fileutil.AtomicWrite(path, bytes.TrimRight([]byte(rendered), "\n"), fileutil.ConfigFilePermission)
}

// writeAuthJson 写入 ~/.codex/auth.json
// 保留已有的认证字段，只覆盖 profile 中定义的 key
func (c *Codex) writeAuthJson(profile config.Profile) error {
	path := c.pm.CodexAuth()

	// 读取已有认证配置
	existing := make(map[string]interface{})
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("解析已有认证配置失败: %w", err)
		}
	}

	// 只覆盖 OPENAI_API_KEY
	if apiKey, ok := profile[config.KeyOpenAIAPIKey]; ok {
		existing[config.KeyOpenAIAPIKey] = apiKey
	}

	authData, err := fileutil.MarshalJSONNoTrailingNewline(existing)
	if err != nil {
		return err
	}

	return fileutil.AtomicWrite(path, authData, fileutil.AuthFilePermission)
}

var codexTableHeaderPattern = regexp.MustCompile(`^\s*\[([^\[\]]+)\]\s*$`)

func stripManagedCodexConfig(original, selectedProvider string) (string, []string) {
	lines := splitLinesKeepNewline(original)
	if len(lines) == 0 {
		return fmt.Sprintf("%s = %q\n", config.KeyModelProvider, selectedProvider), nil
	}

	var (
		out                []string
		providerOrder      []string
		currentSection     string
		skippingProviders  bool
		replacedTopLevelKV bool
	)

	for _, line := range lines {
		if header, ok := parseCodexTableHeader(line); ok {
			currentSection = header

			if providerName, managed := parseModelProvidersHeader(header); managed {
				skippingProviders = true
				if providerName != "" {
					providerOrder = appendUnique(providerOrder, providerName)
				}
				continue
			}

			skippingProviders = false
			out = append(out, line)
			continue
		}

		if skippingProviders {
			continue
		}

		if currentSection == "" && isTopLevelKeyAssignment(line, config.KeyModelProvider) {
			out = append(out, fmt.Sprintf("%s = %q\n", config.KeyModelProvider, selectedProvider))
			replacedTopLevelKV = true
			continue
		}

		out = append(out, line)
	}

	if !replacedTopLevelKV {
		out = insertTopLevelKey(out, config.KeyModelProvider, selectedProvider)
	}

	return strings.TrimRight(strings.Join(out, ""), "\n"), providerOrder
}

func renderCodexConfig(body string, providers map[string]interface{}, providerOrder []string) (string, error) {
	var buf bytes.Buffer

	if body != "" {
		buf.WriteString(body)
	}

	for _, name := range orderedProviderNames(providers, providerOrder) {
		pMap, _ := providers[name].(map[string]interface{})
		if pMap == nil {
			continue
		}

		if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteString("\n")
		}

		buf.WriteString(fmt.Sprintf("[model_providers.%q]\n", name))
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(pMap); err != nil {
			return "", fmt.Errorf("编码 model_providers.%s 失败: %w", name, err)
		}
	}

	return buf.String(), nil
}

func orderedProviderNames(providers map[string]interface{}, existingOrder []string) []string {
	seen := make(map[string]struct{}, len(existingOrder))
	ordered := make([]string, 0, len(providers))

	for _, name := range existingOrder {
		if _, ok := providers[name]; ok {
			ordered = append(ordered, name)
			seen[name] = struct{}{}
		}
	}

	remaining := make([]string, 0, len(providers))
	for name := range providers {
		if _, ok := seen[name]; ok {
			continue
		}
		remaining = append(remaining, name)
	}
	sort.Strings(remaining)

	return append(ordered, remaining...)
}

func splitLinesKeepNewline(text string) []string {
	if text == "" {
		return nil
	}
	return strings.SplitAfter(text, "\n")
}

func parseCodexTableHeader(line string) (string, bool) {
	matches := codexTableHeaderPattern.FindStringSubmatch(line)
	if len(matches) != 2 {
		return "", false
	}
	return strings.TrimSpace(matches[1]), true
}

func parseModelProvidersHeader(header string) (string, bool) {
	if header == "model_providers" {
		return "", true
	}

	const prefix = "model_providers."
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	name := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if unquoted, err := strconv.Unquote(name); err == nil {
		name = unquoted
	}

	return name, true
}

func isTopLevelKeyAssignment(line, key string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return false
	}
	return strings.HasPrefix(trimmed, key+" =")
}

func insertTopLevelKey(lines []string, key, value string) []string {
	insertLine := fmt.Sprintf("%s = %q\n", key, value)
	if len(lines) == 0 {
		return []string{insertLine}
	}

	firstHeader := len(lines)
	for i, line := range lines {
		if _, ok := parseCodexTableHeader(line); ok {
			firstHeader = i
			break
		}
	}

	insertAt := firstHeader
	for insertAt > 0 && strings.TrimSpace(lines[insertAt-1]) == "" {
		insertAt--
	}

	lines = append(lines[:insertAt], append([]string{insertLine}, lines[insertAt:]...)...)
	return lines
}

func appendUnique(values []string, candidate string) []string {
	for _, value := range values {
		if value == candidate {
			return values
		}
	}
	return append(values, candidate)
}
