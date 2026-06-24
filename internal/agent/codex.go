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

const (
	codexKey                   = "codex"
	codexConfigDir             = ".codex"
	codexSettingsFile          = "config.toml"
	codexAuthFile              = "auth.json"
	codexKeyBaseURL            = "base_url"
	codexKeyWireAPI            = "wire_api"
	codexKeyRequiresOpenAIAuth = "requires_openai_auth"
	codexKeyOpenAIAPIKey       = "OPENAI_API_KEY"
	codexKeyModelProvider      = "model_provider"
)

// Codex manages Codex CLI configuration.
type Codex struct {
	pm *config.PathManager
}

func NewCodex(pm *config.PathManager) *Codex {
	return &Codex{pm: pm}
}

func (c *Codex) Key() string {
	return codexKey
}

func (c *Codex) Name() string {
	return "Codex"
}

func (c *Codex) Description() string {
	return "Codex CLI"
}

func (c *Codex) LoadConfig() (*config.Config, error) {
	return config.Load(c.pm.AgentEnvsConfig(), c.Key())
}

func (c *Codex) SaveConfig(cfg *config.Config) error {
	return cfg.Save(c.pm.AgentEnvsConfig(), c.Key())
}

// ApplyProfile writes the selected profile into Codex native files.
func (c *Codex) ApplyProfile(name string, profileMap config.Profile) error {
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	wg.Add(2)
	go func() {
		defer wg.Done()
		if err := c.writeConfigToml(name, profileMap); err != nil {
			errChan <- err
		}
	}()
	go func() {
		defer wg.Done()
		if err := c.writeAuthJson(profileMap); err != nil {
			errChan <- err
		}
	}()

	wg.Wait()
	close(errChan)

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Codex) ProfileFieldList() []ProfileField {
	return []ProfileField{
		{Key: codexKeyBaseURL, Label: "API", Secret: false},
		{Key: codexKeyOpenAIAPIKey, Label: "Token", Secret: true},
	}
}

func (c *Codex) BuildProfile(input ProfileInput) config.Profile {
	return config.Profile{
		codexKeyBaseURL:            input.FieldValueMap[codexKeyBaseURL],
		codexKeyWireAPI:            "responses",
		codexKeyRequiresOpenAIAuth: true,
		codexKeyOpenAIAPIKey:       input.FieldValueMap[codexKeyOpenAIAPIKey],
	}
}

func (c *Codex) ProfileSummaryItemList(profileMap config.Profile) []ProfileSummaryItem {
	url, _ := profileMap.String(codexKeyBaseURL)
	token, _ := profileMap.String(codexKeyOpenAIAPIKey)
	return []ProfileSummaryItem{
		{Label: "API", Value: url},
		{Label: "Token", Value: token, Secret: true},
	}
}

func (c *Codex) writeConfigToml(name string, profileMap config.Profile) error {
	path := c.pm.HomePath(codexConfigDir, codexSettingsFile)

	var original string
	existingMap := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
		original = string(data)
		if err := toml.Unmarshal(data, &existingMap); err != nil {
			return fmt.Errorf("parse existing config failed: %w", err)
		}
	}

	providerMap, _ := existingMap["model_providers"].(map[string]any)
	if providerMap == nil {
		providerMap = make(map[string]any)
	}
	baseURL, _ := profileMap.String(codexKeyBaseURL)
	wireAPI, _ := profileMap.String(codexKeyWireAPI)
	providerEntryMap := map[string]any{
		"base_url": baseURL,
		"name":     name,
		"wire_api": wireAPI,
	}
	if auth, ok := profileMap.Bool(codexKeyRequiresOpenAIAuth); ok {
		providerEntryMap[codexKeyRequiresOpenAIAuth] = auth
	}
	providerMap[name] = providerEntryMap

	body, providerOrderList := stripManagedCodexConfig(original, name)
	rendered, err := renderCodexConfig(body, providerMap, providerOrderList)
	if err != nil {
		return err
	}
	rendered = normalizeCodexTableSpacing(rendered)

	return fileutil.AtomicWrite(path, bytes.TrimRight([]byte(rendered), "\n"), fileutil.ConfigFilePermission)
}

func (c *Codex) writeAuthJson(profileMap config.Profile) error {
	path := c.pm.HomePath(codexConfigDir, codexAuthFile)

	existingMap := make(map[string]any)
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &existingMap); err != nil {
			return fmt.Errorf("parse existing auth failed: %w", err)
		}
	}

	if apiKey, ok := profileMap[codexKeyOpenAIAPIKey]; ok {
		existingMap[codexKeyOpenAIAPIKey] = apiKey
	}

	authData, err := fileutil.MarshalJSONNoTrailingNewline(existingMap)
	if err != nil {
		return err
	}

	return fileutil.AtomicWrite(path, authData, fileutil.AuthFilePermission)
}

var codexTableHeaderPattern = regexp.MustCompile(`^\s*\[([^\[\]]+)\]\s*$`)

func stripManagedCodexConfig(original, selectedProvider string) (string, []string) {
	lineList := splitLinesKeepNewline(original)
	if len(lineList) == 0 {
		return fmt.Sprintf("%s = %q\n", codexKeyModelProvider, selectedProvider), nil
	}

	var (
		outList            []string
		providerOrderList  []string
		currentSection     string
		skippingProviders  bool
		replacedTopLevelKV bool
	)

	for _, line := range lineList {
		if header, ok := parseCodexTableHeader(line); ok {
			currentSection = header

			if providerName, managed := parseModelProvidersHeader(header); managed {
				skippingProviders = true
				if providerName != "" {
					providerOrderList = appendUnique(providerOrderList, providerName)
				}
				continue
			}

			skippingProviders = false
			outList = append(outList, line)
			continue
		}

		if skippingProviders {
			continue
		}

		if currentSection == "" && isTopLevelKeyAssignment(line, codexKeyModelProvider) {
			outList = append(outList, fmt.Sprintf("%s = %q\n", codexKeyModelProvider, selectedProvider))
			replacedTopLevelKV = true
			continue
		}

		outList = append(outList, line)
	}

	if !replacedTopLevelKV {
		outList = insertTopLevelKey(outList, codexKeyModelProvider, selectedProvider)
	}

	return strings.TrimRight(strings.Join(outList, ""), "\n"), providerOrderList
}

func renderCodexConfig(body string, providerMap map[string]any, providerOrderList []string) (string, error) {
	var buf bytes.Buffer

	if body != "" {
		buf.WriteString(body)
	}

	for _, name := range orderedProviderNames(providerMap, providerOrderList) {
		profileMap, _ := providerMap[name].(map[string]any)
		if profileMap == nil {
			continue
		}

		if buf.Len() > 0 && buf.Bytes()[buf.Len()-1] != '\n' {
			buf.WriteString("\n")
		}

		buf.WriteString(fmt.Sprintf("[model_providers.%q]\n", name))
		enc := toml.NewEncoder(&buf)
		if err := enc.Encode(profileMap); err != nil {
			return "", fmt.Errorf("encode model_providers.%s failed: %w", name, err)
		}
	}

	return buf.String(), nil
}

func normalizeCodexTableSpacing(text string) string {
	lineList := splitLinesKeepNewline(text)
	if len(lineList) == 0 {
		return ""
	}

	outList := make([]string, 0, len(lineList))
	for _, line := range lineList {
		if _, ok := parseCodexTableHeader(line); ok {
			for len(outList) > 0 && strings.TrimSpace(outList[len(outList)-1]) == "" {
				outList = outList[:len(outList)-1]
			}
			if len(outList) > 0 {
				outList = append(outList, "\n")
			}
		}
		outList = append(outList, line)
	}

	return strings.Join(outList, "")
}

func orderedProviderNames(providerMap map[string]any, existingOrderList []string) []string {
	seenMap := make(map[string]struct{}, len(existingOrderList))
	orderedList := make([]string, 0, len(providerMap))

	for _, name := range existingOrderList {
		if _, ok := providerMap[name]; ok {
			orderedList = append(orderedList, name)
			seenMap[name] = struct{}{}
		}
	}

	remainingList := make([]string, 0, len(providerMap))
	for name := range providerMap {
		if _, ok := seenMap[name]; ok {
			continue
		}
		remainingList = append(remainingList, name)
	}
	sort.Strings(remainingList)

	return append(orderedList, remainingList...)
}

func splitLinesKeepNewline(text string) []string {
	if text == "" {
		return nil
	}
	return strings.SplitAfter(text, "\n")
}

func parseCodexTableHeader(line string) (string, bool) {
	matchList := codexTableHeaderPattern.FindStringSubmatch(line)
	if len(matchList) != 2 {
		return "", false
	}
	return strings.TrimSpace(matchList[1]), true
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

func insertTopLevelKey(lineList []string, key, value string) []string {
	insertLine := fmt.Sprintf("%s = %q\n", key, value)
	if len(lineList) == 0 {
		return []string{insertLine}
	}

	firstHeader := len(lineList)
	for i, line := range lineList {
		if _, ok := parseCodexTableHeader(line); ok {
			firstHeader = i
			break
		}
	}

	insertAt := firstHeader
	for insertAt > 0 && strings.TrimSpace(lineList[insertAt-1]) == "" {
		insertAt--
	}

	lineList = append(lineList[:insertAt], append([]string{insertLine}, lineList[insertAt:]...)...)
	return lineList
}

func appendUnique(valueList []string, candidate string) []string {
	for _, value := range valueList {
		if value == candidate {
			return valueList
		}
	}
	return append(valueList, candidate)
}
