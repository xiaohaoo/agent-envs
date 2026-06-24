package agent

import "agent-envs/internal/config"

// ProfileField 是创建或修改 profile 时需要用户填写的字段。
type ProfileField struct {
	Key    string
	Label  string
	Secret bool
}

// ProfileInput 是 UI 收集到的 profile 字段值。
type ProfileInput struct {
	FieldValueMap map[string]string
}

// ProfileSummaryItem 是配置列表中展示的一行摘要。
type ProfileSummaryItem struct {
	Label  string
	Value  string
	Secret bool
}

// Agent 抽象不同代理的配置管理。
type Agent interface {
	// Key 返回统一配置文件中的 agent 段名。
	Key() string

	// Name 返回代理名称（用于显示）。
	Name() string

	// Description 返回代理描述（用于选择页显示）。
	Description() string

	// LoadConfig 加载代理的配置。
	LoadConfig() (*config.Config, error)

	// SaveConfig 保存代理的配置。
	SaveConfig(cfg *config.Config) error

	// ApplyProfile 将 profile 应用到代理的设置文件。
	ApplyProfile(name string, profileMap config.Profile) error

	// ProfileFieldList 返回创建或修改 profile 时需要填写的字段。
	ProfileFieldList() []ProfileField

	// BuildProfile 将 UI 输入转换为该代理自己的配置结构。
	BuildProfile(input ProfileInput) config.Profile

	// ProfileSummaryItemList 返回配置列表中展示的摘要字段。
	ProfileSummaryItemList(profileMap config.Profile) []ProfileSummaryItem
}

// Available 返回当前内置支持的代理列表。
func Available(pm *config.PathManager) []Agent {
	return []Agent{
		NewClaude(pm),
		NewCodex(pm),
	}
}
