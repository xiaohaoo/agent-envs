# Agent Envs

[简体中文](README.zh-CN.md)

`agent-envs` is a terminal UI for switching Claude Code and Codex CLI profiles. It stores reusable profile definitions in TOML files, lets you pick one interactively, and writes the selected values into each tool's native configuration files.

## Features

- Support both Claude Code and Codex CLI
- Fast profile switching from a Bubble Tea based TUI
- Preserve unrelated settings in existing config files when applying a profile
- Keep multiple providers in one place instead of editing config files by hand
- Cross-platform support for macOS, Linux, and Windows
- Built-in version output with `--version`

## Installation

### Download from GitHub Releases

Download the archive for your platform from the [Releases](https://github.com/xiaohaoo/agent-envs/releases) page:

| Platform | Architecture | Filename |
| -------- | ------------ | -------- |
| macOS (Apple Silicon) | arm64 | `agent-envs-darwin-arm64.tar.gz` |
| macOS (Intel) | amd64 | `agent-envs-darwin-amd64.tar.gz` |
| Linux | amd64 | `agent-envs-linux-amd64.tar.gz` |
| Linux | arm64 | `agent-envs-linux-arm64.tar.gz` |
| Windows | amd64 | `agent-envs-windows-amd64.tar.gz` |
| Windows | arm64 | `agent-envs-windows-arm64.tar.gz` |

```bash
# Extract the archive (macOS arm64 example)
tar -xzf agent-envs-darwin-arm64.tar.gz

# Move the binary into your PATH
sudo mv agent-envs-darwin-arm64/agent-envs /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/xiaohaoo/agent-envs.git
cd agent-envs

# Build the current platform binary
make build

# Optional: verify the build
./agent-envs --version
```

### Install the Local Build

```bash
make install
```

### Build Requirements

- Go 1.24 or later

## Quick Start

1. Create one or both profile files:
   - `~/.claude/agent-envs.toml`
   - `~/.codex/agent-envs.toml`
2. Run `agent-envs`
3. Choose `Claude Code` or `Codex`
4. Select the profile you want to activate

## Configuration

### Claude Code

Profile file: `~/.claude/agent-envs.toml`

```toml
active = "Primary Provider"

["Primary Provider"]
ANTHROPIC_AUTH_TOKEN = "sk-ant-..."
ANTHROPIC_BASE_URL = "https://api.example.com"

["Backup Provider"]
ANTHROPIC_AUTH_TOKEN = "sk-ant-..."
ANTHROPIC_BASE_URL = "https://api.backup.example.com"
```

`agent-envs` merges the selected profile into the `env` field of `~/.claude/settings.json` and preserves unrelated existing environment variables.

`~/.claude/settings.json` must already exist. If Claude Code has never been launched on the machine, create the file first:

```json
{
  "env": {}
}
```

Claude profiles are treated as raw environment variables:

- Every key in the selected profile is merged into `~/.claude/settings.json`
- Existing `env` keys that are not present in the selected profile are left untouched
- `agent-envs` does not create `~/.claude/settings.json` or the `~/.claude` directory for you

### Codex CLI

Profile file: `~/.codex/agent-envs.toml`

```toml
active = "Primary Provider"

["Primary Provider"]
base_url = "https://api.example.com"
wire_api = "responses"
requires_openai_auth = true
OPENAI_API_KEY = "sk-..."

["Backup Provider"]
base_url = "https://api.backup.example.com"
wire_api = "responses"
requires_openai_auth = true
OPENAI_API_KEY = "sk-..."
```

`agent-envs` currently applies these Codex profile keys:

- `base_url` -> `[model_providers."<profile name>"].base_url`
- `wire_api` -> `[model_providers."<profile name>"].wire_api`
- `requires_openai_auth` -> `[model_providers."<profile name>"].requires_openai_auth`
- `OPENAI_API_KEY` -> `~/.codex/auth.json`

The provider `name` field is always written from the profile name. Extra keys in `~/.codex/agent-envs.toml` stay in that file, but are not written into native Codex config files.

When switching Codex profiles, the profile name becomes the active `model_provider`. `agent-envs` also:

- Preserves unrelated top-level settings in `~/.codex/config.toml`
- Preserves other provider entries under `model_providers`
- Rewrites the selected `[model_providers."<profile name>"]` table with the managed fields above
- Updates `OPENAI_API_KEY` in `~/.codex/auth.json` only when the selected profile provides one

If the `~/.codex` directory already exists, `config.toml` and `auth.json` can be created on the first successful switch.

## Usage

### Launch the Program

```bash
agent-envs
```

### Show Version

```bash
agent-envs --version
```

### Keyboard Controls

#### Agent Selection Screen

- `↑/↓` or `k/j` to move
- `Enter` or `Space` to select
- `Esc`, `q`, or `Ctrl+C` to quit

#### Profile List Screen

- `↑/↓` or `k/j` to move
- `Enter` or `Space` to switch to the selected profile
- `Esc` to return to agent selection
- `q` or `Ctrl+C` to quit

## What Changes on Switch

### Claude Code

1. Read `~/.claude/agent-envs.toml`
2. Merge the selected profile into the `env` field of `~/.claude/settings.json`
3. Leave existing `env` keys in place if the new profile does not define them
4. Update `active` in `~/.claude/agent-envs.toml`

### Codex CLI

1. Read `~/.codex/agent-envs.toml`
2. Update top-level `model_provider` in `~/.codex/config.toml`
3. Replace or create the selected `[model_providers."<profile name>"]` table using `name`, `base_url`, `wire_api`, and optional `requires_openai_auth`
4. Preserve unrelated top-level Codex settings and other provider tables
5. Merge `OPENAI_API_KEY` into `~/.codex/auth.json` only when the selected profile contains it
6. Preserve other existing auth fields in `~/.codex/auth.json`
7. Write `~/.codex/auth.json` with `0600` permissions
8. Update `active` in `~/.codex/agent-envs.toml`

## UI Preview

The current UI text is Simplified Chinese:

```text
⚡ 选择代理类型

▸ Claude Code (Anthropic Claude Code)
  Codex (Codex CLI)

↑/↓ 移动  •  Enter 选择  •  Esc/q 退出
```

```text
⚡ Claude Code Envs

▸ ● Primary Provider
    URL: https://api.example.com
    Key: sk-ant-****abcd
    ───────────────────────────────────
    Backup Provider
    URL: https://api.backup.example.com
    Key: sk-ant-****wxyz

↑/↓ 移动  •  Enter 切换  •  Esc 返回  •  q 退出
```

## Development

### Project Structure

```text
agent-envs/
├── cmd/agent-envs/main.go          # Program entry point
├── internal/
│   ├── agent/
│   │   ├── agent.go                # Agent interface and factory
│   │   ├── claude.go               # Claude Code implementation
│   │   └── codex.go                # Codex CLI implementation
│   ├── config/
│   │   ├── config.go               # Config loading and saving
│   │   ├── errors.go               # Config-related errors
│   │   ├── keys.go                 # Shared config keys
│   │   ├── paths.go                # Path manager
│   │   └── profile.go              # Profile helpers
│   ├── fileutil/
│   │   ├── atomic.go               # Atomic file writes
│   │   └── json.go                 # JSON helpers
│   └── ui/
│       ├── model.go                # Bubble Tea update/view flow
│       ├── styles.go               # UI styles
│       ├── view_profiles.go        # Profile list rendering
│       └── view_selector.go        # Agent selector rendering
├── .github/workflows/release.yml   # Release workflow
├── Makefile
├── README.md
└── README.zh-CN.md
```

### Architecture

The project follows a three-layer dependency flow: `config -> agent -> ui`

- `config` handles parsing, serialization, and file paths
- `agent` applies profile changes for Claude Code and Codex
- `ui` renders the terminal interface and handles interaction

### Common Commands

```bash
# Build for the current platform
make build

# Build and run locally
make run

# Install the built binary
make install

# Run tests
make test

# Run go vet
make vet

# Format code
make fmt

# Run golangci-lint if installed
make lint

# Build archives for all supported platforms
make release

# Test + release + checksums
make all

# Clean build artifacts
make clean
```

### Release a New Version

```bash
git tag v1.0.0
git push origin v1.0.0
```

Pushing a `v*` tag triggers the GitHub Actions release workflow.

## Troubleshooting

### Configuration File Not Found

Create the required config directories first:

```bash
mkdir -p ~/.claude ~/.codex
```

Then create the profile files using the examples above.

### `active` Points to a Missing Profile

Make sure `active = "..."` matches one of the section names in `agent-envs.toml`. The program refuses to load a config when the active profile does not exist.

### Claude Code `settings.json` Is Missing

Create `~/.claude/settings.json` before switching Claude profiles:

```json
{
  "env": {}
}
```

### Permission Issues

`~/.codex/auth.json` is written with `0600` permissions. If you need to fix it manually:

```bash
chmod 600 ~/.codex/auth.json
```

`~/.codex/config.toml` and `~/.codex/auth.json` can be created automatically on the first switch, but the parent `~/.codex` directory must already exist.

### Build Errors

```bash
go mod tidy
make build
```

### IDE Error: `open /dev/tty: device not configured`

This is expected for TUI programs started from non-terminal runners. Run `agent-envs` in a real terminal or in your IDE's integrated terminal.

## License

MIT License

## Contributing

Issues and pull requests are welcome.

## Author

[xiaohaoo](https://github.com/xiaohaoo)
