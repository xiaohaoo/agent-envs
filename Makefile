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

.PHONY: build clean test release all

# 默认构建当前平台
build:
	go build -ldflags "$(LDFLAGS)" -o $(APP_NAME) $(CMD_PATH)

# 运行测试
test:
	go test ./...

# 清理构建产物
clean:
	rm -rf $(DIST_DIR) $(APP_NAME)

# 多平台编译
release: clean
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; \
		arch=$${platform#*/}; \
		output=$(APP_NAME); \
		if [ "$$os" = "windows" ]; then output="$(APP_NAME).exe"; fi; \
		echo "🔨 编译 $$os/$$arch ..."; \
		GOOS=$$os GOARCH=$$arch go build -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP_NAME)-$$os-$$arch/$$output $(CMD_PATH) || exit 1; \
		cd $(DIST_DIR) && tar -czf $(APP_NAME)-$$os-$$arch.tar.gz $(APP_NAME)-$$os-$$arch/ && cd ..; \
		echo "  ✅ $(DIST_DIR)/$(APP_NAME)-$$os-$$arch.tar.gz"; \
	done
	@echo ""
	@echo "🎉 全部编译完成！"
	@ls -lh $(DIST_DIR)/*.tar.gz

# 生成 checksums
checksums:
	@cd $(DIST_DIR) && sha256sum *.tar.gz > checksums.txt 2>/dev/null || shasum -a 256 *.tar.gz > checksums.txt
	@echo "✅ checksums.txt 已生成"
	@cat $(DIST_DIR)/checksums.txt

# 完整发布流程
all: test release checksums
