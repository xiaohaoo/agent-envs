package agent

import (
	"agent-envs/internal/config"
	"fmt"
)

// Type 代理类型
type Type string

const (
	// TypeClaude Claude Code 代理
	TypeClaude Type = "claude"
	// TypeCodex Codex CLI 代理
	TypeCodex Type = "codex"
)

// Agent 代理接口，抽象不同代理的配置管理
type Agent interface {
	// Name 返回代理名称（用于显示）
	Name() string

	// LoadConfig 加载代理的配置
	LoadConfig() (*config.Config, error)

	// SaveConfig 保存代理的配置
	SaveConfig(cfg *config.Config) error

	// ApplyProfile 将 profile 应用到代理的设置文件
	ApplyProfile(profile config.Profile) error
}

// New 创建指定类型的代理实例
func New(t Type, pm *config.PathManager) (Agent, error) {
	switch t {
	case TypeClaude:
		return NewClaude(pm), nil
	case TypeCodex:
		return NewCodex(pm), nil
	default:
		return nil, fmt.Errorf("未知的代理类型: %s", t)
	}
}
