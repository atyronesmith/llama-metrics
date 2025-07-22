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
	@echo "$(BLUE)Ollama Monitoring Stack - Available Targets$(NC)"
	@echo "=============================================="
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
	@echo ""
	@echo "$(YELLOW)Quick Start:$(NC) make setup start traffic"

## setup: Complete setup (venv, dependencies, pull model)
setup: venv install pull-model validate
	@echo "$(GREEN)✅ Setup complete!$(NC)"

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
	@$(PIP) install -r requirements.txt
	@$(PIP) install -r requirements_monitoring.txt
	@$(PIP) install -r requirements_traffic.txt
	@echo "$(GREEN)✅ Dependencies installed$(NC)"

## pull-model: Pull the phi3:mini model for Ollama
pull-model:
	@echo "$(BLUE)Pulling phi3:mini model...$(NC)"
	@ollama pull phi3:mini
	@echo "$(GREEN)✅ Model pulled$(NC)"

## start: Start all monitoring services
start: start-ollama start-proxy start-prometheus
	@echo "$(GREEN)✅ All services started$(NC)"
	@echo "$(BLUE)Metrics available at:$(NC) http://localhost:8001/metrics"
	@echo "$(BLUE)Prometheus UI at:$(NC) http://localhost:9090"

## start-ollama: Start Ollama service
start-ollama:
	@if ! pgrep -x "ollama" > /dev/null; then \
		echo "$(BLUE)Starting Ollama...$(NC)"; \
		ollama serve > ollama.log 2>&1 & \
		sleep 3; \
		echo "$(GREEN)✅ Ollama started$(NC)"; \
	else \
		echo "$(YELLOW)Ollama is already running$(NC)"; \
	fi

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

## start-prometheus: Start Prometheus
start-prometheus:
	@if ! (podman ps 2>/dev/null | grep -q prometheus); then \
		echo "$(BLUE)Starting Prometheus container...$(NC)"; \
		./run_prometheus.sh; \
	else \
		echo "$(YELLOW)Prometheus container is already running$(NC)"; \
	fi

## stop: Stop all monitoring services
stop: stop-proxy stop-prometheus
	@echo "$(GREEN)✅ All monitoring services stopped$(NC)"

## stop-proxy: Stop the monitoring proxy
stop-proxy:
	@echo "$(BLUE)Stopping monitoring proxy...$(NC)"
	@pkill -f "ollama_monitoring_proxy" || true
	@echo "$(GREEN)✅ Monitoring proxy stopped$(NC)"

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
	@tail -f proxy.log prometheus.log ollama.log 2>/dev/null || echo "$(YELLOW)No log files found$(NC)"

## logs-proxy: Tail proxy logs
logs-proxy:
	@tail -f proxy.log proxy_fixed.log 2>/dev/null || echo "$(YELLOW)No proxy logs found$(NC)"

## traffic: Generate traffic (interactive mode)
traffic: venv
	@./run_traffic_generator.sh

## traffic-quick: Quick traffic test (10 requests)
traffic-quick: venv
	@./run_traffic_generator.sh --quick

## traffic-demo: Demo traffic (50 requests)
traffic-demo: venv
	@./run_traffic_generator.sh --demo

## traffic-stress: Stress test (1000 requests)
traffic-stress: venv
	@./run_traffic_generator.sh --stress

## traffic-continuous: Continuous traffic generation
traffic-continuous: venv
	@echo "$(BLUE)Starting continuous traffic generation...$(NC)"
	@./generate_traffic.sh

## metrics: Show current metrics
metrics:
	@echo "$(BLUE)Current Metrics:$(NC)"
	@curl -s http://localhost:8001/metrics | grep -E "^ollama_proxy_requests_total|^ollama_proxy_active_requests" | head -20

## health: Check health of monitoring proxy
health:
	@echo "$(BLUE)Checking monitoring proxy health...$(NC)"
	@curl -s http://localhost:8001/health | jq . || echo "$(RED)❌ Proxy not responding$(NC)"

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
	@if [ -f "prometheus_config.yml" ]; then \
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
	@./run_traffic_generator.sh --quick
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