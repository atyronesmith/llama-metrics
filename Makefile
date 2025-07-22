# Ollama Monitoring Stack Makefile
# Run 'make help' to see all available targets

.PHONY: help setup install clean start stop restart status logs test traffic metrics lint commit push all

# Default target
.DEFAULT_GOAL := help

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
	@echo "$(GREEN)🚦 SERVICE COMMANDS:$(NC)"
	@grep -E '^## (start|stop|restart|status)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(GREEN)📊 MONITORING & TESTING:$(NC)"
	@grep -E '^## (traffic|load-test|dashboard|metrics|logs)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(GREEN)🔧 UTILITIES:$(NC)"
	@grep -E '^## (clean|lint|test|validate)' Makefile | sed 's/## /  /' | column -t -s ':'
	@echo ""
	@echo "$(BLUE)📖 More Info:$(NC) See SETUP.md for detailed installation guide"

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

## install: Install all Python dependencies
install: venv
	@echo "$(BLUE)Installing Python dependencies...$(NC)"
	@$(PIP) install --upgrade pip
	@$(PIP) install -r requirements_all.txt
	@echo "$(GREEN)✅ All dependencies installed$(NC)"

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
start: start-ollama start-proxy start-litellm start-portkey start-prometheus start-dashboard
	@echo "$(GREEN)🚀 All services started successfully!$(NC)"
	@echo ""
	@echo "$(BLUE)📊 Dashboard:$(NC)        http://localhost:3001"
	@echo "$(BLUE)📈 Prometheus UI:$(NC)    http://localhost:9090"
	@echo "$(BLUE)🔧 Metrics API:$(NC)      http://localhost:8001/metrics"
	@echo "$(BLUE)🤖 Ollama API:$(NC)       http://localhost:11434"
	@echo "$(BLUE)🔄 LiteLLM Proxy:$(NC)    http://localhost:8000"
	@echo "$(BLUE)🚪 Portkey Gateway:$(NC)  http://localhost:8787"
	@echo ""
	@echo "$(YELLOW)💡 Pro tip: Run 'make traffic' to generate test traffic$(NC)"

## start-with-portkey: Start all services with Portkey-enabled monitoring proxy
start-with-portkey: start-ollama start-proxy-portkey start-litellm start-portkey start-prometheus start-dashboard
	@echo "$(GREEN)🚀 All services with Portkey integration started!$(NC)"
	@echo ""
	@echo "$(BLUE)📊 Dashboard:$(NC)        http://localhost:3001"
	@echo "$(BLUE)📈 Prometheus UI:$(NC)    http://localhost:9090"
	@echo "$(BLUE)🔧 Metrics API:$(NC)      http://localhost:8001/metrics"
	@echo "$(BLUE)🤖 Ollama API:$(NC)       http://localhost:11434"
	@echo "$(BLUE)🔄 LiteLLM Proxy:$(NC)    http://localhost:8000"
	@echo "$(BLUE)🚪 Portkey Gateway:$(NC)  http://localhost:8787"
	@echo ""
	@echo "$(YELLOW)🌟 Portkey routing enabled! Traffic to proxy (11435) routes through Portkey$(NC)"
	@echo "$(YELLOW)💡 Run 'make traffic' to test the integrated stack$(NC)"

## start-ollama: Start Ollama service and monitoring instance
start-ollama:
	@if ! pgrep -x "ollama" > /dev/null; then \
		echo "$(BLUE)Starting Ollama...$(NC)"; \
		ollama serve > ollama.log 2>&1 & \
		sleep 3; \
		echo "$(GREEN)✅ Ollama started$(NC)"; \
	else \
		echo "$(YELLOW)Ollama is already running$(NC)"; \
	fi
	@echo "$(BLUE)Starting dedicated monitoring Ollama instance...$(NC)"
	@bash scripts/start_monitoring_ollama.sh

## start-proxy: Start the monitoring proxy
start-proxy: venv
	@if ! pgrep -f "ollama_monitoring_proxy_fixed.py" > /dev/null; then \
		echo "$(BLUE)Starting monitoring proxy...$(NC)"; \
		$(PYTHON) ollama_monitoring_proxy_fixed.py > proxy.log 2>&1 & \
		sleep 2; \
		echo "$(GREEN)✅ Monitoring proxy started$(NC)"; \
	else \
		echo "$(YELLOW)Monitoring proxy is already running$(NC)"; \
	fi

## start-proxy-portkey: Start the monitoring proxy with Portkey routing
start-proxy-portkey: venv
	@if ! pgrep -f "ollama_monitoring_proxy_fixed.py" > /dev/null; then \
		echo "$(BLUE)Starting monitoring proxy with Portkey routing...$(NC)"; \
		$(PYTHON) ollama_monitoring_proxy_fixed.py --enable-portkey > proxy_portkey.log 2>&1 & \
		sleep 2; \
		echo "$(GREEN)✅ Monitoring proxy with Portkey routing started$(NC)"; \
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
stop: stop-proxy stop-litellm stop-portkey stop-monitoring-ollama stop-prometheus
	@echo "$(GREEN)✅ All monitoring services stopped$(NC)"

## stop-proxy: Stop the monitoring proxy
stop-proxy:
	@echo "$(BLUE)Stopping monitoring proxy...$(NC)"
	@pkill -f "ollama_monitoring_proxy" || true
	@echo "$(GREEN)✅ Monitoring proxy stopped$(NC)"

## stop-monitoring-ollama: Stop the dedicated monitoring Ollama instance
stop-monitoring-ollama:
	@echo "$(BLUE)Stopping monitoring Ollama instance...$(NC)"
	@lsof -ti:11435 | xargs kill -9 2>/dev/null || true
	@echo "$(GREEN)✅ Monitoring Ollama instance stopped$(NC)"

## start-litellm: Start LiteLLM proxy with priority queues
start-litellm: venv
	@if ! lsof -ti:8000 > /dev/null 2>&1; then \
		echo "$(BLUE)Starting LiteLLM proxy...$(NC)"; \
		bash scripts/start_litellm_proxy.sh; \
	else \
		echo "$(YELLOW)LiteLLM proxy is already running on port 8000$(NC)"; \
	fi

## stop-litellm: Stop LiteLLM proxy
stop-litellm:
	@echo "$(BLUE)Stopping LiteLLM proxy...$(NC)"
	@lsof -ti:8000 | xargs kill -9 2>/dev/null || true
	@pkill -f "litellm" || true
	@echo "$(GREEN)✅ LiteLLM proxy stopped$(NC)"

## start-portkey: Start Portkey Gateway with Docker Compose
start-portkey:
	@if ! lsof -ti:8787 > /dev/null 2>&1; then \
		echo "$(BLUE)Starting Portkey Gateway...$(NC)"; \
		if command -v podman-compose >/dev/null 2>&1; then \
			podman-compose -f portkey-compose.yaml up -d; \
		elif command -v docker-compose >/dev/null 2>&1; then \
			docker-compose -f portkey-compose.yaml up -d; \
		elif command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
			docker compose -f portkey-compose.yaml up -d; \
		else \
			echo "$(RED)❌ Neither Docker nor Podman Compose found$(NC)"; \
			exit 1; \
		fi; \
		echo "$(GREEN)✅ Portkey Gateway started at http://localhost:8787$(NC)"; \
	else \
		echo "$(YELLOW)Portkey Gateway is already running on port 8787$(NC)"; \
	fi

## stop-portkey: Stop Portkey Gateway
stop-portkey:
	@echo "$(BLUE)Stopping Portkey Gateway...$(NC)"
	@if command -v podman-compose >/dev/null 2>&1; then \
		podman-compose -f portkey-compose.yaml down 2>/dev/null || true; \
	elif command -v docker-compose >/dev/null 2>&1; then \
		docker-compose -f portkey-compose.yaml down 2>/dev/null || true; \
	elif command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then \
		docker compose -f portkey-compose.yaml down 2>/dev/null || true; \
	fi
	@lsof -ti:8787 | xargs kill -9 2>/dev/null || true
	@echo "$(GREEN)✅ Portkey Gateway stopped$(NC)"

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
	@if pgrep -f "ollama_monitoring_proxy" > /dev/null; then \
		echo "$(GREEN)✅ Monitoring Proxy: Running$(NC)"; \
	else \
		echo "$(RED)❌ Monitoring Proxy: Not running$(NC)"; \
	fi
	@if lsof -ti:8000 > /dev/null 2>&1; then \
		echo "$(GREEN)✅ LiteLLM Proxy: Running$(NC)"; \
	else \
		echo "$(RED)❌ LiteLLM Proxy: Not running$(NC)"; \
	fi
	@if lsof -ti:8787 > /dev/null 2>&1; then \
		echo "$(GREEN)✅ Portkey Gateway: Running$(NC)"; \
	else \
		echo "$(RED)❌ Portkey Gateway: Not running$(NC)"; \
	fi
	@if pgrep -x "prometheus" > /dev/null; then \
		echo "$(GREEN)✅ Prometheus: Running (native)$(NC)"; \
	elif (podman ps 2>/dev/null || docker ps 2>/dev/null) | grep -q prometheus; then \
		echo "$(GREEN)✅ Prometheus: Running (container)$(NC)"; \
	else \
		echo "$(RED)❌ Prometheus: Not running$(NC)"; \
	fi

## logs: Tail all service logs
logs:
	@echo "$(BLUE)Tailing logs (Ctrl+C to stop)...$(NC)"
	@tail -f proxy.log prometheus.log ollama.log litellm.log 2>/dev/null || echo "$(YELLOW)No log files found$(NC)"

## logs-litellm: Tail LiteLLM proxy logs
logs-litellm:
	@tail -f litellm.log 2>/dev/null || echo "$(YELLOW)No LiteLLM logs found$(NC)"

## logs-proxy: Tail proxy logs
logs-proxy:
	@tail -f proxy.log proxy_fixed.log 2>/dev/null || echo "$(YELLOW)No proxy logs found$(NC)"

## traffic: Generate traffic (interactive mode)
traffic: venv
	@./scripts/run_traffic_generator.sh

## traffic-quick: Quick traffic test (10 requests)
traffic-quick: venv
	@./scripts/run_traffic_generator.sh --quick

## traffic-demo: Demo traffic (50 requests)
traffic-demo: venv
	@./scripts/run_traffic_generator.sh --demo

## traffic-stress: Stress test (1000 requests)
traffic-stress: venv
	@./scripts/run_traffic_generator.sh --stress

## traffic-continuous: Continuous traffic generation
traffic-continuous: venv
	@echo "$(BLUE)Starting continuous traffic generation...$(NC)"
	@./generate_traffic.sh

## traffic-portkey: Generate mixed traffic through Portkey Gateway and monitoring proxy
traffic-portkey: venv
	@echo "$(BLUE)Generating mixed Portkey traffic...$(NC)"
	@$(PYTHON) portkey_traffic_generator.py --requests 20 --delay 2.0 --mode mixed

## traffic-portkey-direct: Generate traffic directly to Portkey Gateway
traffic-portkey-direct: venv
	@echo "$(BLUE)Generating direct traffic to Portkey Gateway...$(NC)"
	@$(PYTHON) portkey_traffic_generator.py --requests 15 --delay 1.5 --mode portkey

## traffic-portkey-proxy: Generate traffic through monitoring proxy to Portkey
traffic-portkey-proxy: venv
	@echo "$(BLUE)Generating traffic through monitoring proxy to Portkey...$(NC)"
	@$(PYTHON) portkey_traffic_generator.py --requests 15 --delay 1.5 --mode proxy

## metrics: Show current metrics
metrics:
	@echo "$(BLUE)Current Metrics:$(NC)"
	@curl -s http://localhost:8001/metrics | grep -E "^ollama_proxy_requests_total|^ollama_proxy_active_requests" | head -20

## health: Check health of monitoring proxy
health:
	@echo "$(BLUE)Checking monitoring proxy health...$(NC)"
	@curl -s http://localhost:8001/health | jq . || echo "$(RED)❌ Proxy not responding$(NC)"

## health-litellm: Check LiteLLM proxy health
health-litellm:
	@echo "$(BLUE)Checking LiteLLM proxy health...$(NC)"
	@curl -s http://localhost:8000/health | jq . || echo "$(RED)❌ LiteLLM proxy not responding$(NC)"

## health-portkey: Check Portkey Gateway health
health-portkey:
	@echo "$(BLUE)Checking Portkey Gateway health...$(NC)"
	@curl -s http://localhost:8787/health | jq . || echo "$(RED)❌ Portkey Gateway not responding$(NC)"

## prometheus-ui: Open Prometheus UI in browser
prometheus-ui:
	@echo "$(BLUE)Opening Prometheus UI...$(NC)"
	@open http://localhost:9090 || xdg-open http://localhost:9090 || echo "$(YELLOW)Please open http://localhost:9090 in your browser$(NC)"

## test: Run monitoring tests
test: venv
	@echo "$(BLUE)Running monitoring tests...$(NC)"
	@$(PYTHON) test_ollama_monitoring.py

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
	@if [ -f "docs/prometheus_config.yml" ]; then \
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
clean:
	@echo "$(BLUE)Cleaning up...$(NC)"
	@rm -f *.log
	@rm -f monitoring_pids.txt
	@rm -rf __pycache__
	@find . -name "*.pyc" -delete
	@echo "$(GREEN)✅ Cleanup complete$(NC)"

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
	@./run_prometheus.sh

## all: Complete setup, start services, and run demo
all: setup start demo

# Advanced targets for development

## debug-proxy: Run proxy in foreground for debugging
debug-proxy: venv
	@echo "$(BLUE)Running proxy in debug mode...$(NC)"
	@$(PYTHON) ollama_monitoring_proxy_fixed.py

## watch-metrics: Continuously watch metrics
watch-metrics:
	@watch -n 2 'curl -s http://localhost:8001/metrics | grep -E "^ollama_proxy" | head -20'

## benchmark: Run performance benchmark
benchmark: venv
	@echo "$(BLUE)Running performance benchmark...$(NC)"
	@$(PYTHON) -c "print('Starting benchmark with 100 requests...')"
	@./scripts/run_traffic_generator.sh --quick
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
dashboard: venv
	@echo "$(BLUE)Starting Ollama Dashboard...$(NC)"
	@$(PYTHON) dashboard.py

## start-dashboard: Start dashboard in background
start-dashboard: venv
	@if ! pgrep -f "dashboard.py" > /dev/null; then \
		echo "$(BLUE)Starting dashboard in background...$(NC)"; \
		nohup $(PYTHON) dashboard.py > dashboard.log 2>&1 & \
		sleep 2; \
		echo "$(GREEN)✅ Dashboard started at http://localhost:3001$(NC)"; \
	else \
		echo "$(YELLOW)Dashboard is already running$(NC)"; \
	fi

## stop-dashboard: Stop dashboard
stop-dashboard:
	@echo "$(BLUE)Stopping dashboard...$(NC)"
	@pkill -f "dashboard.py" || true
	@echo "$(GREEN)✅ Dashboard stopped$(NC)"

## install-dashboard: Install dashboard dependencies
install-dashboard: venv
	@echo "$(BLUE)Installing dashboard dependencies...$(NC)"
	@$(PIP) install -r requirements_dashboard.txt
	@echo "$(GREEN)✅ Dashboard dependencies installed$(NC)"

## load-test: Interactive high-performance load testing scenarios
load-test: venv
	@echo "$(BLUE)Starting High-Performance Load Testing...$(NC)"
	@./scripts/load_test_scenarios.sh

## load-test-quick: Quick safe load test (2 minutes)
load-test-quick: venv
	@echo "$(BLUE)Running Quick Load Test...$(NC)"
	@$(PYTHON) scripts/high_performance_load_tester.py \
		--pattern constant \
		--rps 3.0 \
		--concurrent 5 \
		--duration 120 \
		--prompts short

## load-test-queue: Queue stress test for testing queue visualization
load-test-queue: venv
	@echo "$(BLUE)Running Queue Stress Test...$(NC)"
	@echo "$(YELLOW)Watch queue metrics at http://localhost:3001$(NC)"
	@$(PYTHON) scripts/high_performance_load_tester.py \
		--pattern constant \
		--rps 25.0 \
		--concurrent 5 \
		--requests 500 \
		--prompts short medium

## load-test-burst: Burst load test with periodic spikes
load-test-burst: venv
	@echo "$(BLUE)Running Burst Load Test...$(NC)"
	@$(PYTHON) scripts/high_performance_load_tester.py \
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
	@$(PYTHON) scripts/high_performance_load_tester.py \
		--pattern chaos \
		--rps 20.0 \
		--concurrent 5 \
		--requests 500 \
		--burst-size 30 \
		--prompts short medium long