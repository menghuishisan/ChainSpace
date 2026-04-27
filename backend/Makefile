.PHONY: all build run test clean docker migrate lint help

# 变量定义
APP_NAME := chainspace-backend
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | awk '{print $$3}')
LDFLAGS := -ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# 默认目标
all: build

# 构建
build:
	@echo "Building $(APP_NAME)..."
	go build $(LDFLAGS) -o bin/$(APP_NAME) ./cmd/server

# 运行开发服务器
run:
	@echo "Starting development server..."
	go run ./cmd/server

# 运行测试
test:
	@echo "Running tests..."
	go test -v -cover ./...

# 运行测试并生成覆盖率报告
test-coverage:
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# 代码检查
lint:
	@echo "Running linter..."
	golangci-lint run ./...

# 格式化代码
fmt:
	@echo "Formatting code..."
	go fmt ./...
	goimports -w .

# 清理构建产物
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# 构建Docker镜像
docker:
	@echo "Building Docker image..."
	docker build -t chainspace/backend:$(VERSION) .
	docker tag chainspace/backend:$(VERSION) chainspace/backend:latest

# 推送Docker镜像
docker-push:
	@echo "Pushing Docker image..."
	docker push chainspace/backend:$(VERSION)
	docker push chainspace/backend:latest

# 数据库迁移 - 升级
migrate-up:
	@echo "Running database migrations..."
	go run ./cmd/migrate up

# 数据库迁移 - 回滚
migrate-down:
	@echo "Rolling back database migrations..."
	go run ./cmd/migrate down

# 数据库迁移 - 创建新迁移文件
migrate-create:
	@echo "Creating new migration..."
	@read -p "Enter migration name: " name; \
	go run ./cmd/migrate create $$name

# 生成Swagger文档
swagger:
	@echo "Generating Swagger documentation..."
	swag init -g cmd/server/main.go -o api/swagger

# 安装开发依赖
deps:
	@echo "Installing dependencies..."
	go mod download
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# 初始化数据库（创建平台管理员）
init-admin:
	@echo "Initializing platform admin..."
	go run ./cmd/init

# 生成模拟数据
seed:
	@echo "Seeding database with test data..."
	go run ./cmd/seed

# 帮助
help:
	@echo "ChainSpace Backend Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build          - Build the application"
	@echo "  make run            - Run the development server"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage report"
	@echo "  make lint           - Run linter"
	@echo "  make fmt            - Format code"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make docker         - Build Docker image"
	@echo "  make docker-push    - Push Docker image"
	@echo "  make migrate-up     - Run database migrations"
	@echo "  make migrate-down   - Rollback database migrations"
	@echo "  make migrate-create - Create new migration file"
	@echo "  make swagger        - Generate Swagger documentation"
	@echo "  make deps           - Install development dependencies"
	@echo "  make init-admin     - Initialize platform admin"
	@echo "  make seed           - Seed database with test data"
	@echo "  make help           - Show this help message"
