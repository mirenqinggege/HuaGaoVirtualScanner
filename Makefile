# 华高扫描仪虚拟模拟器 - Makefile
.PHONY: build build-all clean run test deps

# 版本信息
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

BIN_DIR := bin
BIN_NAME := scanner-sim

# 当前平台构建
build:
	@echo "🔨 Building for current platform..."
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME) .

# 全平台交叉编译
build-all: build-linux-amd64 build-linux-arm64 build-darwin-amd64 build-darwin-arm64 build-windows-amd64
	@echo "✅ All platforms built to $(BIN_DIR)/"

build-linux-amd64:
	@echo "🔨 Building linux/amd64..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-amd64 .

build-linux-arm64:
	@echo "🔨 Building linux/arm64..."
	@mkdir -p $(BIN_DIR)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-linux-arm64 .

build-darwin-amd64:
	@echo "🔨 Building darwin/amd64..."
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-darwin-amd64 .

build-darwin-arm64:
	@echo "🔨 Building darwin/arm64..."
	@mkdir -p $(BIN_DIR)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-darwin-arm64 .

build-windows-amd64:
	@echo "🔨 Building windows/amd64..."
	@mkdir -p $(BIN_DIR)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BIN_DIR)/$(BIN_NAME)-windows-amd64.exe .

# 清理构建产物
clean:
	@echo "🧹 Cleaning..."
	@rm -rf $(BIN_DIR)

# 下载依赖
deps:
	go mod tidy
	go mod download

# 运行
run:
	go run . --verbose

# 运行测试
test:
	go test ./... -v

# 格式化代码
fmt:
	go fmt ./...

# 静态检查 (需要安装 golangci-lint)
lint:
	golangci-lint run ./...

# 帮助
help:
	@echo "华高扫描仪虚拟模拟器 - 构建工具"
	@echo ""
	@echo "目标:"
	@echo "  build            - 为当前平台构建"
	@echo "  build-all        - 全平台交叉编译"
	@echo "  clean            - 清理构建产物"
	@echo "  deps             - 下载依赖"
	@echo "  run              - 运行（详细日志模式）"
	@echo "  test             - 运行测试"
	@echo "  fmt              - 格式化代码"
	@echo "  lint             - 静态检查"
	@echo ""
	@echo "用法:"
	@echo "  make build"
	@echo "  make build-all"
	@echo "  make run"
	@echo ""
	@echo "  # 直接运行并指定参数:"
	@echo "  go run . --image-dir ./images --scan-delay 200 --verbose"
