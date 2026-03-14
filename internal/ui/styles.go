package ui

import "github.com/charmbracelet/lipgloss"

// 主题色 - 明亮配色
var (
	PrimaryColor = lipgloss.Color("#61AFEF") // 亮蓝
	AccentColor  = lipgloss.Color("#56B6C2") // 青色
	SuccessColor = lipgloss.Color("#98C379") // 绿色
	ErrorColor   = lipgloss.Color("#E06C75") // 红色
)

// 文字颜色
var (
	TextBright = lipgloss.Color("#D4D4D4") // 更亮的灰
	TextNormal = lipgloss.Color("#CCCCCC") // 亮灰
	TextMuted  = lipgloss.Color("#9CA3AF") // 中灰
	TextDim    = lipgloss.Color("#6B7280") // 暗灰
)

// 样式定义
var (
	// TitleStyle 标题样式 - 简洁无背景
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1)

	// SelectedItemStyle 选中项样式 - 柔和高亮
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(PrimaryColor).
				Bold(true)

	// NormalItemStyle 普通项样式
	NormalItemStyle = lipgloss.NewStyle().
			Foreground(TextNormal)

	// ActiveMarkerStyle 激活标记样式
	ActiveMarkerStyle = lipgloss.NewStyle().
				Foreground(SuccessColor).
				Bold(true)

	// CursorStyle 光标指示器
	CursorStyle = lipgloss.NewStyle().
			Foreground(PrimaryColor).
			Bold(true)

	// URLStyle URL 样式
	URLStyle = lipgloss.NewStyle().
			Foreground(AccentColor)

	// TokenStyle Token 样式
	TokenStyle = lipgloss.NewStyle().
			Foreground(TextMuted)

	// LabelStyle 标签样式
	LabelStyle = lipgloss.NewStyle().
			Foreground(TextDim)

	// SuccessStyle 成功消息
	SuccessStyle = lipgloss.NewStyle().
			Foreground(SuccessColor).
			Bold(true)

	// ErrorStyle 错误消息
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ErrorColor).
			Bold(true)

	// HelpStyle 帮助文本
	HelpStyle = lipgloss.NewStyle().
			Foreground(TextMuted).
			MarginTop(1)

	// DividerStyle 分隔线
	DividerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#9CA3AF")) // 更浅的灰色
)
