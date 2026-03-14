package main

import (
	"agent-envs/internal/config"
	"agent-envs/internal/ui"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

// 通过 ldflags 注入的版本信息
var (
	version   = "dev"
	buildTime = "unknown"
)

func main() {
	// 处理 --version 参数
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("agent-envs %s (built %s)\n", version, buildTime)
		return
	}
	// 创建路径管理器
	pm, err := config.NewPathManager()
	if err != nil {
		fmt.Fprintf(os.Stderr, "初始化失败: %v\n", err)
		os.Exit(1)
	}

	// 创建 UI 模型
	m := ui.NewModel(pm)

	// 启动 Bubble Tea 程序
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "启动失败: %v\n", err)
		os.Exit(1)
	}
}
