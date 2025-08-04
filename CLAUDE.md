# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a comprehensive Ollama monitoring solution featuring:
- **Go services**: Proxy, Dashboard, Health checker, and Mac metrics service with shared libraries
- **Python traffic generators**: Load testing with 1000+ curated questions across 10 categories
- **Prometheus integration**: Metrics collection and visualization
- **Real-time dashboard**: WebSocket-based monitoring with AI-powered insights

## Architecture

### Service Architecture
- **Go workspace**: Multi-module workspace with shared packages (`services/go.work`)
- **Services structure**: Proxy, Dashboard, Health checker, Mac metrics service, and Shared libraries
- **Python components**: Organized in `python/` directory with shared utilities and containerization
- **Monitoring flow**: `[Traffic] → [Go Proxy :11435] → [Ollama :11434] → [Metrics :8001]`

### Key Components
1. **Go Services** (`services/`):
   - `proxy/`: Monitoring proxy with request queuing and metrics collection
   - `dashboard/`: Real-time web dashboard with WebSocket updates
   - `health/`: Health checker with AI-powered analysis
   - `mac-metrics/`: Native Go Mac system metrics collector (replaces Python version)
   - `shared/`: Common configuration, models, and metrics packages

2. **Python Components** (`python/`):
   - `shared/`: Common utilities (config_manager.py, version.py)
   - `container/`: Docker containerization files
   - `requirements*.txt`: Dependency specifications for different use cases

3. **Script Libraries** (`scripts/`):
   - `traffic/`: Load testing and traffic generation scripts
   - `monitoring/`: System monitoring utilities
   - `deployment/`: Installation and deployment scripts

4. **Configuration**: Centralized YAML-based configuration management
5. **Testing**: Comprehensive test framework with unit, integration, and e2e tests

## Current Status (August 2025)

### ✅ Fully Operational Stack
All monitoring services are now running successfully with complete Prometheus integration:

**Active Services:**
- **Ollama**: LLM service running on port 11434
- **Go Monitoring Proxy**: Request interception and metrics collection on port 11435 (metrics: 8001)
- **Go Dashboard**: Real-time WebSocket dashboard on port 3001 with Prometheus button
- **Go Health Checker**: AI-powered health analysis on port 8080 (server mode)
- **Go Mac Metrics**: Native Go Mac system metrics collection on port 8002
- **Prometheus**: Containerized metrics storage and querying on port 9090

**Recent Achievements:**
- ✅ **Converted Python to Go**: Replaced Python mac_metrics.py with native Go service
- ✅ **Enhanced Mac metrics collection**: Native Go powermetrics integration with proper error handling
- ✅ **Eliminated dependency issues**: No more Python/sudo environment conflicts
- ✅ **Fixed JSON parsing errors**: Clean proxy logs with proper Go service integration
- ✅ **Improved dashboard layout**: Reorganized metrics sections for better visual grouping
- ✅ **Enhanced UI organization**: Moved service status to navbar, grouped related metrics
- ✅ **Resolved all Prometheus scraping errors**: 6/6 targets healthy including new Go service
- ✅ **Fixed GPU metrics collection**: Replaced faulty JSON parsing with text parsing for powermetrics
- ✅ **Integrated dashboard with Mac metrics**: Dashboard now displays real-time GPU utilization and power data
- ✅ **Cleaned up deprecated Python code**: Removed unused monitoring scripts and shared utilities

**Prometheus Metrics Coverage:**
- **Proxy Service**: Request rates, latencies, token generation, system resources
- **Health Checker**: Service status, response times, system health metrics
- **Go Mac Metrics**: Real-time GPU utilization, CPU/GPU power consumption, memory pressure, thermal state (native Go with text parsing)
- **Dashboard**: Go runtime metrics (GC, memory, goroutines)
- **Prometheus**: Self-monitoring metrics

## Essential Commands

### Quick Start
```bash
make setup start          # Complete setup and start all services
make traffic              # Generate test traffic
```

### Development Workflow
```bash
# Build all Go services
make build                # Build proxy and dashboard
make build-all           # Build for multiple platforms
make build-health        # Build health checker
make build-mac-metrics   # Build Go Mac metrics service

# Run individual services (development mode)
make run-proxy           # Run proxy in foreground
make run-dashboard       # Run dashboard in foreground

# Service management (starts all 6 services)
make start               # Start all services: ollama, proxy, prometheus, dashboard, health, go-mac-metrics
make stop                # Stop all services
make restart             # Restart all services
make status              # Check service status (shows all 6 services)

# Individual service control
make start-health        # Start health checker server (port 8080)
make start-mac-metrics   # Start Go Mac metrics server (port 8002)
make stop-health         # Stop health checker
make stop-mac-metrics    # Stop Mac metrics
```

### Testing Commands
```bash
# Run all tests
cd test && make test     # Run unit and integration tests
cd test && make test-coverage  # Run with coverage reporting

# Service-specific tests
cd test && make test-proxy     # Test proxy service only
cd test && make test-dashboard # Test dashboard service only
cd services/mac-metrics && make test # Test Go Mac metrics service

# Test specific functionality
cd test && make test-specific TEST=TestName MODULE=proxy
```

### Traffic Generation
```bash
make traffic-quick       # Quick test (10 requests)
make traffic-demo        # Demo mode (50 requests) 
make traffic-stress      # Stress test (1000 requests)
make load-test-queue     # Queue stress testing
```

### Health and Monitoring
```bash
# Health checks (requires health service running)
make health              # Comprehensive health check with AI analysis
make health-simple       # Quick health check
make health-readiness    # Kubernetes-style readiness check
make health-liveness     # Kubernetes-style liveness check
make health-analyzed     # AI-powered health analysis with recommendations

# Monitoring and metrics
make metrics             # Show current metrics from proxy
make logs                # Tail all service logs (includes health.log, mac_metrics.log)
make logs-proxy          # Proxy-specific logs

# Prometheus and visualization
# Dashboard now includes Prometheus button for direct access to http://localhost:9090
```

## Go Development

### Workspace Structure
- Uses Go 1.24.4 with workspace mode (`services/go.work`)
- Workspace files located in `services/` directory to contain all Go-related files
- Each service has its own module with standardized Makefile
- Shared packages for common functionality

### Go Mac Metrics Service
The native Go Mac metrics service (`services/mac-metrics/`) provides:
- **macOS system metrics**: GPU utilization, CPU/GPU power, memory pressure, thermal state
- **Prometheus integration**: Native metrics endpoint at `/metrics`
- **JSON compatibility**: Legacy `/metrics/json` endpoint for existing integrations
- **Health monitoring**: Service health endpoint at `/health`
- **Graceful degradation**: Works without sudo (memory pressure only) or with sudo (full metrics)
- **Platform detection**: macOS-only with proper build constraints (`//go:build darwin`)

**Key features:**
- No Python dependencies (replaces `scripts/monitoring/mac_metrics.py`)
- **Text-based powermetrics parsing**: Fixed GPU metrics collection with regex-based parsing
- **Real-time GPU monitoring**: Accurate GPU utilization and power consumption tracking
- Better error handling for powermetrics failures
- Native Prometheus metrics without conversion
- Integrated health checks and service information

### Build Process
```bash
# Individual services (from project root)
cd services/proxy && make build
cd services/dashboard && make build
cd services/health && make build
cd services/mac-metrics && make build

# All services from root (recommended)
make build              # Build proxy and dashboard
make build-all          # Build for multiple platforms
make build-mac-metrics  # Build Go Mac metrics service
```

### Service Dependencies
- All Go services depend on `services/shared` module
- Shared packages: `config`, `metrics`, `models`
- Workspace manages dependencies across all modules automatically
- Use `go work sync` from `services/` directory to update workspace dependencies

## Python Environment

**CRITICAL**: All Python scripts require the virtual environment:
```bash
# Always activate first
source venv/bin/activate
python script.py

# Or use venv Python directly
./venv/bin/python script.py
```

### Python Package Structure
- **Dependencies**: Located in `python/requirements*.txt`
  - `requirements.txt`: Minimal dependencies for traffic generation
  - `requirements_all.txt`: Complete dependency set
  - `requirements_app.txt`: Dependencies for advanced features
  - `requirements_traffic.txt`: Traffic generation specific dependencies
- **Shared utilities**: `python/shared/` contains reusable Python modules
- **Container support**: `python/container/Dockerfile` for containerized deployments

## Configuration Management

### Main Configuration
- **Primary config**: `config/llama-metrics.yml` (centralized settings)
- **Service configs**: `config/services/` (service-specific overrides)
- **Prometheus config**: `config/prometheus/prometheus.yml`

### Remote Ollama Server Support
The application supports connecting to remote Ollama servers:

**Configuration file**: Edit `config/llama-metrics.yml`:
```yaml
server:
  ollama_url: "http://192.168.1.100:11434"  # Remote server
models:
  default_model: "llama3:latest"             # Available on remote server
```

**Environment variables**: For runtime configuration:
```bash
export OLLAMA_URL="http://ollama-server.local:11434"
export DEFAULT_MODEL="llama3:latest"
```

### Environment Variables
Key environment variables for configuration:
```bash
# Ollama Configuration
OLLAMA_URL=http://192.168.1.100:11434  # Remote Ollama server
DEFAULT_MODEL=llama3:latest            # Default model for traffic generation

# Performance Optimization
OLLAMA_NUM_PARALLEL=2                  # Prevent resource contention
MAX_CONCURRENCY=4                      # Proxy worker threads
MAX_QUEUE_SIZE=100                     # Request buffer size
```

## Port Configuration

- **11434**: Ollama API (default)
- **11435**: Go monitoring proxy
- **8001**: Proxy metrics endpoint
- **3001**: Go dashboard (now with Prometheus UI button)
- **9090**: Prometheus UI (container)
- **8080**: Health checker server (comprehensive health monitoring)
- **8002**: Mac system metrics server (GPU, power, thermal)

## Testing Strategy

### Test Structure
- **Unit tests**: Individual service testing
- **Integration tests**: Cross-service functionality
- **E2E tests**: Full system testing
- **Load tests**: Performance and stress testing

### Running Tests
```bash
# From test directory
make test                # All tests
make test-race          # Race condition detection
make test-coverage      # Coverage reporting
make test-ci            # CI-friendly test run
```

## Development Best Practices

### Go Services
- Follow the Makefile template in `services/Makefile.template`
- Use shared packages for common functionality
- Implement proper error handling and logging
- Add comprehensive tests for new features

### Python Scripts
- Always use the virtual environment
- Follow existing logging patterns
- Use structured configuration via YAML files

### Shell Scripts
- Must pass `shellcheck` validation
- Use proper quoting and error handling
- Include descriptive help messages

## Common Development Tasks

### Adding New Go Service
1. Create service directory under `services/`
2. Copy and customize `services/Makefile.template`
3. Add module to `services/go.work`
4. Use shared packages for common functionality
5. Run `go work sync` from `services/` directory to update workspace

Example: The `mac-metrics` service was created following this pattern:
- Created `services/mac-metrics/` with standard Go project structure
- Added build targets to main Makefile (`build-mac-metrics`, `start-mac-metrics`)
- Integrated with Prometheus and existing monitoring infrastructure

### Adding New Metrics
1. Define metrics in `services/shared/metrics/`
2. Implement collection in relevant service
3. Update dashboard visualization if needed
4. Add tests for new metrics

### Debugging Services
```bash
# Run services in foreground for debugging
make run-proxy           # Proxy with debug output
make run-dashboard       # Dashboard with debug output
make debug-proxy         # Python proxy debug mode

# Check service health
make health-analyzed     # AI-powered health analysis
make logs-proxy          # Proxy-specific logs
```

## Troubleshooting

### Build Issues
- Ensure Go 1.24.4+ is installed
- Run `go work sync` from `services/` directory to sync workspace
- Check `services/go.work` includes all required modules
- Individual service `go mod tidy` should rarely be needed due to workspace

### Service Issues
- Use `make status` to check service states
- Check logs with `make logs` or `make logs-proxy`
- Verify ports aren't conflicting with `lsof -i`

### Python Issues
- Activate virtual environment: `source venv/bin/activate`
- Reinstall dependencies: `make install`
- Check Python version compatibility

## Performance Optimization

### Ollama Configuration
- Set `OLLAMA_NUM_PARALLEL=2` to prevent resource contention
- Use `OLLAMA_MAX_LOADED_MODELS=2` for memory management

### Proxy Configuration
- Max concurrency of 4 matches Ollama capabilities
- Queue size of 100 provides adequate buffering
- Monitor queue metrics for optimal sizing

### Load Testing
- Use appropriate test patterns for different scenarios
- Monitor system resources during load tests
- Adjust concurrency based on system capabilities

## Data Management

### Prometheus Data
- **Data directory**: `prometheus-data/` contains runtime Prometheus data
- **Not versioned**: Prometheus data files are in `.gitignore` and should never be committed
- **Structure**: Directory structure should exist but data files (`prometheus-data/wal/*`, `prometheus-data/queries.active`) are generated at runtime
- **Cleanup**: Use `make clean` to remove logs and temporary files (data directory preserved)