# Makefile for Ollama Monitoring Proxy

# Variables
BINARY_NAME=ollama-proxy
BUILD_DIR=build
CMD_DIR=cmd/proxy
MAIN_FILE=$(CMD_DIR)/main.go
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"

# Version info
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS_VERSION=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Colors
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
NC=\033[0m # No Color

.PHONY: all build run clean test fmt lint vet deps help docker

## help: Show this help message
help:
	@echo "Ollama Monitoring Proxy - Makefile Commands"
	@echo "==========================================="
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
	@echo ""

## all: Build the binary
all: build

## build: Build the proxy binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) $(LDFLAGS_VERSION) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-static: Build a statically linked binary
build-static:
	@echo "$(GREEN)Building static $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@CGO_ENABLED=0 GOOS=linux $(GO) build $(GOFLAGS) $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)✅ Static build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## run: Run the proxy
run:
	@echo "$(GREEN)Starting $(BINARY_NAME)...$(NC)"
	@$(GO) run $(MAIN_FILE)

## run-with-helper: Run proxy with Mac metrics helper (requires sudo)
run-with-helper:
	@echo "$(GREEN)Starting Mac metrics helper (requires sudo)...$(NC)"
	@echo "$(YELLOW)You may be prompted for your password$(NC)"
	@sudo python3 scripts/mac_metrics_helper.py &
	@sleep 2
	@echo "$(GREEN)Starting $(BINARY_NAME)...$(NC)"
	@$(GO) run $(MAIN_FILE)

## run-debug: Run with debug logging
run-debug:
	@echo "$(GREEN)Starting $(BINARY_NAME) in debug mode...$(NC)"
	@LOG_LEVEL=debug $(GO) run $(MAIN_FILE)

## clean: Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning...$(NC)"
	@rm -rf $(BUILD_DIR)
	@$(GO) clean
	@echo "$(GREEN)✅ Clean complete$(NC)"

## test: Run tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	@$(GO) test -v -cover ./...

## test-race: Run tests with race detector
test-race:
	@echo "$(GREEN)Running tests with race detector...$(NC)"
	@$(GO) test -v -race ./...

## benchmark: Run benchmarks
benchmark:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@$(GO) test -bench=. -benchmem ./internal/metrics/

## coverage: Generate test coverage report
coverage:
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@$(GO) test -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✅ Coverage report: coverage.html$(NC)"

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GO) fmt ./...
	@echo "$(GREEN)✅ Format complete$(NC)"

## lint: Run linters
lint:
	@echo "$(GREEN)Running linters...$(NC)"
	@if command -v golangci-lint >/dev/null; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with:$(NC)"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GO) vet ./...
	@echo "$(GREEN)✅ Vet complete$(NC)"

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "$(GREEN)✅ Dependencies ready$(NC)"

## update-deps: Update dependencies
update-deps:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	@$(GO) get -u ./...
	@$(GO) mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

## docker-build: Build Docker image
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(BINARY_NAME):$(VERSION) -t $(BINARY_NAME):latest .
	@echo "$(GREEN)✅ Docker image built: $(BINARY_NAME):$(VERSION)$(NC)"

## docker-run: Run in Docker
docker-run:
	@echo "$(GREEN)Running in Docker...$(NC)"
	@docker run -p 11435:11435 -p 8001:8001 $(BINARY_NAME):latest

## install: Install the binary to /usr/local/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ Installed to /usr/local/bin/$(BINARY_NAME)$(NC)"

## uninstall: Remove the binary from /usr/local/bin
uninstall:
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "$(GREEN)✅ Uninstalled$(NC)"

## dev: Run with hot reload (requires air)
dev:
	@if command -v air >/dev/null; then \
		air; \
	else \
		echo "$(YELLOW)air not installed. Install with:$(NC)"; \
		echo "  go install github.com/cosmtrek/air@latest"; \
		echo ""; \
		echo "Running without hot reload..."; \
		$(MAKE) run; \
	fi

## proto: Generate protobuf files (if needed)
proto:
	@echo "$(GREEN)Generating protobuf files...$(NC)"
	@echo "$(YELLOW)No proto files to generate$(NC)"

# Development shortcuts
.PHONY: b r t c

b: build
r: run
t: test
c: clean