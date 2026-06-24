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
		b.WriteString(HelpStyle.Render("按 a 添加配置"))
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

		indent := "    "
		for _, item := range ag.ProfileSummaryItemList(profileMap) {
			value := valueOrNA(item.Value)
			style := URLStyle
			if item.Secret {
				value = maskSecret(value)
				style = TokenStyle
			}
			b.WriteString(fmt.Sprintf("%s%s %s\n",
				indent,
				LabelStyle.Render(item.Label+":"),
				style.Render(value)))
		}

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
	b.WriteString(HelpStyle.Render("↑/↓ 移动  •  Enter 切换  •  a 添加  •  e 修改  •  d 删除  •  Esc 返回  •  q 退出"))
	b.WriteString("\n")

	return b.String()
}

// RenderProfileForm 渲染添加或修改配置视图
func RenderProfileForm(action string, agentName string, step profileFormStep, name string, fieldList []agent.ProfileField, valueMap map[string]string, input string, cursor int, message string, msgIsErr bool) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚡ " + action + " " + agentName + " 配置"))
	b.WriteString("\n")

	writeAddLine(&b, "名称", name, input, cursor, step == profileFormStepName, false)
	for i, field := range fieldList {
		fieldStep := profileFormStep(i + 1)
		writeAddLine(&b, field.Label, valueMap[field.Key], input, cursor, step == fieldStep, field.Secret)
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

	submitAction := "下一步"
	if isLastProfileFormStep(step, fieldList) {
		submitAction = "保存"
	}
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("Enter " + submitAction + "  •  ←/→ 移动  •  Ctrl+U 清空  •  Esc 取消"))
	b.WriteString("\n")

	return b.String()
}

// RenderDeleteProfile 渲染删除确认视图
func RenderDeleteProfile(agentName string, name string, active bool, message string, msgIsErr bool) string {
	var b strings.Builder

	b.WriteString(TitleStyle.Render("⚡ 删除 " + agentName + " 配置"))
	b.WriteString("\n")
	b.WriteString(LabelStyle.Render("将删除配置: "))
	b.WriteString(SelectedItemStyle.Render(name))
	b.WriteString("\n")
	if active {
		b.WriteString(ErrorStyle.Render("这是当前激活配置，删除后将清空 active"))
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
	b.WriteString(HelpStyle.Render("y 确认删除  •  n/Esc 取消  •  q 退出"))
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

func isLastProfileFormStep(step profileFormStep, fieldList []agent.ProfileField) bool {
	return int(step) >= len(fieldList)
}

func maskSecret(value string) string {
	if value == "" || value == "N/A" {
		return value
	}
	valueRuneList := []rune(value)
	if len(valueRuneList) <= 12 {
		return "****"
	}
	return string(valueRuneList[:8]) + "****" + string(valueRuneList[len(valueRuneList)-4:])
}
