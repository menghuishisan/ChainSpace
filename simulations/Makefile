.PHONY: all build clean test run docker docker-build docker-push

BINARY_NAME=chainspace-sim
BUILD_DIR=build
IMAGE_NAME=chainspace/simulation
IMAGE_TAG=1.0.0
DEPLOY_DIR=../deploy/images/simulation

all: build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/chainspace-sim

build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/chainspace-sim

clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)

test:
	@echo "Running tests..."
	go test -v ./...

run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME) serve --port=8080

fmt:
	@echo "Formatting code..."
	go fmt ./...

lint:
	@echo "Linting code..."
	golangci-lint run

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Docker相关目标
docker: build-linux docker-build
	@echo "Docker image built: $(IMAGE_NAME):$(IMAGE_TAG)"

docker-build: build-linux
	@echo "Building Docker image..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(DEPLOY_DIR)/
	cd $(DEPLOY_DIR) && docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .
	@rm $(DEPLOY_DIR)/$(BINARY_NAME)

docker-push: docker-build
	@echo "Pushing Docker image..."
	docker push $(IMAGE_NAME):$(IMAGE_TAG)

# 本地开发
dev: build
	./$(BUILD_DIR)/$(BINARY_NAME) serve --port=8080 --debug
