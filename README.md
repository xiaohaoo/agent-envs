# Agent Envs Switcher

一个优雅的 TUI 工具，用于快速切换 Claude Code 和 Codex 的环境配置。

## 功能特性

- 🚀 支持 Claude Code 和 Codex 两种 AI 代理工具
- 🎨 美观的终端用户界面（基于 Bubble Tea）
- ⚡ 快速切换不同的 API 配置
- 🔒 自动管理认证信息
- 📝 配置文件自动同步
- 🌍 支持 macOS / Linux / Windows 多平台

## 安装

### 从 GitHub Release 下载

前往 [Releases](https://github.com/你的用户名/agent-envs/releases) 页面下载对应平台的二进制文件：

| 平台 | 架构 | 文件名 |
| ---- | ---- | ------ |
| macOS (Apple Silicon) | arm64 | `agent-envs-darwin-arm64.tar.gz` |
| macOS (Intel) | amd64 | `agent-envs-darwin-amd64.tar.gz` |
| Linux | amd64 | `agent-envs-linux-amd64.tar.gz` |
| Linux | arm64 | `agent-envs-linux-arm64.tar.gz` |
| Windows | amd64 | `agent-envs-windows-amd64.tar.gz` |
| Windows | arm64 | `agent-envs-windows-arm64.tar.gz` |

```bash
# 下载并解压（以 macOS arm64 为例）
tar -xzf agent-envs-darwin-arm64.tar.gz

# 移动到系统路径
sudo mv agent-envs-darwin-arm64/agent-envs /usr/local/bin/
```

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/你的用户名/agent-envs.git
cd agent-envs

# 编译当前平台
make build

# 或编译全部平台
make release

# 可选：移动到系统路径
sudo mv agent-envs /usr/local/bin/
```

### 依赖

- Go 1.24 或更高版本
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI 框架
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - 终端样式库
- [TOML](https://github.com/BurntSushi/toml) - TOML 解析库

## 配置文件

### Claude Code 配置

配置文件位置：`~/.claude/agent-envs.toml`

```toml
active = "MMKG中转站"

["MMKG中转站"]
ANTHROPIC_AUTH_TOKEN = "sk-xxx..."
ANTHROPIC_BASE_URL = "https://code.mmkg.cloud"

["MiniMax"]
ANTHROPIC_AUTH_TOKEN = "sk-xxx..."
ANTHROPIC_BASE_URL = "https://api.minimaxi.com/anthropic"
```

### Codex 配置

配置文件位置：`~/.codex/agent-envs.toml`

```toml
active = "GMN中转站"

["GMN中转站"]
name = "codex"
base_url = "https://gmn.chuangzuoli.com"
wire_api = "responses"
requires_openai_auth = "true"
OPENAI_API_KEY = "sk-xxx..."
model_provider = "codex"
model = "gpt-5.3-codex"
model_reasoning_effort = "medium"

["MMKG中转站"]
name = "codex"
base_url = "https://code.mmkg.cloud"
wire_api = "responses"
requires_openai_auth = "true"
OPENAI_API_KEY = "sk-xxx..."
model_provider = "codex"
model = "gpt-5.3-codex"
model_reasoning_effort = "medium"
```

## 使用方法

### 启动程序

```bash
agent-envs
```

### 查看版本

```bash
agent-envs --version
```

### 键盘操作

#### 选择代理类型界面

- `↑/↓` 或 `k/j` - 移动光标
- `Enter` 或 `空格` - 选择代理类型
- `Esc` 或 `q` - 退出程序
- `Ctrl+C` - 退出程序

#### 配置列表界面

- `↑/↓` 或 `k/j` - 移动光标
- `Enter` 或 `空格` - 切换到选中的配置
- `Esc` - 返回到选择代理类型界面
- `q` - 退出程序
- `Ctrl+C` - 退出程序

## 工作原理

### Claude Code

切换配置时，程序会：
1. 读取 `~/.claude/agent-envs.toml` 中的配置
2. 将选中的配置写入 `~/.claude/settings.json` 的 `env` 字段
3. 更新 `active` 字段到配置文件

### Codex

切换配置时，程序会：
1. 读取 `~/.codex/agent-envs.toml` 中的配置
2. 将配置信息写入 `~/.codex/config.toml`
3. 将认证信息写入 `~/.codex/auth.json`（权限 600）
4. 更新 `active` 字段到配置文件

## 界面预览

```
⚡ 选择代理类型

▸ Claude Code (Anthropic Claude Code)
  Codex (Codex CLI)

↑/↓ 移动  •  Enter 选择  •  Esc/q 退出
```

```
⚡ Claude Code Envs

▸ ● MMKG中转站
    URL: https://code.mmkg.cloud
    Key: sk-980a0****eb01
    ───────────────────────────────────
    MiniMax
    URL: https://api.minimaxi.com/anthropic
    Key: sk-cp-S****A3K2

↑/↓ 移动  •  Enter 切换  •  Esc 返回  •  q 退出
```

## 配色方案

使用 One Dark 主题的明亮配色：
- 主色调：亮蓝 (#61AFEF)
- 强调色：青色 (#56B6C2)
- 成功色：绿色 (#98C379)
- 错误色：红色 (#E06C75)

## 开发

### 项目结构

```
agent-envs/
├── cmd/agent-envs/main.go          # 程序入口
├── internal/
│   ├── config/
│   │   ├── errors.go               # 错误定义
│   │   ├── profile.go              # Profile 类型和方法
│   │   ├── paths.go                # 路径管理器
│   │   ├── config.go               # 配置加载/保存
│   │   └── config_test.go          # 单元测试
│   ├── agent/
│   │   ├── agent.go                # Agent 接口定义
│   │   ├── claude.go               # Claude Code 实现
│   │   ├── codex.go                # Codex CLI 实现
│   │   └── agent_test.go           # 单元测试
│   └── ui/
│       ├── styles.go               # 样式定义
│       ├── model.go                # Bubble Tea 模型
│       ├── view_selector.go        # 代理选择视图
│       ├── view_profiles.go        # 配置列表视图
│       └── ui_test.go              # 单元测试
├── .github/workflows/release.yml   # 自动发布工作流
├── Makefile                         # 多平台编译
└── README.md
```

### 架构设计

项目采用三层架构，依赖方向为 `config → agent → ui`：

- **config 包** — 最独立的包，负责配置文件的解析、保存和路径管理
- **agent 包** — 依赖 config，定义 `Agent` 接口并提供 Claude/Codex 实现
- **ui 包** — 依赖 agent 和 config，实现 Bubble Tea TUI 界面

### 常用命令

```bash
# 编译当前平台
make build

# 运行测试
make test

# 编译全部 6 个平台
make release

# 测试 + 编译 + 生成 checksums
make all

# 清理构建产物
make clean
```

### 发布新版本

```bash
# 打 tag 并推送，GitHub Actions 自动构建发布
git tag v1.0.0
git push origin v1.0.0
```

## 注意事项

1. **终端环境**：本程序是 TUI 应用，需要在真正的终端中运行（不支持 IDE 内置运行器）
2. **配置文件格式**：确保 TOML 文件格式正确
3. **权限问题**：Codex 的 `auth.json` 文件权限为 600
4. **路径问题**：配置文件必须在对应的主目录下
5. **备份配置**：修改配置前建议备份原有配置文件

## 故障排除

### 配置文件不存在

```bash
# 创建 Claude Code 配置目录
mkdir -p ~/.claude

# 创建 Codex 配置目录
mkdir -p ~/.codex
```

### 权限问题

```bash
# 修复 Codex auth.json 权限
chmod 600 ~/.codex/auth.json
```

### 编译错误

```bash
# 安装依赖
go mod tidy

# 重新编译
make build
```

### IDE 中运行报错 "open /dev/tty: device not configured"

这是正常现象。TUI 程序需要真正的终端环境，请在系统终端或 IDE 的 Terminal 面板中运行。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 作者

Created with ❤️ by Kiro (Claude Opus 4.6)
