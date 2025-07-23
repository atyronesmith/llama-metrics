# Health Checker Makefile

# Variables
BINARY_NAME=healthcheck
BUILD_DIR=build
GO=go
GOFLAGS=-v

# Colors
GREEN=\033[0;32m
BLUE=\033[0;34m
YELLOW=\033[0;33m
RED=\033[0;31m
NC=\033[0m

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "$(BLUE)Building health checker...$(NC)"
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/healthcheck/main.go
	@echo "$(GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Run in server mode
.PHONY: run
run: build
	@echo "$(BLUE)Running health checker server...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) -mode server

# Run CLI checks
.PHONY: check
check: build
	@echo "$(BLUE)Running comprehensive health check...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) -mode cli -check comprehensive

.PHONY: check-simple
check-simple: build
	@echo "$(BLUE)Running simple health check...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) -mode cli -check simple

.PHONY: check-readiness
check-readiness: build
	@echo "$(BLUE)Running readiness check...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) -mode cli -check readiness

.PHONY: check-liveness
check-liveness: build
	@echo "$(BLUE)Running liveness check...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) -mode cli -check liveness

.PHONY: check-analyzed
check-analyzed: build
	@echo "$(BLUE)Running comprehensive health check with LLM analysis...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME) -mode cli -check analyzed

# Download dependencies
.PHONY: deps
deps:
	@echo "$(BLUE)Downloading dependencies...$(NC)"
	$(GO) mod download
	$(GO) mod tidy
	@echo "$(GREEN)✅ Dependencies updated$(NC)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@echo "$(GREEN)✅ Clean complete$(NC)"

# Test
.PHONY: test
test:
	@echo "$(BLUE)Running tests...$(NC)"
	$(GO) test -v ./...

# Install binary
.PHONY: install
install: build
	@echo "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "$(GREEN)✅ Installed to /usr/local/bin/$(BINARY_NAME)$(NC)"

# Help
.PHONY: help
help:
	@echo "$(BLUE)Health Checker Makefile Commands:$(NC)"
	@echo "  make build         - Build the health checker"
	@echo "  make run           - Run health checker server"
	@echo "  make check         - Run comprehensive health check"
	@echo "  make check-simple  - Run simple health check"
	@echo "  make check-readiness - Run readiness check"
	@echo "  make check-liveness  - Run liveness check"
	@echo "  make check-analyzed  - Run health check with LLM analysis"
	@echo "  make deps          - Download dependencies"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make test          - Run tests"
	@echo "  make install       - Install binary to /usr/local/bin"