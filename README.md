# Agent Envs

[中文文档](README.zh-CN.md)

An elegant TUI tool for quickly switching between Claude Code and Codex environment configurations.

## Features

- 🚀 Support for both Claude Code and Codex AI agent tools
- 🎨 Beautiful terminal user interface (powered by Bubble Tea)
- ⚡ Fast switching between different API configurations
- 🔒 Automatic authentication management
- 📝 Automatic configuration file synchronization
- 🌍 Cross-platform support: macOS / Linux / Windows

## Installation

### Download from GitHub Releases

Visit the [Releases](https://github.com/xiaohaoo/agent-envs/releases) page to download the binary for your platform:

| Platform | Architecture | Filename |
| -------- | ------------ | -------- |
| macOS (Apple Silicon) | arm64 | `agent-envs-darwin-arm64.tar.gz` |
| macOS (Intel) | amd64 | `agent-envs-darwin-amd64.tar.gz` |
| Linux | amd64 | `agent-envs-linux-amd64.tar.gz` |
| Linux | arm64 | `agent-envs-linux-arm64.tar.gz` |
| Windows | amd64 | `agent-envs-windows-amd64.tar.gz` |
| Windows | arm64 | `agent-envs-windows-arm64.tar.gz` |

```bash
# Download and extract (macOS arm64 example)
tar -xzf agent-envs-darwin-arm64.tar.gz

# Move to system path
sudo mv agent-envs-darwin-arm64/agent-envs /usr/local/bin/
```

### Build from Source

```bash
# Clone the repository
git clone https://github.com/xiaohaoo/agent-envs.git
cd agent-envs

# Build for current platform
make build

# Or build for all platforms
make release

# Optional: Install to system path
sudo mv agent-envs /usr/local/bin/
```

### Requirements

- Go 1.24 or higher
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Terminal styling library
- [TOML](https://github.com/BurntSushi/toml) - TOML parser

## Configuration

### Claude Code Configuration

Configuration file location: `~/.claude/agent-envs.toml`

```toml
active = "Primary Provider"

["Primary Provider"]
ANTHROPIC_AUTH_TOKEN = "sk-xxx..."
ANTHROPIC_BASE_URL = "https://api.example.com"

["Secondary Provider"]
ANTHROPIC_AUTH_TOKEN = "sk-xxx..."
ANTHROPIC_BASE_URL = "https://api.another.com"
```

### Codex Configuration

Configuration file location: `~/.codex/agent-envs.toml`

```toml
active = "Primary Provider"

["Primary Provider"]
name = "codex"
base_url = "https://api.example.com"
wire_api = "responses"
requires_openai_auth = "true"
OPENAI_API_KEY = "sk-xxx..."
model_provider = "codex"
model = "gpt-5.3-codex"

["Secondary Provider"]
name = "codex"
base_url = "https://api.another.com"
wire_api = "responses"
requires_openai_auth = "true"
OPENAI_API_KEY = "sk-xxx..."
model_provider = "codex"
model = "gpt-5.3-codex"
```

## Usage

### Launch the Program

```bash
agent-envs
```

### Check Version

```bash
agent-envs --version
```

### Keyboard Controls

#### Agent Type Selection Screen

- `↑/↓` or `k/j` - Move cursor
- `Enter` or `Space` - Select agent type
- `Esc` or `q` - Exit program
- `Ctrl+C` - Exit program

#### Configuration List Screen

- `↑/↓` or `k/j` - Move cursor
- `Enter` or `Space` - Switch to selected configuration
- `Esc` - Return to agent type selection
- `q` - Exit program
- `Ctrl+C` - Exit program

## How It Works

### Claude Code

When switching configurations, the program:
1. Reads configuration from `~/.claude/agent-envs.toml`
2. Writes the selected configuration to the `env` field in `~/.claude/settings.json`
3. Updates the `active` field in the configuration file

### Codex

When switching configurations, the program:
1. Reads configuration from `~/.codex/agent-envs.toml`
2. Writes configuration to `~/.codex/config.toml`
3. Writes authentication to `~/.codex/auth.json` (with 600 permissions)
4. Updates the `active` field in the configuration file

## UI Preview

```
⚡ Select Agent Type

▸ Claude Code (Anthropic Claude Code)
  Codex (Codex CLI)

↑/↓ Move  •  Enter Select  •  Esc/q Exit
```

```
⚡ Claude Code Envs

▸ ● Primary Provider
    URL: https://api.example.com
    Key: sk-********eb01
    ───────────────────────────────────
    Secondary Provider
    URL: https://api.another.com
    Key: sk-********A3K2

↑/↓ Move  •  Enter Switch  •  Esc Back  •  q Exit
```

## Color Scheme

Uses bright colors from the One Dark theme:
- Primary: Bright Blue (#61AFEF)
- Accent: Cyan (#56B6C2)
- Success: Green (#98C379)
- Error: Red (#E06C75)

## Development

### Project Structure

```
agent-envs/
├── cmd/agent-envs/main.go          # Program entry point
├── internal/
│   ├── config/
│   │   ├── errors.go               # Error definitions
│   │   ├── profile.go              # Profile type and methods
│   │   ├── paths.go                # Path manager
│   │   ├── keys.go                 # Configuration key constants
│   │   └── config.go               # Configuration loading/saving
│   ├── agent/
│   │   ├── agent.go                # Agent interface definition
│   │   ├── claude.go               # Claude Code implementation
│   │   └── codex.go                # Codex CLI implementation
│   ├── fileutil/
│   │   ├── atomic.go               # Atomic file write utilities
│   │   └── json.go                 # JSON serialization helpers
│   └── ui/
│       ├── styles.go               # Style definitions
│       ├── model.go                # Bubble Tea model
│       ├── view_selector.go        # Agent selection view
│       └── view_profiles.go        # Configuration list view
├── .github/workflows/release.yml   # Automated release workflow
├── Makefile                         # Multi-platform build
└── README.md
```

### Architecture

The project uses a three-layer architecture with dependency direction: `config → agent → ui`

- **config package** — Most independent package, handles configuration parsing, saving, and path management
- **agent package** — Depends on config, defines `Agent` interface and provides Claude/Codex implementations
- **ui package** — Depends on agent and config, implements Bubble Tea TUI interface

### Common Commands

```bash
# Build for current platform
make build

# Run tests
make test

# Build for all 6 platforms
make release

# Test + build + generate checksums
make all

# Clean build artifacts
make clean
```

### Release New Version

```bash
# Create tag and push, GitHub Actions will automatically build and release
git tag v1.0.0
git push origin v1.0.0
```

## Notes

1. **Terminal Environment**: This is a TUI application and must run in a real terminal (IDE built-in runners are not supported)
2. **Configuration Format**: Ensure TOML file format is correct
3. **Permissions**: Codex's `auth.json` file has 600 permissions
4. **Path Requirements**: Configuration files must be in the corresponding home directory
5. **Backup**: Recommend backing up existing configuration files before modifications

## Troubleshooting

### Configuration File Not Found

```bash
# Create Claude Code configuration directory
mkdir -p ~/.claude

# Create Codex configuration directory
mkdir -p ~/.codex
```

### Permission Issues

```bash
# Fix Codex auth.json permissions
chmod 600 ~/.codex/auth.json
```

### Build Errors

```bash
# Install dependencies
go mod tidy

# Rebuild
make build
```

### IDE Error: "open /dev/tty: device not configured"

This is expected behavior. TUI programs require a real terminal environment. Please run in a system terminal or IDE's Terminal panel.

## License

MIT License

## Contributing

Issues and Pull Requests are welcome!

## Author

[xiaohaoo](https://github.com/xiaohaoo)
