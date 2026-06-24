package ui

import (
	"agent-envs/internal/agent"
	"agent-envs/internal/config"
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type fakeAgent struct {
	key      string
	name     string
	savedCfg *config.Config
	applyErr error
	saveErr  error
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
	return nil
}

func (f *fakeAgent) BuildProfile(input agent.ProfileInput) config.Profile {
	return config.Profile{
		"api_url": input.APIURL,
		"token":   input.Token,
	}
}

func (f *fakeAgent) SummarizeProfile(profileMap config.Profile) agent.ProfileSummary {
	url, _ := profileMap.String("api_url")
	return agent.ProfileSummary{URL: url, Token: profileMap.MaskToken()}
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
