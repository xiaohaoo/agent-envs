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

.PHONY: help build run install clean test vet fmt lint release checksums all

# 默认目标：显示帮助
help:
	@echo "Agent Envs - 可用命令："
	@echo ""
	@echo "  make build      - 编译当前平台"
	@echo "  make run        - 编译并运行"
	@echo "  make install    - 安装到 /usr/local/bin"
	@echo "  make test       - 运行测试"
	@echo "  make vet        - 运行 go vet"
	@echo "  make fmt        - 格式化代码"
	@echo "  make lint       - 运行 golangci-lint（如果已安装）"
	@echo "  make release    - 多平台编译"
	@echo "  make checksums  - 生成校验和"
	@echo "  make all        - 测试 + 编译 + 校验和"
	@echo "  make clean      - 清理构建产物"
	@echo ""

# 编译当前平台
build:
	@echo "🔨 编译 $(APP_NAME) ($(VERSION))..."
	@$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(APP_NAME) $(CMD_PATH)
	@echo "✅ 编译完成: ./$(APP_NAME)"

# 编译并运行
run: build
	@echo "🚀 运行 $(APP_NAME)..."
	@./$(APP_NAME)

# 安装到系统路径
install: build
	@echo "📦 安装到 /usr/local/bin/$(APP_NAME)..."
	@sudo mv $(APP_NAME) /usr/local/bin/
	@echo "✅ 安装完成"

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
	@rm -rf $(DIST_DIR) $(APP_NAME) $(APP_NAME).exe coverage.out
	@echo "✅ 清理完成"

# 多平台编译
release: clean vet
	@echo "🚀 开始多平台编译..."
	@mkdir -p $(DIST_DIR)
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS := $(word 1,$(subst /, ,$(platform)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(platform)))) \
		$(eval OUTPUT := $(APP_NAME)$(if $(filter windows,$(OS)),.exe,)) \
		echo "🔨 编译 $(OS)/$(ARCH)..."; \
		GOOS=$(OS) GOARCH=$(ARCH) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" \
			-o $(DIST_DIR)/$(APP_NAME)-$(OS)-$(ARCH)/$(OUTPUT) $(CMD_PATH) || exit 1; \
		cd $(DIST_DIR) && tar -czf $(APP_NAME)-$(OS)-$(ARCH).tar.gz $(APP_NAME)-$(OS)-$(ARCH)/ && cd ..; \
		echo "  ✅ $(DIST_DIR)/$(APP_NAME)-$(OS)-$(ARCH).tar.gz"; \
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
