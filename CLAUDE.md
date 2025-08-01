# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a comprehensive Ollama monitoring solution featuring:
- **Go services**: Proxy, Dashboard, and Health checker with shared libraries
- **Python traffic generators**: Load testing with 1000+ curated questions across 10 categories
- **Prometheus integration**: Metrics collection and visualization
- **Real-time dashboard**: WebSocket-based monitoring with AI-powered insights

## Architecture

### Service Architecture
- **Go workspace**: Multi-module workspace with shared packages (`go.work`)
- **Services structure**: Proxy, Dashboard, Health checker, and Shared libraries
- **Monitoring flow**: `[Traffic] → [Go Proxy :11435] → [Ollama :11434] → [Metrics :8001]`

### Key Components
1. **Go Services** (`services/`):
   - `proxy/`: Monitoring proxy with request queuing and metrics collection
   - `dashboard/`: Real-time web dashboard with WebSocket updates
   - `health/`: Health checker with AI-powered analysis
   - `shared/`: Common configuration, models, and metrics packages

2. **Python Scripts**: Traffic generation and legacy metrics servers
3. **Configuration**: Centralized YAML-based configuration management
4. **Testing**: Comprehensive test framework with unit, integration, and e2e tests

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

# Run individual services (development mode)
make run-proxy           # Run proxy in foreground
make run-dashboard       # Run dashboard in foreground

# Service management
make start               # Start all services (optimized settings)
make stop                # Stop all services
make restart             # Restart all services
make status              # Check service status
```

### Testing Commands
```bash
# Run all tests
cd test && make test     # Run unit and integration tests
cd test && make test-coverage  # Run with coverage reporting

# Service-specific tests
cd test && make test-proxy     # Test proxy service only
cd test && make test-dashboard # Test dashboard service only

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
make health              # Comprehensive health check with AI analysis
make health-simple       # Quick health check
make metrics             # Show current metrics
make logs                # Tail all service logs
```

## Go Development

### Workspace Structure
- Uses Go 1.24.4 with workspace mode (`go.work`)
- Each service has its own module with standardized Makefile
- Shared packages for common functionality

### Build Process
```bash
# Individual services
cd services/proxy && make build
cd services/dashboard && make build
cd services/health && make build

# All services from root
make build-all
```

### Service Dependencies
- All Go services depend on `services/shared` module
- Shared packages: `config`, `metrics`, `models`
- Use `go mod tidy` in each service directory for dependency management

## Python Environment

**CRITICAL**: All Python scripts require the virtual environment:
```bash
# Always activate first
source venv/bin/activate
python script.py

# Or use venv Python directly
./venv/bin/python script.py
```

## Configuration Management

### Main Configuration
- **Primary config**: `config/llama-metrics.yml` (centralized settings)
- **Service configs**: `config/services/` (service-specific overrides)
- **Prometheus config**: `config/prometheus/prometheus.yml`

### Environment Variables
Key environment variables for optimization:
```bash
OLLAMA_NUM_PARALLEL=2      # Prevent resource contention
MAX_CONCURRENCY=4          # Proxy worker threads
MAX_QUEUE_SIZE=100         # Request buffer size
```

## Port Configuration

- **11434**: Ollama API (default)
- **11435**: Go monitoring proxy
- **8001**: Metrics endpoint
- **3001**: Go dashboard
- **9090**: Prometheus UI
- **8080**: Health checker server

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
3. Add module to `go.work`
4. Use shared packages for common functionality

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
- Run `go mod tidy` in service directories
- Check `go.work` includes all required modules

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