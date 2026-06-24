package agent

import "agent-envs/internal/config"

// ProfileInput 是 UI 收集到的通用服务商信息。
type ProfileInput struct {
	APIURL string
	Token  string
}

// ProfileSummary 是配置列表中展示的通用信息。
type ProfileSummary struct {
	URL   string
	Token string
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

	// BuildProfile 将通用输入转换为该代理自己的配置结构。
	BuildProfile(input ProfileInput) config.Profile

	// SummarizeProfile 返回配置列表中展示的 URL 和 token。
	SummarizeProfile(profileMap config.Profile) ProfileSummary
}

// Available 返回当前内置支持的代理列表。
func Available(pm *config.PathManager) []Agent {
	return []Agent{
		NewClaude(pm),
		NewCodex(pm),
	}
}
