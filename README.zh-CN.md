# Agent Envs

[English](README.md)

`agent-envs` 是一个用于切换 Claude Code 和 Codex CLI 配置档的终端界面工具。你可以把常用服务商配置集中写在 TOML 文件里，通过交互式界面选择目标配置，然后让程序把对应值写入各自原生配置文件。

## 功能特性

- 同时支持 Claude Code 和 Codex CLI
- 基于 Bubble Tea 的终端交互界面，切换速度快
- 应用配置时保留原有无关设置，不会粗暴覆盖整个文件
- 把多个服务商配置集中管理，减少手动改配置的成本
- 支持 macOS、Linux、Windows
- 内置 `--version` 版本输出

## 安装

### 从 GitHub Releases 下载

前往 [Releases](https://github.com/xiaohaoo/agent-envs/releases) 页面，下载适合你平台的压缩包：

| 平台 | 架构 | 文件名 |
| ---- | ---- | ------ |
| macOS (Apple Silicon) | arm64 | `agent-envs-darwin-arm64.tar.gz` |
| macOS (Intel) | amd64 | `agent-envs-darwin-amd64.tar.gz` |
| Linux | amd64 | `agent-envs-linux-amd64.tar.gz` |
| Linux | arm64 | `agent-envs-linux-arm64.tar.gz` |
| Windows | amd64 | `agent-envs-windows-amd64.tar.gz` |
| Windows | arm64 | `agent-envs-windows-arm64.tar.gz` |

```bash
# 下载后解压（以 macOS arm64 为例）
tar -xzf agent-envs-darwin-arm64.tar.gz

# 移动到 PATH
sudo mv agent-envs-darwin-arm64/agent-envs /usr/local/bin/
```

### 从源码编译

```bash
git clone https://github.com/xiaohaoo/agent-envs.git
cd agent-envs

# 编译当前平台
make build

# 可选：确认版本输出
./agent-envs --version
```

### 安装本地编译结果

```bash
make install
```

### 构建要求

- Go 1.24 或更高版本

## 快速开始

1. 创建一个或两个配置文件：
   - `~/.claude/agent-envs.toml`
   - `~/.codex/agent-envs.toml`
2. 运行 `agent-envs`
3. 选择 `Claude Code` 或 `Codex`
4. 选中想要启用的配置档

## 配置说明

### Claude Code

配置文件：`~/.claude/agent-envs.toml`

```toml
active = "主服务商"

["主服务商"]
ANTHROPIC_AUTH_TOKEN = "sk-ant-..."
ANTHROPIC_BASE_URL = "https://api.example.com"

["备用服务商"]
ANTHROPIC_AUTH_TOKEN = "sk-ant-..."
ANTHROPIC_BASE_URL = "https://api.backup.example.com"
```

切换 Claude Code 配置时，`agent-envs` 会把所选配置合并到 `~/.claude/settings.json` 的 `env` 字段，并保留其中原有的其他环境变量。

`~/.claude/settings.json` 需要事先存在。如果这台机器上还没运行过 Claude Code，可以先手动创建一个最小文件：

```json
{
  "env": {}
}
```

Claude 配置档会被当作原始环境变量处理：

- 选中配置中的每个键都会合并进 `~/.claude/settings.json` 的 `env`
- 新配置里没有声明的已有 `env` 键不会被删除
- `agent-envs` 不会替你创建 `~/.claude/settings.json` 或 `~/.claude` 目录

### Codex CLI

配置文件：`~/.codex/agent-envs.toml`

```toml
active = "主服务商"

["主服务商"]
base_url = "https://api.example.com"
wire_api = "responses"
requires_openai_auth = true
OPENAI_API_KEY = "sk-..."

["备用服务商"]
base_url = "https://api.backup.example.com"
wire_api = "responses"
requires_openai_auth = true
OPENAI_API_KEY = "sk-..."
```

`agent-envs` 当前会应用这些 Codex 配置键：

- `base_url` -> `[model_providers."<配置名>"].base_url`
- `wire_api` -> `[model_providers."<配置名>"].wire_api`
- `requires_openai_auth` -> `[model_providers."<配置名>"].requires_openai_auth`
- `OPENAI_API_KEY` -> `~/.codex/auth.json`

provider 的 `name` 字段会始终由配置档名称自动写入。`~/.codex/agent-envs.toml` 里额外存在的键会保留在这个文件中，但不会被写进 Codex 原生配置文件。

切换 Codex 配置时，所选配置档名称会成为当前 `model_provider`。此外，`agent-envs` 还会：

- 保留 `~/.codex/config.toml` 中无关的顶层设置
- 保留 `model_providers` 下其他 provider 的配置
- 用上面这些受支持字段重写当前选中的 `[model_providers."<配置名>"]`
- 仅在当前配置档提供了 `OPENAI_API_KEY` 时，才更新 `~/.codex/auth.json` 中的该字段

如果 `~/.codex` 目录已经存在，那么 `config.toml` 和 `auth.json` 可以在第一次成功切换时自动创建。

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

#### 代理选择界面

- `↑/↓` 或 `k/j`：移动
- `Enter` 或 `Space`：选择
- `Esc`、`q` 或 `Ctrl+C`：退出

#### 配置列表界面

- `↑/↓` 或 `k/j`：移动
- `Enter` 或 `Space`：切换到当前选中的配置
- `Esc`：返回代理选择界面
- `q` 或 `Ctrl+C`：退出

## 切换时会改哪些文件

### Claude Code

1. 读取 `~/.claude/agent-envs.toml`
2. 将所选配置合并到 `~/.claude/settings.json` 的 `env` 字段
3. 新配置未声明的已有 `env` 键保持不变
4. 更新 `~/.claude/agent-envs.toml` 中的 `active`

### Codex CLI

1. 读取 `~/.codex/agent-envs.toml`
2. 更新 `~/.codex/config.toml` 顶层的 `model_provider`
3. 用 `name`、`base_url`、`wire_api` 以及可选的 `requires_openai_auth` 替换或创建当前选中的 `[model_providers."<配置名>"]`
4. 保留其他无关的 Codex 顶层设置和 provider 配置
5. 仅在当前配置档包含 `OPENAI_API_KEY` 时，才将其合并写入 `~/.codex/auth.json`
6. 保留 `~/.codex/auth.json` 中其他已有认证字段
7. 以 `0600` 权限写入 `~/.codex/auth.json`
8. 更新 `~/.codex/agent-envs.toml` 中的 `active`

## 界面预览

当前界面文案为简体中文：

```text
⚡ 选择代理类型

▸ Claude Code (Anthropic Claude Code)
  Codex (Codex CLI)

↑/↓ 移动  •  Enter 选择  •  Esc/q 退出
```

```text
⚡ Claude Code Envs

▸ ● 主服务商
    URL: https://api.example.com
    Key: sk-ant-****abcd
    ───────────────────────────────────
    备用服务商
    URL: https://api.backup.example.com
    Key: sk-ant-****wxyz

↑/↓ 移动  •  Enter 切换  •  Esc 返回  •  q 退出
```

## 开发

### 项目结构

```text
agent-envs/
├── cmd/agent-envs/main.go          # 程序入口
├── internal/
│   ├── agent/
│   │   ├── agent.go                # Agent 接口与工厂
│   │   ├── claude.go               # Claude Code 实现
│   │   └── codex.go                # Codex CLI 实现
│   ├── config/
│   │   ├── config.go               # 配置加载与保存
│   │   ├── errors.go               # 配置相关错误
│   │   ├── keys.go                 # 共用配置键
│   │   ├── paths.go                # 路径管理
│   │   └── profile.go              # Profile 辅助方法
│   ├── fileutil/
│   │   ├── atomic.go               # 原子写文件
│   │   └── json.go                 # JSON 辅助方法
│   └── ui/
│       ├── model.go                # Bubble Tea 更新与视图流程
│       ├── styles.go               # UI 样式
│       ├── view_profiles.go        # 配置列表渲染
│       └── view_selector.go        # 代理选择渲染
├── .github/workflows/release.yml   # 发布工作流
├── Makefile
├── README.md
└── README.zh-CN.md
```

### 架构设计

项目采用三层依赖结构：`config -> agent -> ui`

- `config` 负责解析、序列化和路径处理
- `agent` 负责把配置应用到 Claude Code 和 Codex
- `ui` 负责终端界面渲染和交互

### 常用命令

```bash
# 编译当前平台
make build

# 本地编译并运行
make run

# 安装编译结果
make install

# 运行测试
make test

# 运行 go vet
make vet

# 格式化代码
make fmt

# 运行 golangci-lint（若已安装）
make lint

# 构建所有支持平台的发布包
make release

# 测试 + 发布包 + checksums
make all

# 清理构建产物
make clean
```

### 发布新版本

```bash
git tag v1.0.0
git push origin v1.0.0
```

推送 `v*` 标签后会自动触发 GitHub Actions 发布流程。

## 故障排除

### 找不到配置文件

先创建配置目录：

```bash
mkdir -p ~/.claude ~/.codex
```

然后按上面的示例创建对应的配置文件。

### `active` 指向了不存在的配置档

请确认 `active = "..."` 指向的名称和 `agent-envs.toml` 里的某个 section 完全一致。如果 `active` 对应的配置档不存在，程序会拒绝加载。

### Claude Code 的 `settings.json` 不存在

切换 Claude 配置前，请先创建 `~/.claude/settings.json`：

```json
{
  "env": {}
}
```

### 权限问题

`~/.codex/auth.json` 会以 `0600` 权限写入。如果需要手动修复：

```bash
chmod 600 ~/.codex/auth.json
```

`~/.codex/config.toml` 和 `~/.codex/auth.json` 可以在第一次切换时自动创建，但前提是上层 `~/.codex` 目录已经存在。

### 编译错误

```bash
go mod tidy
make build
```

### IDE 中出现 `open /dev/tty: device not configured`

这是 TUI 程序在非终端运行器中启动时的正常现象。请在系统终端或 IDE 的 Terminal 面板中运行 `agent-envs`。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request。

## 作者

[xiaohaoo](https://github.com/xiaohaoo)
