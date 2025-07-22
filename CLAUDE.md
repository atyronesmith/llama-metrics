# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a comprehensive monitoring solution for Ollama AI models using Prometheus metrics. The stack includes:
- A monitoring proxy that intercepts Ollama API requests to collect metrics
- Traffic generators for load testing and continuous monitoring
- Prometheus integration for metrics collection and visualization
- 1000+ curated test questions across 10 categories

## Key Architecture

The monitoring flow: `[Traffic Generator] → [Monitoring Proxy :11435] → [Ollama API :11434]`
- Proxy intercepts all requests, collects metrics, forwards to Ollama
- Metrics exposed on port 8001 for Prometheus scraping
- Traffic generators send diverse questions for realistic load testing

## Essential Commands

### Setup and Installation
```bash
make setup          # Complete setup (venv, dependencies, model)
make install        # Install Python dependencies only
make pull-model     # Pull phi3:mini model for Ollama
```

### Running Services
```bash
make start          # Start all services (Ollama, proxy, Prometheus)
make stop           # Stop monitoring services
make restart        # Restart all services
make status         # Check service status
```

### Traffic Generation
```bash
make traffic        # Interactive traffic mode
make traffic-quick  # Quick test (10 requests)
make traffic-demo   # Demo mode (50 requests)
make traffic-stress # Stress test (1000 requests)
```

### Development
```bash
make test           # Run monitoring tests
make lint           # Run shellcheck on shell scripts
make metrics        # Show current metrics
make logs           # Tail all service logs
```

### Dashboard
```bash
make dashboard      # Start dashboard (foreground)
make start-dashboard # Start dashboard (background)
make stop-dashboard # Stop dashboard
make install-dashboard # Install dashboard dependencies
```

## Python Virtual Environment Requirement

**CRITICAL**: All Python scripts MUST run within the virtual environment at `./venv/`:

```bash
# Activate venv first
source venv/bin/activate
python script.py

# OR use venv Python directly
./venv/bin/python script.py
```

## Core Components

1. **ollama_monitoring_proxy_fixed.py** - Main monitoring proxy
   - Runs on ports 11435 (proxy) and 8001 (metrics)
   - Collects comprehensive metrics: latency, tokens/sec, errors, system stats
   - Must run in venv: `./venv/bin/python ollama_monitoring_proxy_fixed.py`

2. **traffic_generator.py** - Load testing tool
   - Loads questions from `questions/` directory
   - Supports various modes and configurable parameters
   - Must run in venv: `./venv/bin/python traffic_generator.py`

3. **test_ollama_monitoring.py** - Test suite
   - Tests proxy functionality, streaming, metrics collection
   - Must run in venv: `./venv/bin/python test_ollama_monitoring.py`

4. **dashboard.py** - Real-time web dashboard
   - Comprehensive LLM performance visualization
   - Real-time metrics with WebSocket updates
   - Token generation, memory, GPU, and power graphs
   - Must run in venv: `./venv/bin/python dashboard.py`

5. **portkey_traffic_generator.py** - Portkey-specific traffic generator
   - Generates mixed traffic through Portkey Gateway and monitoring proxy
   - Supports direct Portkey, proxy-routed, and mixed modes
   - Must run in venv: `./venv/bin/python portkey_traffic_generator.py`

## Portkey AI Gateway Integration

The monitoring stack now includes comprehensive Portkey AI Gateway integration for advanced LLM routing, caching, and observability.

### Architecture with Portkey

```
[Traffic Generator] → [Monitoring Proxy :11435] → [Portkey Gateway :8787] → [Ollama API :11434]
                                  ↓
                        [Prometheus Metrics :8001] → [Dashboard :3001]
```

### Portkey Features Supported

- **Smart Routing**: Route requests through Portkey for load balancing and fallbacks
- **Caching**: Leverage Portkey's semantic caching capabilities
- **Observability**: Dual metrics collection from both proxy and Portkey
- **Configuration**: JSON-based Portkey configuration for model routing

### Portkey Configuration Files

- `portkey-config.json` - Portkey Gateway configuration for Ollama integration
- `portkey-compose.yaml` - Docker Compose configuration for Portkey Gateway
- `portkey-gateway/` - Complete Portkey Gateway codebase with plugins and providers

### Portkey Commands

#### Service Management
```bash
make start-portkey       # Start Portkey Gateway with Docker Compose
make stop-portkey        # Stop Portkey Gateway
make health-portkey      # Check Portkey Gateway health
make start-with-portkey  # Start all services with Portkey integration enabled
```

#### Traffic Generation
```bash
make traffic-portkey         # Mixed traffic through both Portkey and proxy
make traffic-portkey-direct  # Direct traffic to Portkey Gateway
make traffic-portkey-proxy   # Traffic through monitoring proxy to Portkey
```

#### Proxy Modes
```bash
make start-proxy         # Standard proxy (direct to Ollama)
make start-proxy-portkey # Proxy with Portkey routing enabled
```

### Portkey Routing Metrics

The dashboard now includes dedicated Portkey routing metrics:

- **Portkey Requests**: Total requests routed through Portkey
- **Direct Requests**: Total requests sent directly to Ollama
- **Routing Ratio**: Percentage of traffic routed through Portkey
- **Routing Latency**: Latency comparison between routing methods

### Monitoring Proxy Portkey Integration

The monitoring proxy supports Portkey routing with command-line options:

```bash
# Enable Portkey routing
./venv/bin/python ollama_monitoring_proxy_fixed.py --enable-portkey

# Custom Portkey configuration
./venv/bin/python ollama_monitoring_proxy_fixed.py \
  --enable-portkey \
  --portkey-host localhost \
  --portkey-port 8787
```

All metrics include a `routing` label with values `portkey` or `direct` for granular monitoring.

## Port Configuration

- 11434: Ollama API (default)
- 11435: Monitoring proxy 
- 8787: Portkey AI Gateway
- 8001: Prometheus metrics endpoint
- 9090: Prometheus UI
- 3001: Web dashboard
- 8000: Alternative metrics server (app.py or enhanced_metrics_server.py)

## Testing Approach

Use the Makefile for all testing:
```bash
make test           # Run Python test suite
make lint           # Validate shell scripts with shellcheck
make validate       # Check setup and configuration
```

### Testing Portkey Integration

```bash
# Test complete Portkey stack
make start-with-portkey     # Start all services with Portkey enabled
make traffic-portkey        # Generate mixed Portkey traffic
make health-portkey         # Verify Portkey Gateway health
make metrics               # View routing metrics

# Test individual components
make traffic-portkey-direct # Test direct Portkey routing
make traffic-portkey-proxy  # Test proxy-to-Portkey routing
make stop-portkey          # Clean shutdown
```

## Common Issues

1. **ModuleNotFoundError**: Not using virtual environment
   - Fix: `source venv/bin/activate` or use `./venv/bin/python`

2. **Port conflicts**: Service already running
   - Fix: `make stop` then `make start`

3. **Ollama connection errors**: Ollama not running
   - Fix: `ollama serve` or `make start-ollama`

## Monitoring Best Practices

- Always run traffic through the monitoring proxy (port 11435) not directly to Ollama
- Use `make status` to verify all services are running before testing
- Check metrics at http://localhost:8001/metrics
- View Prometheus dashboard at http://localhost:9090

## Shell Script Standards

All shell scripts must pass shellcheck validation:
```bash
shellcheck *.sh
```

Scripts use proper quoting, error handling, and portable syntax.