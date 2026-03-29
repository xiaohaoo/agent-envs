package ui

import (
	"agent-envs/internal/config"
	"fmt"
	"strings"
)

// RenderProfiles 渲染配置列表视图
func RenderProfiles(agentName string, cfg *config.Config, names []string, cursor int, message string, msgIsErr bool, width int) string {
	var b strings.Builder

	// 标题
	b.WriteString(TitleStyle.Render("⚡ " + agentName + " Envs"))
	b.WriteString("\n")

	// 配置列表
	for i, name := range names {
		profile := cfg.Profiles[name]
		isActive := name == cfg.Active
		isCursor := i == cursor

		// 构建前缀: 光标 + 激活标记
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

		// 名称行 - 紧凑无多余空格
		var nameLine string
		if isCursor {
			nameLine = prefix + " " + SelectedItemStyle.Render(name)
		} else {
			nameLine = prefix + " " + NormalItemStyle.Render(name)
		}

		b.WriteString(nameLine)
		b.WriteString("\n")

		// 详情行 - 对齐缩进
		url, token := extractProfileInfo(profile)

		indent := "    "
		b.WriteString(fmt.Sprintf("%s%s %s\n",
			indent,
			LabelStyle.Render("URL:"),
			URLStyle.Render(url)))
		b.WriteString(fmt.Sprintf("%s%s %s\n",
			indent,
			LabelStyle.Render("Key:"),
			TokenStyle.Render(token)))

		// 分隔线 - 根据终端宽度动态生成
		if i < len(names)-1 {
			dividerWidth := width - 4 // 减去缩进的 4 个空格
			if dividerWidth < 10 {
				dividerWidth = 35 // 最小宽度
			}
			divider := strings.Repeat("─", dividerWidth)
			b.WriteString(DividerStyle.Render(indent + divider))
			b.WriteString("\n")
		}
	}

	// 操作消息
	if message != "" {
		b.WriteString("\n")
		if msgIsErr {
			b.WriteString(ErrorStyle.Render("✗ " + message))
		} else {
			b.WriteString(SuccessStyle.Render("✓ " + message))
		}
		b.WriteString("\n")
	}

	// 帮助栏
	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑/↓ 移动  •  Enter 切换  •  Esc 返回  •  q 退出"))
	b.WriteString("\n")

	return b.String()
}

// extractProfileInfo 从 profile 中提取 URL 和 token 信息
func extractProfileInfo(profile config.Profile) (url, token string) {
	// 尝试 Claude 格式
	if val, ok := profile.String(config.KeyAnthropicBaseURL); ok {
		url = val
		token = profile.MaskToken()
		return
	}

	// 尝试 Codex 格式
	if val, ok := profile.String(config.KeyBaseURL); ok {
		url = val
		token = profile.MaskToken()
		return
	}

	return "N/A", "N/A"
}
