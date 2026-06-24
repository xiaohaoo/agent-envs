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

type profileFormStep int

const (
	profileFormStepName profileFormStep = iota
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
	editing          bool                // 是否在修改配置阶段
	deleting         bool                // 是否在删除确认阶段
	formStep         profileFormStep     // 当前表单步骤
	editName         string              // 正在修改的原配置名称
	deleteName       string              // 待删除配置名称
	addName          string              // 待添加配置名称
	formFieldList    []agent.ProfileField
	formValueMap     map[string]string
	addInputRuneList []rune // 当前输入框内容
	addCursor        int    // 添加表单输入光标位置
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
		if m.adding || m.editing {
			return m.updateAddForm(msg)
		}
		if m.deleting {
			return m.updateDeleteConfirm(msg)
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

		case "e":
			if !m.selecting {
				m.startEdit()
			}

		case "d":
			if !m.selecting {
				m.startDelete()
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

	if m.adding || m.editing {
		action := "添加"
		if m.editing {
			action = "修改"
		}
		return RenderProfileForm(action, m.agent.Name(), m.formStep, m.addName, m.formFieldList, m.formValueMap, m.addInputValue(), m.addCursor, m.message, m.msgIsErr)
	}

	if m.deleting {
		return RenderDeleteProfile(m.agent.Name(), m.deleteName, m.deleteName == m.cfg.Active, m.message, m.msgIsErr)
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
	m.editing = false
	m.deleting = false
	m.formStep = profileFormStepName
	m.editName = ""
	m.addName = ""
	m.formFieldList = nil
	if m.agent != nil {
		m.formFieldList = m.agent.ProfileFieldList()
	}
	m.formValueMap = make(map[string]string, len(m.formFieldList))
	m.resetAddInput()
	m.message = ""
	m.msgIsErr = false
}

func (m *Model) cancelAdd() {
	wasEditing := m.editing
	m.adding = false
	m.editing = false
	m.editName = ""
	m.resetAddInput()
	if wasEditing {
		m.message = "已取消修改配置"
	} else {
		m.message = "已取消添加配置"
	}
	m.msgIsErr = false
}

func (m *Model) submitAddInput() {
	value := cleanAddInput(m.addInputValue())
	if value == "" {
		m.message = "输入不能为空"
		m.msgIsErr = true
		return
	}

	switch m.formStep {
	case profileFormStepName:
		if _, exists := m.cfg.ProfileMap[value]; exists && (!m.editing || value != m.editName) {
			m.message = fmt.Sprintf("配置「%s」已存在", value)
			m.msgIsErr = true
			return
		}
		m.addName = value
	default:
		field, ok := m.currentProfileField()
		if !ok {
			m.message = "表单字段无效"
			m.msgIsErr = true
			return
		}
		m.formValueMap[field.Key] = value
	}

	if !m.advanceProfileFormStep() {
		if m.editing {
			m.finishEdit()
			return
		}
		m.finishAdd()
		return
	}
	m.setAddInput(m.currentProfileFormValue())
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

func (m *Model) setAddInput(value string) {
	m.addInputRuneList = cleanInputRuneList([]rune(value))
	m.addCursor = len(m.addInputRuneList)
}

func (m *Model) currentProfileField() (agent.ProfileField, bool) {
	fieldIndex := int(m.formStep) - 1
	if fieldIndex < 0 || fieldIndex >= len(m.formFieldList) {
		return agent.ProfileField{}, false
	}
	return m.formFieldList[fieldIndex], true
}

func (m *Model) currentProfileFormValue() string {
	if m.formStep == profileFormStepName {
		return m.addName
	}

	field, ok := m.currentProfileField()
	if !ok {
		return ""
	}
	return m.formValueMap[field.Key]
}

func (m *Model) advanceProfileFormStep() bool {
	if int(m.formStep) >= len(m.formFieldList) {
		return false
	}
	m.formStep++
	return true
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

	m.cfg.ProfileMap[m.addName] = m.agent.BuildProfile(agent.ProfileInput{FieldValueMap: m.formValueMap})
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

func (m *Model) startEdit() {
	if len(m.nameList) == 0 {
		m.message = "暂无配置可修改"
		m.msgIsErr = true
		return
	}

	m.clampProfileCursor()
	name := m.nameList[m.cursor]
	profileMap := m.cfg.ProfileMap[name]

	m.adding = false
	m.editing = true
	m.deleting = false
	m.formStep = profileFormStepName
	m.editName = name
	m.addName = name
	m.formFieldList = m.agent.ProfileFieldList()
	m.formValueMap = make(map[string]string, len(m.formFieldList))
	for _, field := range m.formFieldList {
		if value, ok := profileMap.String(field.Key); ok {
			m.formValueMap[field.Key] = value
		}
	}
	m.setAddInput(name)
	m.message = ""
	m.msgIsErr = false
}

func (m *Model) finishEdit() {
	if m.cfg == nil {
		m.cfg = &config.Config{ProfileMap: make(map[string]config.Profile)}
	}
	if m.cfg.ProfileMap == nil {
		m.cfg.ProfileMap = make(map[string]config.Profile)
	}

	oldName := m.editName
	oldProfileMap := m.cfg.ProfileMap[oldName]
	oldActive := m.cfg.Active

	nextProfileMap := make(config.Profile, len(oldProfileMap)+4)
	for key, value := range oldProfileMap {
		nextProfileMap[key] = value
	}
	for key, value := range m.agent.BuildProfile(agent.ProfileInput{FieldValueMap: m.formValueMap}) {
		nextProfileMap[key] = value
	}

	if m.addName != oldName {
		delete(m.cfg.ProfileMap, oldName)
	}
	m.cfg.ProfileMap[m.addName] = nextProfileMap
	if oldActive == oldName {
		m.cfg.Active = m.addName
	}

	if oldActive == oldName {
		if err := m.agent.ApplyProfile(m.addName, nextProfileMap); err != nil {
			if m.addName != oldName {
				delete(m.cfg.ProfileMap, m.addName)
			}
			m.cfg.ProfileMap[oldName] = oldProfileMap
			m.cfg.Active = oldActive
			m.message = fmt.Sprintf("写入设置失败: %v", err)
			m.msgIsErr = true
			return
		}
	}

	if err := m.agent.SaveConfig(m.cfg); err != nil {
		if m.addName != oldName {
			delete(m.cfg.ProfileMap, m.addName)
		}
		m.cfg.ProfileMap[oldName] = oldProfileMap
		m.cfg.Active = oldActive
		m.message = fmt.Sprintf("保存配置失败: %v", err)
		m.msgIsErr = true
		return
	}

	name := m.addName
	m.nameList = m.cfg.SortedNames()
	for i, currentName := range m.nameList {
		if currentName == name {
			m.cursor = i
			break
		}
	}

	m.editing = false
	m.editName = ""
	m.resetAddInput()
	m.message = fmt.Sprintf("已修改配置「%s」", name)
	m.msgIsErr = false
}

func (m *Model) startDelete() {
	if len(m.nameList) == 0 {
		m.message = "暂无配置可删除"
		m.msgIsErr = true
		return
	}

	m.clampProfileCursor()
	m.adding = false
	m.editing = false
	m.deleting = true
	m.deleteName = m.nameList[m.cursor]
	m.message = ""
	m.msgIsErr = false
}

func (m Model) updateDeleteConfirm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "q":
		m.quitting = true
		return m, tea.Quit
	case "esc", "n":
		m.deleting = false
		m.deleteName = ""
		m.message = "已取消删除配置"
		m.msgIsErr = false
		return m, nil
	case "y":
		m.finishDelete()
		return m, nil
	}

	return m, nil
}

func (m *Model) finishDelete() {
	if m.cfg == nil || m.cfg.ProfileMap == nil {
		m.deleting = false
		m.message = "暂无配置可删除"
		m.msgIsErr = true
		return
	}

	name := m.deleteName
	oldProfileMap := m.cfg.ProfileMap[name]
	if oldProfileMap == nil {
		m.deleting = false
		m.deleteName = ""
		m.message = fmt.Sprintf("配置「%s」不存在", name)
		m.msgIsErr = true
		return
	}

	oldActive := m.cfg.Active
	delete(m.cfg.ProfileMap, name)
	if m.cfg.Active == name {
		m.cfg.Active = ""
	}

	if err := m.agent.SaveConfig(m.cfg); err != nil {
		m.cfg.ProfileMap[name] = oldProfileMap
		m.cfg.Active = oldActive
		m.deleting = false
		m.deleteName = ""
		m.message = fmt.Sprintf("保存配置失败: %v", err)
		m.msgIsErr = true
		return
	}

	m.nameList = m.cfg.SortedNames()
	if m.cursor >= len(m.nameList) {
		m.cursor = len(m.nameList) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	m.deleting = false
	m.deleteName = ""
	m.message = fmt.Sprintf("已删除配置「%s」", name)
	m.msgIsErr = false
}

func (m *Model) clampProfileCursor() {
	if m.cursor < 0 {
		m.cursor = 0
		return
	}
	if m.cursor >= len(m.nameList) {
		m.cursor = len(m.nameList) - 1
	}
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
