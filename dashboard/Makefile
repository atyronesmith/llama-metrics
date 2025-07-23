.PHONY: all build run clean test deps

# Variables
BINARY_NAME=dashboard
MAIN_PATH=cmd/dashboard/main.go
BUILD_DIR=build

# Default target
all: build

# Build the application
build:
	@echo "Building dashboard..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Run the application
run:
	@echo "Starting dashboard..."
	go run $(MAIN_PATH)

# Run with specific environment
run-dev:
	@echo "Starting dashboard in development mode..."
	DASHBOARD_ENV=development go run $(MAIN_PATH)

run-prod:
	@echo "Starting dashboard in production mode..."
	DASHBOARD_ENV=production go run $(MAIN_PATH)

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean complete"

# Run tests
test:
	@echo "Running tests..."
	go test -v ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "Dependencies installed"

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "Formatting complete"

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run
	@echo "Linting complete"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# macOS
	GOOS=darwin GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	# Linux
	GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	@echo "Multi-platform build complete"

# Development with hot reload (requires air)
dev:
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	@echo "Starting with hot reload..."
	air

# Help
help:
	@echo "Available targets:"
	@echo "  make build    - Build the application"
	@echo "  make run      - Run the application"
	@echo "  make run-dev  - Run in development mode"
	@echo "  make run-prod - Run in production mode"
	@echo "  make clean    - Clean build artifacts"
	@echo "  make test     - Run tests"
	@echo "  make deps     - Download dependencies"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Run linter"
	@echo "  make dev      - Run with hot reload"
	@echo "  make help     - Show this help message"