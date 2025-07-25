# Standardized Makefile for llama-metrics services
# Proxy service - Ollama monitoring proxy

# Service-specific variables
BINARY_NAME=ollama-proxy
SERVICE_NAME=proxy
CMD_DIR=cmd/proxy

# Common variables
BUILD_DIR=build
MAIN_FILE=$(CMD_DIR)/main.go
GO=go
GOFLAGS=-v
LDFLAGS=-ldflags="-s -w"

# Version info
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS_VERSION=-ldflags="-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)"

# Colors
GREEN=\033[0;32m
YELLOW=\033[1;33m
RED=\033[0;31m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
.DEFAULT_GOAL := help

.PHONY: all build run clean test fmt lint vet deps help dev install docker

## help: Show this help message
help:
	@echo "$(BLUE)$(SERVICE_NAME) - Makefile Commands$(NC)"
	@echo "======================================"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
	@echo ""

## all: Format, vet, lint, test, and build
all: fmt vet lint test build

## build: Build the service binary
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GO) build $(GOFLAGS) $(LDFLAGS_VERSION) -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_FILE)
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## run: Build and run the service
run: build
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME)

## dev: Run the service in development mode (with hot reload if available)
dev:
	@echo "$(GREEN)Running $(BINARY_NAME) in development mode...$(NC)"
	@$(GO) run $(MAIN_FILE)

## clean: Remove build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@$(GO) clean -cache
	@echo "$(GREEN)Clean complete$(NC)"

## test: Run tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	@$(GO) test -v -race -coverprofile=coverage.out ./...
	@echo "$(GREEN)Tests complete$(NC)"

## test-short: Run short tests
test-short:
	@echo "$(GREEN)Running short tests...$(NC)"
	@$(GO) test -v -short ./...

## fmt: Format code
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	@$(GO) fmt ./...
	@echo "$(GREEN)Formatting complete$(NC)"

## lint: Run golangci-lint
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "$(YELLOW)golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest$(NC)"; \
	fi

## vet: Run go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GO) vet ./...
	@echo "$(GREEN)Vet complete$(NC)"

## deps: Download dependencies
deps:
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@$(GO) mod download
	@$(GO) mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

## install: Install the binary to $GOPATH/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME) to $(GOPATH)/bin...$(NC)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo "$(GREEN)Installation complete$(NC)"

## docker: Build Docker image
docker:
	@echo "$(GREEN)Building Docker image...$(NC)"
	@docker build -t $(SERVICE_NAME):$(VERSION) -t $(SERVICE_NAME):latest .
	@echo "$(GREEN)Docker build complete$(NC)"

## version: Show version information
version:
	@echo "$(BLUE)Version Information$(NC)"
	@echo "==================="
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

# Proxy-specific targets
## run-with-ollama: Run proxy with default Ollama settings
run-with-ollama: build
	@echo "$(GREEN)Running $(BINARY_NAME) with Ollama...$(NC)"
	@./$(BUILD_DIR)/$(BINARY_NAME) --ollama-host localhost --ollama-port 11434

## benchmark: Run performance benchmarks
benchmark:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@$(GO) test -bench=. -benchmem ./...