package ui

import (
	"agent-envs/internal/agent"
	"strings"
)

// RenderSelector 渲染代理选择视图
func RenderSelector(agentList []agent.Agent, cursor int, message string, msgIsErr bool) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚡ 选择代理类型"))
	b.WriteString("\n\n")

	for i, ag := range agentList {
		isCursor := i == cursor
		var prefix string
		if isCursor {
			prefix = CursorStyle.Render("▸ ")
		} else {
			prefix = "  "
		}

		var line string
		if isCursor {
			line = prefix + SelectedItemStyle.Render(ag.Name()) + " " + LabelStyle.Render("("+ag.Description()+")")
		} else {
			line = prefix + NormalItemStyle.Render(ag.Name()) + " " + LabelStyle.Render("("+ag.Description()+")")
		}
		b.WriteString(line)
		b.WriteString("\n")
	}

	if message != "" {
		b.WriteString("\n")
		if msgIsErr {
			b.WriteString(ErrorStyle.Render("✗ " + message))
		} else {
			b.WriteString(SuccessStyle.Render("✓ " + message))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓ 移动  •  Enter 选择  •  Esc/q 退出"))
	b.WriteString("\n")

	return b.String()
}
