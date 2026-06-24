package ui

import (
	"agent-envs/internal/agent"
	"agent-envs/internal/config"
	"fmt"
	"strings"
)

// RenderProfiles 渲染配置列表视图
func RenderProfiles(ag agent.Agent, cfg *config.Config, nameList []string, cursor int, message string, msgIsErr bool, width int) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚡ " + ag.Name() + " Envs"))
	b.WriteString("\n")

	if len(nameList) == 0 {
		b.WriteString(LabelStyle.Render("暂无配置"))
		b.WriteString("\n")
		b.WriteString(HelpStyle.Render("按 a 添加 API 地址和 Token"))
		b.WriteString("\n")
	}
	for i, name := range nameList {
		profileMap := cfg.ProfileMap[name]
		isActive := name == cfg.Active
		isCursor := i == cursor

		var prefix string
		if isCursor {
			prefix = CursorStyle.Render("▸ ")
		} else {
			prefix = "  "
		}

		if isActive {
			prefix += ActiveMarkerStyle.Render("●")
		} else {
			prefix += " "
		}

		var nameLine string
		if isCursor {
			nameLine = prefix + " " + SelectedItemStyle.Render(name)
		} else {
			nameLine = prefix + " " + NormalItemStyle.Render(name)
		}

		b.WriteString(nameLine)
		b.WriteString("\n")

		summary := ag.SummarizeProfile(profileMap)

		indent := "    "
		b.WriteString(fmt.Sprintf("%s%s %s\n",
			indent,
			LabelStyle.Render("URL:"),
			URLStyle.Render(valueOrNA(summary.URL))))
		b.WriteString(fmt.Sprintf("%s%s %s\n",
			indent,
			LabelStyle.Render("Key:"),
			TokenStyle.Render(valueOrNA(summary.Token))))

		if i < len(nameList)-1 {
			dividerWidth := width - 4
			if dividerWidth < 10 {
				dividerWidth = 35
			}
			divider := strings.Repeat("─", dividerWidth)
			b.WriteString(DividerStyle.Render(indent + divider))
			b.WriteString("\n")
		}
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
	b.WriteString(HelpStyle.Render("↑/↓ 移动  •  Enter 切换  •  a 添加  •  Esc 返回  •  q 退出"))
	b.WriteString("\n")

	return b.String()
}

// RenderAddProfile 渲染添加配置视图
func RenderAddProfile(agentName string, step addStep, name string, apiURL string, token string, input string, cursor int, message string, msgIsErr bool) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚡ 添加 " + agentName + " 配置"))
	b.WriteString("\n")

	writeAddLine(&b, "名称", name, input, cursor, step == addStepName, false)
	writeAddLine(&b, "API", apiURL, input, cursor, step == addStepURL, false)
	writeAddLine(&b, "Token", token, input, cursor, step == addStepToken, true)

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
	b.WriteString(HelpStyle.Render("Enter 下一步  •  ←/→ 移动  •  Ctrl+U 清空  •  Esc 取消"))
	b.WriteString("\n")

	return b.String()
}

func writeAddLine(b *strings.Builder, label string, savedValue string, input string, cursor int, active bool, mask bool) {
	value := savedValue
	if active {
		value = renderInputWithCursor(input, cursor, mask)
	} else if mask && value != "" {
		value = strings.Repeat("*", len([]rune(value)))
	}
	if value == "" {
		value = "待输入"
	}

	prefix := "  "
	style := NormalItemStyle
	if active {
		prefix = CursorStyle.Render("▸ ")
		style = SelectedItemStyle
	}

	b.WriteString(prefix)
	b.WriteString(LabelStyle.Render(label + ": "))
	b.WriteString(style.Render(value))
	b.WriteString("\n")
}

func renderInputWithCursor(input string, cursor int, mask bool) string {
	inputRuneList := []rune(input)
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(inputRuneList) {
		cursor = len(inputRuneList)
	}

	before := string(inputRuneList[:cursor])
	after := string(inputRuneList[cursor:])
	if mask {
		before = strings.Repeat("*", cursor)
		after = strings.Repeat("*", len(inputRuneList)-cursor)
	}
	return before + "▌" + after
}

func valueOrNA(value string) string {
	if value == "" {
		return "N/A"
	}
	return value
}
