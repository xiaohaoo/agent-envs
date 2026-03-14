package ui

import (
	"agent-envs/internal/agent"
	"agent-envs/internal/config"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Model Bubble Tea 模型
type Model struct {
	pm        *config.PathManager // 路径管理器
	agent     agent.Agent         // 当前代理
	cfg       *config.Config      // 当前配置
	names     []string            // 排序后的 profile 名称
	cursor    int                 // 当前光标位置
	message   string              // 操作结果消息
	msgIsErr  bool                // 消息是否为错误
	quitting  bool                // 是否退出
	selecting bool                // 是否在选择代理类型阶段
}

// NewModel 创建新的 UI 模型
func NewModel(pm *config.PathManager) Model {
	return Model{
		pm:        pm,
		selecting: true,
		cursor:    0,
	}
}

// Init 初始化模型
func (m Model) Init() tea.Cmd {
	return nil
}

// Update 处理消息更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			// q 键直接退出
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.selecting {
				// 在选择代理类型阶段，ESC 退出
				m.quitting = true
				return m, tea.Quit
			} else {
				// 在配置列表阶段，ESC 返回到选择代理类型
				m.selecting = true
				m.cursor = 0
				m.message = ""
				m.cfg = nil
				m.names = nil
				m.agent = nil
			}

		case "up", "k":
			if m.selecting {
				// 在选择代理类型阶段，只有两个选项
				if m.cursor > 0 {
					m.cursor--
				}
			} else {
				if m.cursor > 0 {
					m.cursor--
				}
			}

		case "down", "j":
			if m.selecting {
				// 在选择代理类型阶段，只有两个选项
				if m.cursor < 1 {
					m.cursor++
				}
			} else {
				if m.cursor < len(m.names)-1 {
					m.cursor++
				}
			}

		case "enter", " ":
			if m.selecting {
				// 选择代理类型
				var agentType agent.Type
				if m.cursor == 0 {
					agentType = agent.TypeClaude
				} else {
					agentType = agent.TypeCodex
				}

				// 创建代理实例
				ag, err := agent.New(agentType, m.pm)
				if err != nil {
					m.message = fmt.Sprintf("创建代理失败: %v", err)
					m.msgIsErr = true
					m.quitting = true
					return m, tea.Quit
				}
				m.agent = ag

				// 加载对应的配置
				cfg, err := ag.LoadConfig()
				if err != nil {
					m.message = fmt.Sprintf("加载配置失败: %v", err)
					m.msgIsErr = true
					m.quitting = true
					return m, tea.Quit
				}
				m.cfg = cfg
				m.names = cfg.SortedNames()
				m.cursor = 0

				// 光标默认指向当前激活的 profile
				for i, name := range m.names {
					if name == cfg.Active {
						m.cursor = i
						break
					}
				}
				m.selecting = false
			} else {
				m.doSwitch()
			}
		}
	}

	return m, nil
}

// View 渲染视图
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	if m.selecting {
		return RenderSelector(m.cursor)
	}

	return RenderProfiles(m.agent.Name(), m.cfg, m.names, m.cursor, m.message, m.msgIsErr)
}

// doSwitch 执行切换配置的操作
func (m *Model) doSwitch() {
	name := m.names[m.cursor]
	if name == m.cfg.Active {
		m.message = fmt.Sprintf("「%s」已经是当前配置", name)
		m.msgIsErr = false
		return
	}

	m.cfg.Active = name
	if err := m.agent.SaveConfig(m.cfg); err != nil {
		m.message = fmt.Sprintf("保存配置失败: %v", err)
		m.msgIsErr = true
		return
	}

	profile := m.cfg.Profiles[name]
	if err := m.agent.ApplyProfile(profile); err != nil {
		m.message = fmt.Sprintf("写入设置失败: %v", err)
		m.msgIsErr = true
		return
	}

	m.message = fmt.Sprintf("✓ 已切换到「%s」", name)
	m.msgIsErr = false
}
