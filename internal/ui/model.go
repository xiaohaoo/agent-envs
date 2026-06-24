package ui

import (
	"agent-envs/internal/agent"
	"agent-envs/internal/config"
	"errors"
	"fmt"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
)

type addStep int

const (
	addStepName addStep = iota
	addStepURL
	addStepToken
)

// Model Bubble Tea 模型
type Model struct {
	pm               *config.PathManager // 路径管理器
	agentList        []agent.Agent       // 可选代理列表
	agent            agent.Agent         // 当前代理
	cfg              *config.Config      // 当前配置
	nameList         []string            // 排序后的 profile 名称
	cursor           int                 // 当前光标位置
	message          string              // 操作结果消息
	msgIsErr         bool                // 消息是否为错误
	quitting         bool                // 是否退出
	selecting        bool                // 是否在选择代理类型阶段
	width            int                 // 终端宽度
	adding           bool                // 是否在添加配置阶段
	addStep          addStep             // 当前添加步骤
	addName          string              // 待添加配置名称
	addURL           string              // 待添加 API 地址
	addToken         string              // 待添加 token
	addInputRuneList []rune              // 当前输入框内容
	addCursor        int                 // 添加表单输入光标位置
}

// NewModel 创建新的 UI 模型
func NewModel(pm *config.PathManager) Model {
	return Model{
		pm:        pm,
		agentList: agent.Available(pm),
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
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyMsg:
		if m.adding {
			return m.updateAddForm(msg)
		}

		switch msg.String() {
		case "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "q":
			m.quitting = true
			return m, tea.Quit

		case "esc":
			if m.selecting {
				m.quitting = true
				return m, tea.Quit
			}
			m.selecting = true
			m.cursor = 0
			m.message = ""
			m.cfg = nil
			m.nameList = nil
			m.agent = nil

		case "up", "k":
			if m.selecting {
				if len(m.agentList) > 0 {
					m.cursor = (m.cursor + len(m.agentList) - 1) % len(m.agentList)
				}
			} else {
				if len(m.nameList) == 0 {
					return m, nil
				}
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(m.nameList) - 1
				}
			}

		case "down", "j":
			if m.selecting {
				if len(m.agentList) > 0 {
					m.cursor = (m.cursor + 1) % len(m.agentList)
				}
			} else {
				if len(m.nameList) == 0 {
					return m, nil
				}
				if m.cursor < len(m.nameList)-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			}

		case "enter", " ":
			if m.selecting {
				return m.selectAgent()
			}
			m.doSwitch()

		case "a":
			if !m.selecting {
				m.startAdd()
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
		return RenderSelector(m.agentList, m.cursor, m.message, m.msgIsErr)
	}

	if m.adding {
		return RenderAddProfile(m.agent.Name(), m.addStep, m.addName, m.addURL, m.addToken, m.addInputValue(), m.addCursor, m.message, m.msgIsErr)
	}

	return RenderProfiles(m.agent, m.cfg, m.nameList, m.cursor, m.message, m.msgIsErr, m.width)
}

func (m Model) selectAgent() (tea.Model, tea.Cmd) {
	if len(m.agentList) == 0 {
		m.message = "暂无可用代理"
		m.msgIsErr = true
		return m, nil
	}
	if m.cursor < 0 || m.cursor >= len(m.agentList) {
		m.cursor = 0
	}

	ag := m.agentList[m.cursor]
	cfg, err := ag.LoadConfig()
	if err != nil {
		emptyCfg, emptyErr := emptyConfigIfNotFound(err)
		if emptyErr != nil {
			m.message = fmt.Sprintf("加载配置失败: %v", err)
			m.msgIsErr = true
			return m, nil
		}
		cfg = emptyCfg
	}

	m.agent = ag
	m.cfg = cfg
	m.nameList = cfg.SortedNames()
	m.cursor = 0
	for i, name := range m.nameList {
		if name == cfg.Active {
			m.cursor = i
			break
		}
	}
	m.message = ""
	m.msgIsErr = false
	m.selecting = false
	return m, nil
}

func (m Model) updateAddForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.quitting = true
		return m, tea.Quit
	case tea.KeyEsc:
		m.cancelAdd()
		return m, nil
	case tea.KeyEnter:
		m.submitAddInput()
		return m, nil
	case tea.KeyBackspace, tea.KeyCtrlH:
		m.deleteAddInputBeforeCursor()
		return m, nil
	case tea.KeyDelete, tea.KeyCtrlD:
		m.deleteAddInputAtCursor()
		return m, nil
	case tea.KeyCtrlU:
		m.resetAddInput()
		return m, nil
	case tea.KeyCtrlW:
		m.deleteAddInputWordBeforeCursor()
		return m, nil
	case tea.KeyLeft, tea.KeyCtrlB:
		if m.addCursor > 0 {
			m.addCursor--
		}
		return m, nil
	case tea.KeyRight, tea.KeyCtrlF:
		if m.addCursor < len(m.addInputRuneList) {
			m.addCursor++
		}
		return m, nil
	case tea.KeyHome, tea.KeyCtrlA:
		m.addCursor = 0
		return m, nil
	case tea.KeyEnd, tea.KeyCtrlE:
		m.addCursor = len(m.addInputRuneList)
		return m, nil
	case tea.KeySpace:
		m.insertAddInputRuneList([]rune{' '})
		return m, nil
	case tea.KeyRunes:
		m.insertAddInputRuneList(msg.Runes)
		return m, nil
	}

	return m, nil
}

func (m *Model) startAdd() {
	if m.cfg == nil {
		m.cfg = &config.Config{ProfileMap: make(map[string]config.Profile)}
	}
	if m.cfg.ProfileMap == nil {
		m.cfg.ProfileMap = make(map[string]config.Profile)
	}

	m.adding = true
	m.addStep = addStepName
	m.addName = ""
	m.addURL = ""
	m.addToken = ""
	m.resetAddInput()
	m.message = ""
	m.msgIsErr = false
}

func (m *Model) cancelAdd() {
	m.adding = false
	m.resetAddInput()
	m.message = "已取消添加配置"
	m.msgIsErr = false
}

func (m *Model) submitAddInput() {
	value := cleanAddInput(m.addInputValue())
	if value == "" {
		m.message = "输入不能为空"
		m.msgIsErr = true
		return
	}

	switch m.addStep {
	case addStepName:
		if _, exists := m.cfg.ProfileMap[value]; exists {
			m.message = fmt.Sprintf("配置「%s」已存在", value)
			m.msgIsErr = true
			return
		}
		m.addName = value
		m.addStep = addStepURL
	case addStepURL:
		m.addURL = value
		m.addStep = addStepToken
	case addStepToken:
		m.addToken = value
		m.finishAdd()
		return
	}

	m.resetAddInput()
	m.message = ""
	m.msgIsErr = false
}

func cleanAddInput(input string) string {
	return strings.TrimSpace(string(cleanInputRuneList([]rune(input))))
}

func cleanInputRuneList(inputRuneList []rune) []rune {
	cleanRuneList := make([]rune, 0, len(inputRuneList))
	for _, inputRune := range inputRuneList {
		if unicode.IsControl(inputRune) {
			continue
		}
		cleanRuneList = append(cleanRuneList, inputRune)
	}
	return cleanRuneList
}

func (m *Model) addInputValue() string {
	return string(m.addInputRuneList)
}

func (m *Model) resetAddInput() {
	m.addInputRuneList = nil
	m.addCursor = 0
}

func (m *Model) insertAddInputRuneList(inputRuneList []rune) {
	cleanRuneList := cleanInputRuneList(inputRuneList)
	if len(cleanRuneList) == 0 {
		return
	}
	m.clampAddCursor()

	nextRuneList := make([]rune, 0, len(m.addInputRuneList)+len(cleanRuneList))
	nextRuneList = append(nextRuneList, m.addInputRuneList[:m.addCursor]...)
	nextRuneList = append(nextRuneList, cleanRuneList...)
	nextRuneList = append(nextRuneList, m.addInputRuneList[m.addCursor:]...)
	m.addInputRuneList = nextRuneList
	m.addCursor += len(cleanRuneList)
	m.message = ""
	m.msgIsErr = false
}

func (m *Model) deleteAddInputBeforeCursor() {
	m.clampAddCursor()
	if m.addCursor == 0 {
		return
	}
	m.addInputRuneList = append(m.addInputRuneList[:m.addCursor-1], m.addInputRuneList[m.addCursor:]...)
	m.addCursor--
	m.message = ""
	m.msgIsErr = false
}

func (m *Model) deleteAddInputAtCursor() {
	m.clampAddCursor()
	if m.addCursor >= len(m.addInputRuneList) {
		return
	}
	m.addInputRuneList = append(m.addInputRuneList[:m.addCursor], m.addInputRuneList[m.addCursor+1:]...)
	m.message = ""
	m.msgIsErr = false
}

func (m *Model) deleteAddInputWordBeforeCursor() {
	m.clampAddCursor()
	if m.addCursor == 0 {
		return
	}

	start := m.addCursor
	for start > 0 && unicode.IsSpace(m.addInputRuneList[start-1]) {
		start--
	}
	for start > 0 && !unicode.IsSpace(m.addInputRuneList[start-1]) {
		start--
	}
	m.addInputRuneList = append(m.addInputRuneList[:start], m.addInputRuneList[m.addCursor:]...)
	m.addCursor = start
	m.message = ""
	m.msgIsErr = false
}

func (m *Model) clampAddCursor() {
	if m.addCursor < 0 {
		m.addCursor = 0
		return
	}
	if m.addCursor > len(m.addInputRuneList) {
		m.addCursor = len(m.addInputRuneList)
	}
}

func (m *Model) finishAdd() {
	if m.cfg == nil {
		m.cfg = &config.Config{ProfileMap: make(map[string]config.Profile)}
	}
	if m.cfg.ProfileMap == nil {
		m.cfg.ProfileMap = make(map[string]config.Profile)
	}

	m.cfg.ProfileMap[m.addName] = m.agent.BuildProfile(agent.ProfileInput{APIURL: m.addURL, Token: m.addToken})
	if err := m.agent.SaveConfig(m.cfg); err != nil {
		m.message = fmt.Sprintf("保存配置失败: %v", err)
		m.msgIsErr = true
		return
	}

	m.nameList = m.cfg.SortedNames()
	for i, name := range m.nameList {
		if name == m.addName {
			m.cursor = i
			break
		}
	}

	name := m.addName
	m.adding = false
	m.resetAddInput()
	m.message = fmt.Sprintf("已添加配置「%s」，按 Enter 切换", name)
	m.msgIsErr = false
}

// doSwitch 执行切换配置的操作
func (m *Model) doSwitch() {
	if len(m.nameList) == 0 {
		m.message = "暂无配置，请按 a 添加"
		m.msgIsErr = true
		return
	}

	name := m.nameList[m.cursor]
	if name == m.cfg.Active {
		m.message = fmt.Sprintf("「%s」已经是当前配置", name)
		m.msgIsErr = false
		return
	}

	profileMap := m.cfg.ProfileMap[name]
	if err := m.agent.ApplyProfile(name, profileMap); err != nil {
		m.message = fmt.Sprintf("写入设置失败: %v", err)
		m.msgIsErr = true
		return
	}

	previousActive := m.cfg.Active
	m.cfg.Active = name
	if err := m.agent.SaveConfig(m.cfg); err != nil {
		m.cfg.Active = previousActive
		m.message = fmt.Sprintf("保存配置失败: %v", err)
		m.msgIsErr = true
		return
	}

	m.message = fmt.Sprintf("✓ 已切换到「%s」", name)
	m.msgIsErr = false
}

func emptyConfigIfNotFound(err error) (*config.Config, error) {
	if err == nil {
		return nil, nil
	}
	if errors.Is(err, config.ErrConfigNotFound) {
		return &config.Config{ProfileMap: make(map[string]config.Profile)}, nil
	}
	return nil, err
}
