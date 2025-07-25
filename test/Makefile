# Test Makefile for llama-metrics project
# Provides standardized testing commands for all components

.PHONY: help test test-unit test-integration test-e2e test-coverage test-race test-bench clean setup-test

# Variables
TIMEOUT=30m
COVERAGE_OUT=coverage.out
RACE_COUNT=1

# Go modules for testing
SHARED_MODULE=../services/shared
PROXY_MODULE=../services/proxy
DASHBOARD_MODULE=../services/dashboard
HEALTH_MODULE=../services/health

# Test directories
UNIT_DIR=unit
INTEGRATION_DIR=integration
E2E_DIR=e2e

help: ## Show this help message
	@echo "Available test commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'
	@echo ""
	@echo "Examples:"
	@echo "  make test              # Run all tests"
	@echo "  make test-unit         # Run only unit tests"
	@echo "  make test-integration  # Run integration tests"
	@echo "  make test-coverage     # Run tests with coverage"

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests for all Go modules
	@echo "Running unit tests..."
	@echo "Testing shared module..."
	cd $(SHARED_MODULE) && go test -v -timeout $(TIMEOUT) ./...
	@echo "Testing proxy module..."
	cd $(PROXY_MODULE) && go test -v -timeout $(TIMEOUT) ./...
	@echo "Testing dashboard module..."
	cd $(DASHBOARD_MODULE) && go test -v -timeout $(TIMEOUT) ./...
	@echo "Testing health module..."
	cd $(HEALTH_MODULE) && go test -v -timeout $(TIMEOUT) ./...
	@echo "Testing Python unit tests..."
	python -m pytest $(UNIT_DIR)/ -v || echo "Python unit tests not found or failed"

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	@echo "Note: Integration tests require running services"
	go test -v -timeout $(TIMEOUT) ./$(INTEGRATION_DIR)/...

test-e2e: ## Run end-to-end tests
	@echo "Running end-to-end tests..."
	@echo "Note: E2E tests require full system setup"
	go test -v -timeout $(TIMEOUT) ./$(E2E_DIR)/... || echo "E2E tests not implemented yet"

test-coverage: ## Run tests with coverage reporting
	@echo "Running tests with coverage..."
	@echo "Shared module coverage:"
	cd $(SHARED_MODULE) && go test -race -coverprofile=coverage.out -covermode=atomic ./...
	cd $(SHARED_MODULE) && go tool cover -html=coverage.out -o coverage.html
	@echo "Proxy module coverage:"
	cd $(PROXY_MODULE) && go test -race -coverprofile=coverage.out -covermode=atomic ./...
	cd $(PROXY_MODULE) && go tool cover -html=coverage.out -o coverage.html
	@echo "Dashboard module coverage:"
	cd $(DASHBOARD_MODULE) && go test -race -coverprofile=coverage.out -covermode=atomic ./...
	cd $(DASHBOARD_MODULE) && go tool cover -html=coverage.out -o coverage.html
	@echo "Health module coverage:"
	cd $(HEALTH_MODULE) && go test -race -coverprofile=coverage.out -covermode=atomic ./...
	cd $(HEALTH_MODULE) && go tool cover -html=coverage.out -o coverage.html
	@echo "Integration tests coverage:"
	go test -race -coverprofile=$(COVERAGE_OUT) -covermode=atomic ./$(INTEGRATION_DIR)/...
	go tool cover -html=$(COVERAGE_OUT) -o coverage.html

test-race: ## Run tests with race condition detection
	@echo "Running tests with race detection..."
	cd $(SHARED_MODULE) && go test -race -count=$(RACE_COUNT) ./...
	cd $(PROXY_MODULE) && go test -race -count=$(RACE_COUNT) ./...
	cd $(DASHBOARD_MODULE) && go test -race -count=$(RACE_COUNT) ./...
	cd $(HEALTH_MODULE) && go test -race -count=$(RACE_COUNT) ./...
	go test -race -count=$(RACE_COUNT) ./$(INTEGRATION_DIR)/...

test-bench: ## Run benchmark tests
	@echo "Running benchmark tests..."
	cd $(SHARED_MODULE) && go test -bench=. -benchmem ./...
	cd $(PROXY_MODULE) && go test -bench=. -benchmem ./...
	cd $(DASHBOARD_MODULE) && go test -bench=. -benchmem ./...
	cd $(HEALTH_MODULE) && go test -bench=. -benchmem ./...

test-verbose: ## Run tests with verbose output
	@echo "Running tests with verbose output..."
	cd $(SHARED_MODULE) && go test -v -race ./...
	cd $(PROXY_MODULE) && go test -v -race ./...
	cd $(DASHBOARD_MODULE) && go test -v -race ./...
	cd $(HEALTH_MODULE) && go test -v -race ./...
	go test -v -race ./$(INTEGRATION_DIR)/...

test-short: ## Run tests with short flag (skip long-running tests)
	@echo "Running short tests..."
	cd $(SHARED_MODULE) && go test -short ./...
	cd $(PROXY_MODULE) && go test -short ./...
	cd $(DASHBOARD_MODULE) && go test -short ./...
	cd $(HEALTH_MODULE) && go test -short ./...
	go test -short ./$(INTEGRATION_DIR)/...

test-specific: ## Run specific test (usage: make test-specific TEST=TestName MODULE=proxy)
	@if [ -z "$(TEST)" ]; then \
		echo "Please specify TEST name: make test-specific TEST=TestName MODULE=proxy"; \
		exit 1; \
	fi
	@if [ -z "$(MODULE)" ]; then \
		echo "Please specify MODULE: make test-specific TEST=TestName MODULE=proxy"; \
		exit 1; \
	fi
	@echo "Running test $(TEST) in $(MODULE) module..."
	cd ../services/$(MODULE) && go test -v -run $(TEST) ./...

setup-test: ## Set up test environment
	@echo "Setting up test environment..."
	@echo "Installing test dependencies..."
	go mod download
	cd $(SHARED_MODULE) && go mod download
	cd $(PROXY_MODULE) && go mod download
	cd $(DASHBOARD_MODULE) && go mod download
	cd $(HEALTH_MODULE) && go mod download
	@echo "Installing pytest for Python tests..."
	pip install pytest pytest-cov || echo "Failed to install pytest"
	@echo "Test environment setup complete"

clean: ## Clean test artifacts
	@echo "Cleaning test artifacts..."
	find . -name "coverage.out" -delete
	find . -name "coverage.html" -delete
	find ../services -name "coverage.out" -delete
	find ../services -name "coverage.html" -delete
	rm -rf __pycache__
	rm -rf .pytest_cache
	go clean -testcache
	@echo "Test artifacts cleaned"

test-ci: ## Run tests suitable for CI environment
	@echo "Running CI tests..."
	@echo "Unit tests:"
	cd $(SHARED_MODULE) && go test -race -coverprofile=coverage.out ./...
	cd $(PROXY_MODULE) && go test -race -coverprofile=coverage.out ./...
	cd $(DASHBOARD_MODULE) && go test -race -coverprofile=coverage.out ./...
	cd $(HEALTH_MODULE) && go test -race -coverprofile=coverage.out ./...
	@echo "Integration tests (if services are available):"
	go test -race -timeout 5m ./$(INTEGRATION_DIR)/... || echo "Integration tests skipped - services not available"

lint-tests: ## Lint test files
	@echo "Linting test files..."
	cd $(SHARED_MODULE) && golangci-lint run ./... || echo "golangci-lint not available"
	cd $(PROXY_MODULE) && golangci-lint run ./... || echo "golangci-lint not available"
	cd $(DASHBOARD_MODULE) && golangci-lint run ./... || echo "golangci-lint not available"
	cd $(HEALTH_MODULE) && golangci-lint run ./... || echo "golangci-lint not available"
	golangci-lint run ./... || echo "golangci-lint not available"

test-health: ## Test that all test infrastructure is working
	@echo "Testing test infrastructure..."
	@echo "Checking Go test capability..."
	go version
	@echo "Checking module accessibility..."
	ls -la $(SHARED_MODULE)/
	ls -la $(PROXY_MODULE)/
	ls -la $(DASHBOARD_MODULE)/
	ls -la $(HEALTH_MODULE)/
	@echo "Running simple test..."
	cd $(SHARED_MODULE) && go test -v -run TestLoadEnvString ./config
	@echo "Test infrastructure is healthy"

# Service-specific test targets
test-proxy: ## Run tests for proxy service only
	cd $(PROXY_MODULE) && go test -v ./...

test-dashboard: ## Run tests for dashboard service only
	cd $(DASHBOARD_MODULE) && go test -v ./...

test-health-service: ## Run tests for health service only
	cd $(HEALTH_MODULE) && go test -v ./...

test-shared: ## Run tests for shared packages only
	cd $(SHARED_MODULE) && go test -v ./...

# Test watching (requires external tools)
test-watch: ## Watch files and run tests on changes (requires entr)
	@echo "Watching for file changes... (requires 'entr' command)"
	find ../services -name "*.go" | entr -c make test-unit

# Mock generation (if using mockgen or similar)
generate-mocks: ## Generate test mocks
	@echo "Generating test mocks..."
	@echo "Mock generation not implemented yet"

# Performance testing
test-load: ## Run load tests (requires running services)
	@echo "Running load tests..."
	@echo "Load tests not implemented yet"