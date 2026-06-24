package ui

import (
	"agent-envs/internal/agent"
	"agent-envs/internal/config"
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeAgent struct {
	key               string
	name              string
	savedCfg          *config.Config
	appliedName       string
	appliedProfileMap config.Profile
	applyCount        int
	applyErr          error
	saveErr           error
}

func (f *fakeAgent) Key() string {
	return f.key
}

func (f *fakeAgent) Name() string {
	return f.name
}

func (f *fakeAgent) Description() string {
	return "fake agent"
}

func (f *fakeAgent) LoadConfig() (*config.Config, error) {
	return &config.Config{ProfileMap: make(map[string]config.Profile)}, nil
}

func (f *fakeAgent) SaveConfig(cfg *config.Config) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.savedCfg = cfg
	return nil
}

func (f *fakeAgent) ApplyProfile(name string, profileMap config.Profile) error {
	if f.applyErr != nil {
		return f.applyErr
	}
	f.appliedName = name
	f.appliedProfileMap = profileMap
	f.applyCount++
	return nil
}

func (f *fakeAgent) ProfileFieldList() []agent.ProfileField {
	return []agent.ProfileField{
		{Key: "api_url", Label: "Endpoint"},
		{Key: "token", Label: "Secret", Secret: true},
	}
}

func (f *fakeAgent) BuildProfile(input agent.ProfileInput) config.Profile {
	return config.Profile{
		"api_url": input.FieldValueMap["api_url"],
		"token":   input.FieldValueMap["token"],
	}
}

func (f *fakeAgent) ProfileSummaryItemList(profileMap config.Profile) []agent.ProfileSummaryItem {
	url, _ := profileMap.String("api_url")
	token, _ := profileMap.String("token")
	return []agent.ProfileSummaryItem{
		{Label: "Endpoint", Value: url},
		{Label: "Secret", Value: token, Secret: true},
	}
}

func TestAddProfileUsesAgentBuilder(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := Model{
		agent:     fakeAgent,
		cfg:       &config.Config{ProfileMap: make(map[string]config.Profile)},
		selecting: false,
	}

	model = updateModel(t, model, keyRunes("a"))
	if !model.adding {
		t.Fatal("expected add form to be active")
	}

	model = updateModel(t, model, keyRunes("测试服"))
	model = updateModel(t, model, keyType(tea.KeyEnter))
	model = updateModel(t, model, keyRunes("https://api.example.com"))
	model = updateModel(t, model, keyType(tea.KeyEnter))
	model = updateModel(t, model, keyRunes("sk-test-token"))
	model = updateModel(t, model, keyType(tea.KeyEnter))

	if model.adding {
		t.Fatal("expected add form to close after token submit")
	}
	if fakeAgent.savedCfg == nil {
		t.Fatal("expected config to be saved")
	}

	profileMap := fakeAgent.savedCfg.ProfileMap["测试服"]
	if profileMap == nil {
		t.Fatal("expected added profile")
	}
	if got, _ := profileMap.String("api_url"); got != "https://api.example.com" {
		t.Fatalf("api_url = %q", got)
	}
	if got, _ := profileMap.String("token"); got != "sk-test-token" {
		t.Fatalf("token = %q", got)
	}
	if fakeAgent.savedCfg.Active != "" {
		t.Fatalf("active = %q, want empty before explicit switch", fakeAgent.savedCfg.Active)
	}
}

func TestEditProfileRenamesAndPreservesExtraFields(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := Model{
		agent: fakeAgent,
		cfg: &config.Config{
			Active: "旧配置",
			ProfileMap: map[string]config.Profile{
				"旧配置": {
					"api_url": "https://old.example.com",
					"token":   "old-token",
					"extra":   "keep-me",
				},
			},
		},
		nameList:  []string{"旧配置"},
		selecting: false,
	}

	model = updateModel(t, model, keyRunes("e"))
	if !model.editing {
		t.Fatal("expected edit form to be active")
	}

	model = updateModel(t, model, keyType(tea.KeyCtrlU))
	model = updateModel(t, model, keyRunes("新配置"))
	model = updateModel(t, model, keyType(tea.KeyEnter))
	model = updateModel(t, model, keyType(tea.KeyCtrlU))
	model = updateModel(t, model, keyRunes("https://new.example.com"))
	model = updateModel(t, model, keyType(tea.KeyEnter))
	model = updateModel(t, model, keyType(tea.KeyCtrlU))
	model = updateModel(t, model, keyRunes("new-token"))
	model = updateModel(t, model, keyType(tea.KeyEnter))

	if model.editing {
		t.Fatal("expected edit form to close")
	}
	if fakeAgent.savedCfg == nil {
		t.Fatal("expected config to be saved")
	}
	if _, exists := fakeAgent.savedCfg.ProfileMap["旧配置"]; exists {
		t.Fatal("expected old profile name to be removed")
	}
	profileMap := fakeAgent.savedCfg.ProfileMap["新配置"]
	if profileMap == nil {
		t.Fatal("expected renamed profile")
	}
	if got, _ := profileMap.String("api_url"); got != "https://new.example.com" {
		t.Fatalf("api_url = %q", got)
	}
	if got, _ := profileMap.String("token"); got != "new-token" {
		t.Fatalf("token = %q", got)
	}
	if got, _ := profileMap.String("extra"); got != "keep-me" {
		t.Fatalf("extra = %q", got)
	}
	if fakeAgent.savedCfg.Active != "新配置" {
		t.Fatalf("active = %q, want 新配置", fakeAgent.savedCfg.Active)
	}
	if fakeAgent.appliedName != "新配置" {
		t.Fatalf("appliedName = %q, want 新配置", fakeAgent.appliedName)
	}
	if fakeAgent.applyCount != 1 {
		t.Fatalf("applyCount = %d, want 1", fakeAgent.applyCount)
	}
	if got, _ := fakeAgent.appliedProfileMap.String("token"); got != "new-token" {
		t.Fatalf("applied token = %q", got)
	}
}

func TestDeleteActiveProfileClearsActive(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := Model{
		agent: fakeAgent,
		cfg: &config.Config{
			Active: "old",
			ProfileMap: map[string]config.Profile{
				"old": {"token": "old-token"},
				"new": {"token": "new-token"},
			},
		},
		nameList:  []string{"old", "new"},
		cursor:    0,
		selecting: false,
	}

	model = updateModel(t, model, keyRunes("d"))
	if !model.deleting {
		t.Fatal("expected delete confirmation")
	}
	model = updateModel(t, model, keyRunes("y"))

	if model.deleting {
		t.Fatal("expected delete confirmation to close")
	}
	if fakeAgent.savedCfg == nil {
		t.Fatal("expected config to be saved")
	}
	if _, exists := fakeAgent.savedCfg.ProfileMap["old"]; exists {
		t.Fatal("expected old profile to be deleted")
	}
	if fakeAgent.savedCfg.Active != "" {
		t.Fatalf("active = %q, want empty", fakeAgent.savedCfg.Active)
	}
	if len(model.nameList) != 1 || model.nameList[0] != "new" {
		t.Fatalf("nameList = %#v", model.nameList)
	}
}

func TestDeleteProfileCanBeCancelled(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := switchModel(fakeAgent)
	model.cursor = 0

	model = updateModel(t, model, keyRunes("d"))
	model = updateModel(t, model, keyRunes("n"))

	if model.deleting {
		t.Fatal("expected delete confirmation to close")
	}
	if fakeAgent.savedCfg != nil {
		t.Fatal("expected config not to be saved")
	}
	if _, exists := model.cfg.ProfileMap["old"]; !exists {
		t.Fatal("expected profile to remain")
	}
}

func TestDeleteProfileRequiresExplicitY(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := switchModel(fakeAgent)
	model.cursor = 0

	model = updateModel(t, model, keyRunes("d"))
	model = updateModel(t, model, keyType(tea.KeyEnter))

	if !model.deleting {
		t.Fatal("expected delete confirmation to remain open")
	}
	if fakeAgent.savedCfg != nil {
		t.Fatal("expected config not to be saved")
	}
	if _, exists := model.cfg.ProfileMap["old"]; !exists {
		t.Fatal("expected profile to remain")
	}
}

func TestDeleteConfirmQQuits(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := switchModel(fakeAgent)

	model = updateModel(t, model, keyRunes("d"))
	model = updateModel(t, model, keyRunes("q"))

	if !model.quitting {
		t.Fatal("expected q to quit from delete confirmation")
	}
	if fakeAgent.savedCfg != nil {
		t.Fatal("expected config not to be saved")
	}
}

func TestDeleteActiveProfileViewWarnsAboutActive(t *testing.T) {
	view := RenderDeleteProfile("Demo Agent", "old", true, "", false)
	if !strings.Contains(view, "删除后将清空 active") {
		t.Fatalf("expected active delete warning in view: %q", view)
	}
	if strings.Contains(view, "Enter 确认删除") {
		t.Fatalf("expected enter not to confirm delete: %q", view)
	}
}

func TestSwitchDoesNotSaveActiveWhenApplyFails(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent", applyErr: errors.New("apply failed")}
	model := switchModel(fakeAgent)

	model.doSwitch()

	if model.cfg.Active != "old" {
		t.Fatalf("active = %q, want old", model.cfg.Active)
	}
	if fakeAgent.savedCfg != nil {
		t.Fatal("expected config not to be saved when apply fails")
	}
	if !model.msgIsErr {
		t.Fatal("expected error message")
	}
}

func TestSwitchRollsBackActiveWhenSaveFails(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent", saveErr: errors.New("save failed")}
	model := switchModel(fakeAgent)

	model.doSwitch()

	if model.cfg.Active != "old" {
		t.Fatalf("active = %q, want old", model.cfg.Active)
	}
	if !model.msgIsErr {
		t.Fatal("expected error message")
	}
}

func TestAddInputSupportsCursorEditing(t *testing.T) {
	model := Model{}
	model.startAdd()

	model = updateModel(t, model, keyRunes("abc"))
	model = updateModel(t, model, keyType(tea.KeyLeft))
	model = updateModel(t, model, keyType(tea.KeyLeft))
	model = updateModel(t, model, keyRunes("X"))

	if got := model.addInputValue(); got != "aXbc" {
		t.Fatalf("addInputValue() = %q", got)
	}
	if model.addCursor != 2 {
		t.Fatalf("addCursor = %d, want 2", model.addCursor)
	}
}

func TestCleanAddInputRemovesControlChars(t *testing.T) {
	got := cleanAddInput("\x00\x00https://code.wondervoice.hk\x07")
	if got != "https://code.wondervoice.hk" {
		t.Fatalf("cleanAddInput() = %q", got)
	}
}

func TestAddProfileRejectsInvalidName(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := Model{
		agent:     fakeAgent,
		cfg:       &config.Config{ProfileMap: make(map[string]config.Profile)},
		selecting: false,
	}

	model = updateModel(t, model, keyRunes("a"))
	model = updateModel(t, model, keyRunes("bad[name]"))
	model = updateModel(t, model, keyType(tea.KeyEnter))

	if !model.adding {
		t.Fatal("expected add form to remain active")
	}
	if !model.msgIsErr || !strings.Contains(model.message, "不能包含") {
		t.Fatalf("message = %q, msgIsErr = %v", model.message, model.msgIsErr)
	}
	if fakeAgent.savedCfg != nil {
		t.Fatal("expected config not to be saved")
	}
}

func TestAddProfileRejectsInvalidURL(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := Model{
		agent:     fakeAgent,
		cfg:       &config.Config{ProfileMap: make(map[string]config.Profile)},
		selecting: false,
	}

	model = updateModel(t, model, keyRunes("a"))
	model = updateModel(t, model, keyRunes("demo"))
	model = updateModel(t, model, keyType(tea.KeyEnter))
	model = updateModel(t, model, keyRunes("ftp://api.example.com"))
	model = updateModel(t, model, keyType(tea.KeyEnter))

	if !model.adding {
		t.Fatal("expected add form to remain active")
	}
	if !model.msgIsErr || !strings.Contains(model.message, "http 或 https") {
		t.Fatalf("message = %q, msgIsErr = %v", model.message, model.msgIsErr)
	}
	if fakeAgent.savedCfg != nil {
		t.Fatal("expected config not to be saved")
	}
}

func TestAddProfileRejectsSecretWhitespace(t *testing.T) {
	fakeAgent := &fakeAgent{key: "demo", name: "Demo Agent"}
	model := Model{
		agent:     fakeAgent,
		cfg:       &config.Config{ProfileMap: make(map[string]config.Profile)},
		selecting: false,
	}

	model = updateModel(t, model, keyRunes("a"))
	model = updateModel(t, model, keyRunes("demo"))
	model = updateModel(t, model, keyType(tea.KeyEnter))
	model = updateModel(t, model, keyRunes("https://api.example.com"))
	model = updateModel(t, model, keyType(tea.KeyEnter))
	model = updateModel(t, model, keyRunes("sk test"))
	model = updateModel(t, model, keyType(tea.KeyEnter))

	if !model.adding {
		t.Fatal("expected add form to remain active")
	}
	if !model.msgIsErr || !strings.Contains(model.message, "空白字符") {
		t.Fatalf("message = %q, msgIsErr = %v", model.message, model.msgIsErr)
	}
	if fakeAgent.savedCfg != nil {
		t.Fatal("expected config not to be saved")
	}
}

func switchModel(fakeAgent *fakeAgent) Model {
	return Model{
		agent: fakeAgent,
		cfg: &config.Config{
			Active: "old",
			ProfileMap: map[string]config.Profile{
				"old": {"token": "old-token"},
				"new": {"token": "new-token"},
			},
		},
		nameList:  []string{"old", "new"},
		cursor:    1,
		selecting: false,
	}
}

func updateModel(t *testing.T, model Model, msg tea.Msg) Model {
	t.Helper()

	updatedModel, _ := model.Update(msg)
	nextModel, ok := updatedModel.(Model)
	if !ok {
		t.Fatalf("updated model type = %T", updatedModel)
	}
	return nextModel
}

func keyRunes(value string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func keyType(keyType tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: keyType}
}
