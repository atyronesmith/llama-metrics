# Ollama Monitoring Stack Makefile
# Run 'make help' to see all available targets

.PHONY: help setup install clean start stop restart status logs test traffic metrics health lint commit push all

# Default target
.DEFAULT_GOAL := help

# Go build targets
GO_PROXY_DIR := services/proxy
GO_DASHBOARD_DIR := services/dashboard

# Variables
VENV := venv
PYTHON := $(VENV)/bin/python
PIP := $(VENV)/bin/pip
SHELL_SCRIPTS := *.sh

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
BLUE := \033[0;34m
NC := \033[0m # No Color

## help: Show this help message
help:
	@echo "$(BLUE)🚀 Ollama Monitoring Stack - Mac M-Series Edition$(NC)"
	@echo "================================================="
	@echo ""
	@echo "$(GREEN)🎯 QUICK START (New Users):$(NC)"
	@echo "  make setup start        # Complete automated install & start"
	@echo "  make traffic           # Generate test traffic"
	@echo ""
	@echo "$(GREEN)🛠️  SETUP COMMANDS:$(NC)"
	@grep -E '^## (setup|check-system|install-|pull-model|quick-setup)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(GREEN)🔨 BUILD COMMANDS:$(NC)"
	@grep -E '^## (build|build-proxy|build-dashboard|build-all)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(GREEN)🚦 SERVICE COMMANDS:$(NC)"
	@grep -E '^## (start|stop|restart|status)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(GREEN)📊 MONITORING & TESTING:$(NC)"
	@grep -E '^## (traffic|load-test|dashboard|metrics|health|logs)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(GREEN)🔧 UTILITIES:$(NC)"
	@grep -E '^## (clean|lint|test|validate)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(BLUE)📖 More Info:$(NC) See SETUP.md for detailed installation guide"

## build: Build all Go components (proxy and dashboard)
build: build-proxy build-dashboard
	@echo "$(GREEN)✅ All components built successfully$(NC)"

## build-proxy: Build the monitoring proxy
build-proxy:
	@echo "$(BLUE)Building monitoring proxy...$(NC)"
	@cd $(GO_PROXY_DIR) && make build
	@echo "$(GREEN)✅ Proxy built: $(GO_PROXY_DIR)/build/ollama-proxy$(NC)"

## build-dashboard: Build the dashboard
build-dashboard:
	@echo "$(BLUE)Building dashboard...$(NC)"
	@cd $(GO_DASHBOARD_DIR) && make build
	@echo "$(GREEN)✅ Dashboard built: $(GO_DASHBOARD_DIR)/build/llama-dashboard$(NC)"

## build-health: Build the health checker
build-health:
	@echo "$(BLUE)Building health checker...$(NC)"
	@cd services/health && make build
	@echo "$(GREEN)✅ Health checker built: services/health/build/llama-health$(NC)"

## build-mac-metrics: Build the Go Mac metrics service
build-mac-metrics:
	@echo "$(BLUE)Building Mac metrics service...$(NC)"
	@cd services/mac-metrics && make build
	@echo "$(GREEN)✅ Mac metrics service built: services/mac-metrics/build/mac-metrics$(NC)"

## build-all: Build all components for all platforms
build-all:
	@echo "$(BLUE)Building all components for multiple platforms...$(NC)"
	@cd $(GO_PROXY_DIR) && make build-all
	@cd $(GO_DASHBOARD_DIR) && make build-all
	@cd services/health && make build
	@echo "$(GREEN)✅ All platform builds complete$(NC)"

## run-proxy: Run the proxy directly with optimized settings (for debugging)
run-proxy:
	@echo "$(BLUE)Running proxy in foreground with optimized settings...$(NC)"
	@echo "$(YELLOW)Max concurrency: 4, Max queue: 100$(NC)"
	@cd $(GO_PROXY_DIR) && go run cmd/proxy/main.go --max-concurrency 4

## run-dashboard: Run the dashboard directly (for debugging)
run-dashboard:
	@echo "$(BLUE)Running dashboard in foreground...$(NC)"
	@cd $(GO_DASHBOARD_DIR) && make run

## check-system: Verify Mac M-series and system requirements
check-system:
	@echo "$(BLUE)Checking system requirements...$(NC)"
	@if [[ "$$(uname -s)" != "Darwin" ]]; then \
		echo "$(RED)❌ This setup is designed for macOS only$(NC)"; \
		exit 1; \
	fi
	@if [[ "$$(uname -m)" != "arm64" ]]; then \
		echo "$(YELLOW)⚠️  Warning: This setup is optimized for M-series Macs (arm64)$(NC)"; \
		echo "$(YELLOW)   Your system: $$(uname -m)$(NC)"; \
	else \
		echo "$(GREEN)✅ Mac M-series detected$(NC)"; \
	fi
	@if ! command -v python3 >/dev/null 2>&1; then \
		echo "$(RED)❌ Python 3 is required but not installed$(NC)"; \
		echo "$(YELLOW)Please install Python 3 from https://python.org$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ Python 3 found: $$(python3 --version)$(NC)"; \
	fi
	@if ! command -v curl >/dev/null 2>&1; then \
		echo "$(RED)❌ curl is required but not installed$(NC)"; \
		exit 1; \
	else \
		echo "$(GREEN)✅ curl found$(NC)"; \
	fi

## install-ollama: Install Ollama if not present
install-ollama:
	@echo "$(BLUE)Checking Ollama installation...$(NC)"
	@if command -v ollama >/dev/null 2>&1; then \
		echo "$(GREEN)✅ Ollama already installed: $$(ollama --version)$(NC)"; \
	else \
		echo "$(YELLOW)Installing Ollama for macOS...$(NC)"; \
		curl -fsSL https://ollama.ai/install.sh | sh; \
		if command -v ollama >/dev/null 2>&1; then \
			echo "$(GREEN)✅ Ollama installed successfully$(NC)"; \
		else \
			echo "$(RED)❌ Ollama installation failed$(NC)"; \
			echo "$(YELLOW)Please install manually:$(NC)"; \
			echo "  1. Visit https://ollama.ai"; \
			echo "  2. Download Ollama for Mac"; \
			echo "  3. Run the installer"; \
			echo "  4. Restart terminal and run 'make setup' again"; \
			exit 1; \
		fi \
	fi

## install-prometheus: Install Prometheus if not present
install-prometheus:
	@echo "$(BLUE)Checking Prometheus installation...$(NC)"
	@if command -v prometheus >/dev/null 2>&1; then \
		echo "$(GREEN)✅ Prometheus already installed$(NC)"; \
	elif command -v brew >/dev/null 2>&1; then \
		echo "$(YELLOW)Installing Prometheus via Homebrew...$(NC)"; \
		brew install prometheus; \
		if command -v prometheus >/dev/null 2>&1; then \
			echo "$(GREEN)✅ Prometheus installed successfully$(NC)"; \
		else \
			echo "$(RED)❌ Prometheus installation failed$(NC)"; \
			exit 1; \
		fi \
	else \
		echo "$(YELLOW)Homebrew not found. Providing manual installation instructions:$(NC)"; \
		echo ""; \
		echo "$(BLUE)To install Prometheus manually:$(NC)"; \
		echo "  1. Install Homebrew: /bin/bash -c \"\$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""; \
		echo "  2. Run: brew install prometheus"; \
		echo "  3. Or download from: https://prometheus.io/download/"; \
		echo ""; \
		echo "$(YELLOW)For now, continuing with setup (Prometheus will start when available)$(NC)"; \
	fi

## setup: Complete automated setup for Mac M-series (install everything)
setup: check-system install-ollama install-prometheus venv install pull-model validate
	@echo "$(GREEN)🚀 Complete setup finished!$(NC)"
	@echo "$(BLUE)Run 'make start' to launch the monitoring stack$(NC)"

## quick-setup: Setup with existing Ollama/Prometheus
quick-setup: venv install pull-model validate
	@echo "$(GREEN)✅ Quick setup complete!$(NC)"

## venv: Create Python virtual environment
venv:
	@if [ ! -d "$(VENV)" ]; then \
		echo "$(BLUE)Creating virtual environment...$(NC)"; \
		python3 -m venv $(VENV); \
		echo "$(GREEN)✅ Virtual environment created$(NC)"; \
	else \
		echo "$(YELLOW)Virtual environment already exists$(NC)"; \
	fi

## install: Install dependencies and show configuration recommendations
install: venv
	@echo "$(BLUE)Installing Python dependencies...$(NC)"
	@$(PIP) install --upgrade pip
	@$(PIP) install -r python/requirements_all.txt
	@echo "$(GREEN)✅ All dependencies installed$(NC)"
	@echo ""
	@echo "$(YELLOW)⚡ Performance Recommendations:$(NC)"
	@echo "  • Ollama will be started with OLLAMA_NUM_PARALLEL=2 to prevent resource contention"
	@echo "  • Proxy will use max concurrency of 4 to match Ollama's capabilities"
	@echo "  • To customize: Set MAX_CONCURRENCY and OLLAMA_NUM_PARALLEL environment variables"
	@echo ""

## pull-model: Pull and verify phi3:mini model for Ollama
pull-model:
	@echo "$(BLUE)Checking if phi3:mini model is available...$(NC)"
	@if ollama list | grep -q "phi3:mini"; then \
		echo "$(GREEN)✅ phi3:mini model already available$(NC)"; \
	else \
		echo "$(YELLOW)Downloading phi3:mini model (this may take a few minutes)...$(NC)"; \
		if ! ollama serve > /dev/null 2>&1 & OLLAMA_PID=$$!; then \
			echo "$(RED)❌ Failed to start Ollama$(NC)"; \
			exit 1; \
		fi; \
		sleep 3; \
		if ollama pull phi3:mini; then \
			echo "$(GREEN)✅ phi3:mini model downloaded successfully$(NC)"; \
		else \
			echo "$(RED)❌ Failed to download phi3:mini model$(NC)"; \
			kill $$OLLAMA_PID 2>/dev/null || true; \
			exit 1; \
		fi; \
		kill $$OLLAMA_PID 2>/dev/null || true; \
	fi

## start: Start all monitoring services with dashboard
start: start-ollama start-proxy start-prometheus start-dashboard start-health start-mac-metrics
	@echo "$(GREEN)🚀 All services started successfully!$(NC)"
	@echo ""
	@echo "$(BLUE)📊 Dashboard:$(NC)        http://localhost:3001"
	@echo "$(BLUE)📈 Prometheus UI:$(NC)    http://localhost:9090"
	@echo "$(BLUE)🔧 Metrics API:$(NC)      http://localhost:8001/metrics"
	@echo "$(BLUE)🤖 Ollama API:$(NC)       http://localhost:11434"
	@echo ""
	@echo "$(YELLOW)💡 Pro tip: Run 'make traffic' to generate test traffic$(NC)"


## start-ollama: Start Ollama service with optimized settings
start-ollama:
	@if ! pgrep -x "ollama" > /dev/null; then \
		echo "$(BLUE)Starting Ollama with optimized settings...$(NC)"; \
		echo "$(YELLOW)Setting OLLAMA_NUM_PARALLEL=2 to prevent resource contention$(NC)"; \
		OLLAMA_NUM_PARALLEL=2 OLLAMA_MAX_LOADED_MODELS=2 ollama serve > ollama.log 2>&1 & \
		sleep 3; \
		echo "$(GREEN)✅ Ollama started with parallel limit of 2$(NC)"; \
	else \
		echo "$(YELLOW)Ollama is already running$(NC)"; \
		echo "$(YELLOW)Note: If experiencing issues, restart with: make stop-ollama && make start-ollama$(NC)"; \
	fi
	@echo "$(BLUE)Ollama monitoring ready$(NC)"

## start-proxy: Start the monitoring proxy with optimized settings
start-proxy:
	@if ! pgrep -f "ollama-proxy" > /dev/null; then \
		echo "$(BLUE)Building and starting Go monitoring proxy...$(NC)"; \
		echo "$(YELLOW)Using max concurrency of 4 to match Ollama's capabilities$(NC)"; \
		cd services/proxy && make build && ./build/ollama-proxy --max-concurrency 4 > ../../proxy.log 2>&1 & \
		cd ../..; \
		sleep 2; \
		echo "$(GREEN)✅ Go monitoring proxy started with concurrency limit of 4$(NC)"; \
	else \
		echo "$(YELLOW)Monitoring proxy is already running$(NC)"; \
	fi



## start-prometheus: Start Prometheus
start-prometheus:
	@if ! (podman ps 2>/dev/null | grep -q prometheus); then \
		echo "$(BLUE)Starting Prometheus container...$(NC)"; \
		./scripts/run_prometheus.sh; \
	else \
		echo "$(YELLOW)Prometheus container is already running$(NC)"; \
	fi

## stop: Stop all monitoring services
stop: stop-proxy stop-monitoring-ollama stop-prometheus stop-dashboard stop-health stop-mac-metrics
	@echo "$(GREEN)✅ All monitoring services stopped$(NC)"

## stop-proxy: Stop the monitoring proxy
stop-proxy:
	@echo "$(BLUE)Stopping monitoring proxy...$(NC)"
	@pkill -f "ollama-proxy" 2>/dev/null || true
	@lsof -ti:11435 | xargs kill -9 2>/dev/null || true
	@lsof -ti:8001 | xargs kill -9 2>/dev/null || true
	@echo "$(GREEN)✅ Monitoring proxy stopped$(NC)"

## stop-monitoring-ollama: Stop the dedicated monitoring Ollama instance
stop-monitoring-ollama:
	@echo "$(BLUE)Stopping monitoring Ollama instance...$(NC)"
	@lsof -ti:11435 | xargs kill -9 2>/dev/null || true
	@echo "$(GREEN)✅ Monitoring Ollama instance stopped$(NC)"



## stop-prometheus: Stop Prometheus
stop-prometheus:
	@echo "$(BLUE)Stopping Prometheus...$(NC)"
	@pkill -x "prometheus" 2>/dev/null || true
	@(podman stop prometheus 2>/dev/null || docker stop prometheus 2>/dev/null) || true
	@echo "$(GREEN)✅ Prometheus stopped$(NC)"

## restart: Restart all monitoring services
restart: stop start

## status: Show status of all services
status:
	@echo "$(BLUE)Service Status:$(NC)"
	@echo "==============="
	@if pgrep -x "ollama" > /dev/null; then \
		echo "$(GREEN)✅ Ollama: Running$(NC)"; \
	else \
		echo "$(RED)❌ Ollama: Not running$(NC)"; \
	fi
	@if pgrep -f "ollama-proxy" > /dev/null || lsof -ti:11435 > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Monitoring Proxy (Go): Running$(NC)"; \
		if lsof -ti:8001 > /dev/null 2>&1; then \
			echo "    ├─ Proxy:   http://localhost:11435"; \
			echo "    └─ Metrics: http://localhost:8001/metrics"; \
		fi \
	else \
		echo "$(RED)❌ Monitoring Proxy: Not running$(NC)"; \
	fi
	@if pgrep -x "prometheus" > /dev/null; then \
		echo "$(GREEN)✅ Prometheus: Running (native)$(NC)"; \
		echo "    └─ UI: http://localhost:9090"; \
	elif (podman ps 2>/dev/null || docker ps 2>/dev/null) | grep -q prometheus; then \
		echo "$(GREEN)✅ Prometheus: Running (container)$(NC)"; \
		echo "    └─ UI: http://localhost:9090"; \
	else \
		echo "$(RED)❌ Prometheus: Not running$(NC)"; \
	fi
	@if pgrep -f "dashboard" > /dev/null || lsof -ti:3001 > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Dashboard (Go): Running$(NC)"; \
		echo "    └─ URL: http://localhost:3001"; \
	else \
		echo "$(RED)❌ Dashboard: Not running$(NC)"; \
	fi
	@if pgrep -f "llama-health.*server" > /dev/null || lsof -ti:8080 > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Health Checker: Running$(NC)"; \
		echo "    └─ Server: http://localhost:8080"; \
	else \
		echo "$(RED)❌ Health Checker: Not running$(NC)"; \
	fi
	@if pgrep -f "mac-metrics" > /dev/null || lsof -ti:8002 > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Mac Metrics: Running$(NC)"; \
		echo "    └─ Server: http://localhost:8002"; \
	else \
		echo "$(RED)❌ Mac Metrics: Not running$(NC)"; \
	fi

## logs: Tail all service logs
logs:
	@echo "$(BLUE)Tailing logs (Ctrl+C to stop)...$(NC)"
	@tail -f proxy.log prometheus.log ollama.log dashboard.log health.log mac_metrics.log 2>/dev/null || echo "$(YELLOW)No log files found$(NC)"

## logs-proxy: Tail proxy logs
logs-proxy:
	@tail -f proxy.log proxy_fixed.log 2>/dev/null || echo "$(YELLOW)No proxy logs found$(NC)"

## traffic: Generate traffic (interactive mode)
traffic: venv
	@./scripts/traffic/run.sh

## traffic-quick: Quick traffic test (10 requests)
traffic-quick: venv
	@./scripts/traffic/run.sh --quick

## traffic-demo: Demo traffic (50 requests)
traffic-demo: venv
	@./scripts/traffic/run.sh --demo

## traffic-stress: Stress test (1000 requests)
traffic-stress: venv
	@./scripts/traffic/run.sh --stress

## traffic-continuous: Continuous traffic generation
traffic-continuous: venv
	@echo "$(BLUE)Starting continuous traffic generation...$(NC)"
	@./scripts/traffic/simple.sh


## metrics: Show current metrics
metrics:
	@echo "$(BLUE)Current Metrics:$(NC)"
	@curl -s http://localhost:8001/metrics | grep -E "^ollama_proxy_requests_total|^ollama_proxy_active_requests" | head -20

## health: Check health of all services
health: build-health
	@echo "$(BLUE)Checking comprehensive system health...$(NC)"
	@services/health/build/llama-health -config config/llama-metrics.yml -mode cli -check comprehensive

## health-simple: Quick health check
health-simple: build-health
	@echo "$(BLUE)Quick health check...$(NC)"
	@services/health/build/llama-health -config config/llama-metrics.yml -mode cli -check simple

## health-readiness: Check if system is ready
health-readiness: build-health
	@echo "$(BLUE)Checking system readiness...$(NC)"
	@services/health/build/llama-health -config config/llama-metrics.yml -mode cli -check readiness

## health-liveness: Check if system is alive
health-liveness: build-health
	@echo "$(BLUE)Checking system liveness...$(NC)"
	@services/health/build/llama-health -config config/llama-metrics.yml -mode cli -check liveness

## health-server: Run health check server
health-server: build-health
	@echo "$(BLUE)Starting health check server on port 8080...$(NC)"
	@services/health/build/llama-health -config config/llama-metrics.yml -mode server -port 8080

## health-analyzed: Run health check with LLM analysis
health-analyzed: build-health
	@echo "$(BLUE)Running health check with AI-powered analysis...$(NC)"
	@services/health/build/llama-health -config config/llama-metrics.yml -mode cli -check analyzed

## prometheus-ui: Open Prometheus UI in browser
prometheus-ui:
	@echo "$(BLUE)Opening Prometheus UI...$(NC)"
	@open http://localhost:9090 || xdg-open http://localhost:9090 || echo "$(YELLOW)Please open http://localhost:9090 in your browser$(NC)"

## test: Run monitoring tests
test:
	@echo "$(BLUE)Running monitoring tests...$(NC)"
	@cd test && make test

## lint: Run shellcheck on all shell scripts
lint:
	@echo "$(BLUE)Running shellcheck...$(NC)"
	@if command -v shellcheck >/dev/null 2>&1; then \
		shellcheck $(SHELL_SCRIPTS) && echo "$(GREEN)✅ All shell scripts pass shellcheck$(NC)"; \
	else \
		echo "$(RED)❌ shellcheck not installed$(NC)"; \
		echo "Install with: brew install shellcheck"; \
	fi

## validate: Validate setup and configuration
validate: lint
	@echo "$(BLUE)Validating setup...$(NC)"
	@if [ -d "$(VENV)" ]; then \
		echo "$(GREEN)✅ Virtual environment exists$(NC)"; \
	else \
		echo "$(RED)❌ Virtual environment missing$(NC)"; \
	fi
	@if [ -f "config/prometheus/prometheus.yml" ]; then \
		echo "$(GREEN)✅ Prometheus config exists$(NC)"; \
	else \
		echo "$(RED)❌ Prometheus config missing$(NC)"; \
	fi
	@if command -v ollama >/dev/null 2>&1; then \
		echo "$(GREEN)✅ Ollama installed$(NC)"; \
	else \
		echo "$(RED)❌ Ollama not installed$(NC)"; \
	fi

## clean: Clean up generated files and logs
clean: clean-go
	@echo "$(BLUE)Cleaning up...$(NC)"
	@rm -f *.log
	@rm -f monitoring_pids.txt
	@rm -rf __pycache__
	@find . -name "*.pyc" -delete
	@echo "$(GREEN)✅ Cleanup complete$(NC)"

## clean-go: Clean Go build artifacts
clean-go:
	@echo "$(BLUE)Cleaning Go build artifacts...$(NC)"
	@cd $(GO_PROXY_DIR) && make clean
	@cd $(GO_DASHBOARD_DIR) && make clean
	@cd services/health && make clean
	@echo "$(GREEN)✅ Go cleanup complete$(NC)"

## clean-all: Clean everything including venv
clean-all: clean
	@echo "$(BLUE)Removing virtual environment...$(NC)"
	@rm -rf $(VENV)
	@echo "$(GREEN)✅ Full cleanup complete$(NC)"

## commit: Git add and commit all changes
commit:
	@echo "$(BLUE)Committing changes...$(NC)"
	@git add -A
	@git commit -m "Update monitoring stack" || echo "$(YELLOW)Nothing to commit$(NC)"

## push: Push changes to remote
push:
	@echo "$(BLUE)Pushing to remote...$(NC)"
	@git push origin main

## dev: Start development environment (ollama + proxy + traffic)
dev: start
	@echo "$(BLUE)Starting development environment...$(NC)"
	@sleep 2
	@make traffic-continuous

## demo: Run a complete demo
demo: setup start
	@echo "$(BLUE)Running demo...$(NC)"
	@sleep 3
	@make traffic-demo
	@echo ""
	@echo "$(GREEN)Demo complete!$(NC)"
	@echo "$(BLUE)View metrics at:$(NC) http://localhost:8001/metrics"
	@echo "$(BLUE)View Prometheus at:$(NC) http://localhost:9090"

## docker-prometheus: Run Prometheus in Docker/Podman
docker-prometheus:
	@./scripts/run_prometheus.sh

## all: Complete setup, start services, and run demo
all: setup start demo

# Advanced targets for development

## debug-proxy: Run proxy in foreground for debugging
debug-proxy:
	@echo "$(BLUE)Running proxy in debug mode...$(NC)"
	@cd services/proxy && make dev

## watch-metrics: Continuously watch metrics
watch-metrics:
	@watch -n 2 'curl -s http://localhost:8001/metrics | grep -E "^ollama_proxy" | head -20'

## benchmark: Run performance benchmark
benchmark: venv
	@echo "$(BLUE)Running performance benchmark...$(NC)"
	@$(PYTHON) -c "print('Starting benchmark with 100 requests...')"
	@./scripts/traffic/run.sh --quick
	@sleep 2
	@make metrics

## install-tools: Install required system tools
install-tools:
	@echo "$(BLUE)Installing required tools...$(NC)"
	@if [[ "$$(uname)" == "Darwin" ]]; then \
		brew install shellcheck jq watch || true; \
	else \
		echo "$(YELLOW)Please install: shellcheck jq watch$(NC)"; \
	fi

## dashboard: Start the web dashboard
dashboard:
	@echo "$(BLUE)Starting Ollama Dashboard...$(NC)"
	@cd services/dashboard && make run

## start-dashboard: Start dashboard in background
start-dashboard:
	@if ! pgrep -f "dashboard" > /dev/null && ! lsof -ti:3001 > /dev/null 2>&1; then \
		echo "$(BLUE)Building and starting Go dashboard...$(NC)"; \
		cd services/dashboard && make build && ./build/llama-dashboard > ../../dashboard.log 2>&1 & \
		cd ../..; \
		sleep 2; \
		echo "$(GREEN)✅ Dashboard started at http://localhost:3001$(NC)"; \
	else \
		echo "$(YELLOW)Dashboard is already running$(NC)"; \
	fi

## start-health: Start health checker server
start-health: build-health
	@if ! pgrep -f "llama-health.*server" > /dev/null && ! lsof -ti:8080 > /dev/null 2>&1; then \
		echo "$(BLUE)Starting health checker server...$(NC)"; \
		services/health/build/llama-health -config config/llama-metrics.yml -mode server -port 8080 > health.log 2>&1 & \
		sleep 2; \
		echo "$(GREEN)✅ Health checker server started on port 8080$(NC)"; \
	else \
		echo "$(YELLOW)Health checker server is already running$(NC)"; \
	fi

## start-mac-metrics: Start Go Mac system metrics server
start-mac-metrics: build-mac-metrics
	@if ! pgrep -f "mac-metrics" > /dev/null && ! lsof -ti:8002 > /dev/null 2>&1; then \
		echo "$(BLUE)Starting Go Mac system metrics server...$(NC)"; \
		services/mac-metrics/build/mac-metrics > mac_metrics.log 2>&1 & \
		sleep 2; \
		echo "$(GREEN)✅ Go Mac metrics server started on port 8002$(NC)"; \
	else \
		echo "$(YELLOW)Mac metrics server is already running$(NC)"; \
	fi

## stop-dashboard: Stop dashboard
stop-dashboard:
	@echo "$(BLUE)Stopping dashboard...$(NC)"
	@pkill -f "dashboard" 2>/dev/null || true
	@lsof -ti:3001 | xargs kill -9 2>/dev/null || true
	@echo "$(GREEN)✅ Dashboard stopped$(NC)"

## stop-health: Stop health checker server
stop-health:
	@echo "$(BLUE)Stopping health checker server...$(NC)"
	@pkill -f "llama-health.*server" 2>/dev/null || true
	@lsof -ti:8080 | xargs kill -9 2>/dev/null || true
	@echo "$(GREEN)✅ Health checker server stopped$(NC)"

## stop-mac-metrics: Stop Go Mac system metrics server
stop-mac-metrics:
	@echo "$(BLUE)Stopping Mac metrics server...$(NC)"
	@pkill -f "mac-metrics" 2>/dev/null || true
	@lsof -ti:8002 | xargs kill -9 2>/dev/null || true
	@echo "$(GREEN)✅ Mac metrics server stopped$(NC)"

## install-dashboard: Install dashboard dependencies
install-dashboard: venv
	@echo "$(BLUE)Installing dashboard dependencies...$(NC)"
	@$(PIP) install -r python/requirements_app.txt
	@echo "$(GREEN)✅ Dashboard dependencies installed$(NC)"

## load-test: Interactive high-performance load testing scenarios
load-test: venv
	@echo "$(BLUE)Starting High-Performance Load Testing...$(NC)"
	@$(PYTHON) scripts/traffic/high_performance.py

## load-test-quick: Quick safe load test (2 minutes)
load-test-quick: venv
	@echo "$(BLUE)Running Quick Load Test...$(NC)"
	@$(PYTHON) scripts/traffic/high_performance.py \
		--pattern constant \
		--rps 3.0 \
		--concurrent 5 \
		--duration 120 \
		--prompts short

## load-test-queue: Queue stress test for testing queue visualization
load-test-queue: venv
	@echo "$(BLUE)Running Queue Stress Test...$(NC)"
	@echo "$(YELLOW)Watch queue metrics at http://localhost:3001$(NC)"
	@$(PYTHON) scripts/traffic/high_performance.py \
		--pattern constant \
		--rps 25.0 \
		--concurrent 5 \
		--requests 500 \
		--prompts short medium

## load-test-burst: Burst load test with periodic spikes
load-test-burst: venv
	@echo "$(BLUE)Running Burst Load Test...$(NC)"
	@$(PYTHON) scripts/traffic/high_performance.py \
		--pattern burst \
		--rps 20.0 \
		--concurrent 5 \
		--requests 400 \
		--burst-size 50 \
		--burst-interval 10.0 \
		--prompts short medium long

## load-test-chaos: Chaotic random load pattern
load-test-chaos: venv
	@echo "$(BLUE)Running Chaos Load Test...$(NC)"
	@$(PYTHON) scripts/traffic/high_performance.py \
		--pattern chaos \
		--rps 20.0 \
		--concurrent 5 \
		--requests 500 \
		--burst-size 30 \
		--prompts short medium long