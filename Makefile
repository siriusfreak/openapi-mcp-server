# Variables
BINARY_NAME=openapi-mcp-server
GO=go
GOFLAGS=-v
LDFLAGS=-s -w

# Paths
BUILD_DIR=build
CMD_DIR=github.com/siriusfreak/openapi-mcp-server/cmd/server

# Git information for versioning
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
VERSION=1.0.0

# Add version info to binary
LDFLAGS+= -X main.Version=$(VERSION)
LDFLAGS+= -X main.GitCommit=$(GIT_COMMIT)
LDFLAGS+= -X main.BuildTime=$(BUILD_TIME)

.PHONY: all build clean test run install deps lint fmt help

# Default target
all: clean build

# Show help
help:
	@echo "Available targets:"
	@echo "  build       - Build the binary"
	@echo "  clean       - Remove build artifacts"
	@echo "  test        - Run tests"
	@echo "  run         - Run the server"
	@echo "  install     - Install the binary to GOPATH/bin"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  lint        - Run linter (requires golangci-lint)"
	@echo "  fmt         - Format code"
	@echo "  run-example - Run with example Petstore spec"
	@echo "  run-http    - Run in HTTP mode"
	@echo "  docker-build - Build Docker image"
	@echo "  docker-run  - Run Docker container"

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(CMD_DIR)
	# Linux ARM64
	GOOS=linux GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(CMD_DIR)
	# Darwin AMD64
	GOOS=darwin GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(CMD_DIR)
	# Darwin ARM64 (M1/M2)
	GOOS=darwin GOARCH=arm64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(CMD_DIR)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 $(GO) build -ldflags="$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(CMD_DIR)
	@echo "Multi-platform build complete"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@$(GO) clean

# Run tests
test:
	@echo "Running tests..."
	$(GO) test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the server
run:
	$(GO) run $(CMD_DIR)

# Install dependencies
deps:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Lint the code (requires golangci-lint)
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@which goimports > /dev/null && goimports -w . || echo "goimports not found, skipping import formatting"

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "Installed to $(GOPATH)/bin/$(BINARY_NAME)"

# Run with example specs
run-example:
	$(GO) run $(CMD_DIR)/main.go \
		--specs https://petstore.swagger.io/v2/swagger.json \
		--mode stdio

# Run in HTTP mode
run-http:
	$(GO) run $(CMD_DIR)/main.go \
		--specs https://petstore.swagger.io/v2/swagger.json \
		--mode http \
		--port 3000

# Run with multiple specs
run-multiple:
	$(GO) run $(CMD_DIR)/main.go \
		--specs https://petstore.swagger.io/v2/swagger.json \
		--specs https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml \
		--mode stdio

# Docker build
docker-build:
	@echo "Building Docker image..."
	docker build -t $(BINARY_NAME):latest -t $(BINARY_NAME):$(VERSION) .

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run -it --rm -p 3000:3000 $(BINARY_NAME):latest

# Generate mocks (requires mockgen)
mocks:
	@echo "Generating mocks..."
	@which mockgen > /dev/null || (echo "mockgen not found. Install it with: go install github.com/golang/mock/mockgen@latest" && exit 1)
	mockgen -source=internal/registry/registry.go -destination=internal/mocks/registry_mock.go -package=mocks

# Check for security vulnerabilities
security:
	@echo "Checking for vulnerabilities..."
	$(GO) list -json -m all | nancy sleuth

# Update dependencies
update-deps:
	@echo "Updating dependencies..."
	$(GO) get -u ./...
	$(GO) mod tidy
