# Agent Envs - 多平台编译

APP_NAME := agent-envs
CMD_PATH := ./cmd/agent-envs
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

# 发布的目标平台
PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64

DIST_DIR := dist
GO := go
GOFLAGS := -trimpath

# 根据操作系统选择安装路径
ifeq ($(OS),Windows_NT)
	INSTALL_DIR := $(LOCALAPPDATA)\Programs\$(APP_NAME)
	BINARY := $(APP_NAME).exe
else
	INSTALL_DIR := /usr/local/bin
	BINARY := $(APP_NAME)
endif

.PHONY: help build run install clean test vet fmt lint release checksums all

# 默认目标：显示帮助
help:
	@echo "Agent Envs - 可用命令："
	@echo ""
	@echo "  make build      - 编译当前平台"
	@echo "  make run        - 编译并运行"
	@echo "  make install    - 安装到系统路径 (Unix: /usr/local/bin, Windows: %%LOCALAPPDATA%%\Programs)"
	@echo "  make test       - 运行测试"
	@echo "  make vet        - 运行 go vet"
	@echo "  make fmt        - 格式化代码"
	@echo "  make lint       - 运行 golangci-lint（如果已安装）"
	@echo "  make release    - 多平台编译"
	@echo "  make checksums  - 生成校验和"
	@echo "  make all        - 测试 + 编译 + 校验和"
	@echo "  make clean      - 清理构建产物"
	@echo ""
	@echo "Windows 用户也可使用: powershell -ExecutionPolicy Bypass -File install.ps1"
	@echo ""

# 编译当前平台
build:
	@echo "🔨 编译 $(APP_NAME) ($(VERSION))..."
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) $(CMD_PATH)
	@echo "✅ 编译完成: ./$(BINARY)"

# 编译并运行
run: build
	@echo "🚀 运行 $(APP_NAME)..."
	@./$(BINARY)

# 安装到系统路径
install: build
ifeq ($(OS),Windows_NT)
	@echo "📦 安装到 $(INSTALL_DIR)..."
	@if not exist "$(INSTALL_DIR)" mkdir "$(INSTALL_DIR)"
	@copy /Y $(BINARY) "$(INSTALL_DIR)\$(BINARY)"
	@echo "✅ 安装完成"
	@echo "⚠️  请确保 $(INSTALL_DIR) 已添加到 PATH 环境变量"
else
	@echo "📦 安装到 $(INSTALL_DIR)/$(APP_NAME)..."
	@sudo mv $(BINARY) $(INSTALL_DIR)/
	@echo "✅ 安装完成"
endif

# 运行测试
test:
	@echo "🧪 运行测试..."
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "✅ 测试通过"

# 运行 go vet
vet:
	@echo "🔍 运行 go vet..."
	@$(GO) vet ./...
	@echo "✅ 检查通过"

# 格式化代码
fmt:
	@echo "✨ 格式化代码..."
	@$(GO) fmt ./...
	@echo "✅ 格式化完成"

# 运行 linter（需要安装 golangci-lint）
lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "🔍 运行 golangci-lint..."; \
		golangci-lint run ./...; \
		echo "✅ Lint 检查通过"; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过"; \
		echo "   安装: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# 清理构建产物
clean:
	@echo "🧹 清理构建产物..."
ifeq ($(OS),Windows_NT)
	@if exist $(DIST_DIR) rmdir /S /Q $(DIST_DIR)
	@if exist $(BINARY) del /Q $(BINARY)
	@if exist coverage.out del /Q coverage.out
else
	@rm -rf $(DIST_DIR) $(BINARY) coverage.out
endif
	@echo "✅ 清理完成"

# 多平台编译
release: clean vet
	@echo "🚀 开始多平台编译..."
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS), \
		$(eval GOOS_VAL := $(word 1,$(subst /, ,$(platform)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(platform)))) \
		$(eval OUTPUT := $(APP_NAME)$(if $(filter windows,$(GOOS_VAL)),.exe,)) \
		echo "🔨 编译 $(GOOS_VAL)/$(ARCH)..."; \
		GOOS=$(GOOS_VAL) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
			-o $(DIST_DIR)/$(APP_NAME)-$(GOOS_VAL)-$(ARCH)/$(OUTPUT) $(CMD_PATH) || exit 1; \
		cd $(DIST_DIR) && tar -czf $(APP_NAME)-$(GOOS_VAL)-$(ARCH).tar.gz $(APP_NAME)-$(GOOS_VAL)-$(ARCH)/ && cd ..; \
		echo "  ✅ $(DIST_DIR)/$(APP_NAME)-$(GOOS_VAL)-$(ARCH).tar.gz"; \
	)
	@echo ""
	@echo "🎉 全部编译完成！"
	@ls -lh $(DIST_DIR)/*.tar.gz

# 生成 checksums
checksums:
	@if [ ! -d "$(DIST_DIR)" ] || [ -z "$$(ls -A $(DIST_DIR)/*.tar.gz 2>/dev/null)" ]; then \
		echo "❌ 错误: 没有找到构建产物，请先运行 'make release'"; \
		exit 1; \
	fi
	@echo "🔐 生成校验和..."
	@cd $(DIST_DIR) && (sha256sum *.tar.gz > checksums.txt 2>/dev/null || shasum -a 256 *.tar.gz > checksums.txt)
	@echo "✅ checksums.txt 已生成"
	@echo ""
	@cat $(DIST_DIR)/checksums.txt

# 完整发布流程
all: test release checksums
	@echo ""
	@echo "🎉 发布流程完成！"
