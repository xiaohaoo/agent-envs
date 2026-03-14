package ui

import "strings"

// AgentOption 代理选项
type AgentOption struct {
	Name string
	Desc string
}

var agentOptions = []AgentOption{
	{"Claude Code", "Anthropic Claude Code"},
	{"Codex", "Codex CLI"},
}

// RenderSelector 渲染代理选择视图
func RenderSelector(cursor int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚡ 选择代理类型"))
	b.WriteString("\n\n")

	for i, option := range agentOptions {
		isCursor := i == cursor
		var prefix string
		if isCursor {
			prefix = CursorStyle.Render("▸ ")
		} else {
			prefix = "  "
		}

		var line string
		if isCursor {
			line = prefix + SelectedItemStyle.Render(option.Name) + " " + LabelStyle.Render("("+option.Desc+")")
		} else {
			line = prefix + NormalItemStyle.Render(option.Name) + " " + LabelStyle.Render("("+option.Desc+")")
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓ 移动  •  Enter 选择  •  Esc/q 退出"))
	b.WriteString("\n")

	return b.String()
}
