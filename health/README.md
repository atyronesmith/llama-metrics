# Health Checker

A comprehensive health checking service for the Ollama Monitoring Stack, written in Go.

## Features

- **Service Health Checks**: Monitors the health of Ollama, proxy, metrics, and dashboard services
- **System Metrics**: Collects CPU, memory, disk, and network statistics
- **macOS Specific Metrics**: GPU and power information on macOS
- **Multiple Check Types**:
  - Comprehensive: Full system and service health check
  - Simple: Quick system status check
  - Readiness: Kubernetes-style readiness probe
  - Liveness: Kubernetes-style liveness probe
  - **Analyzed**: AI-powered analysis with insights and recommendations
- **Ollama Generation Test**: Verifies Ollama can actually generate responses
- **LLM Health Analysis**: Uses Ollama to analyze health data and provide:
  - Root cause analysis for failures
  - Performance optimization suggestions
  - Actionable recommendations
- **HTTP Server Mode**: Provides health endpoints for monitoring tools
- **CLI Mode**: Command-line interface for one-off checks

## Usage

### CLI Mode

```bash
# Build the health checker
make build

# Run comprehensive health check
./build/healthcheck -mode cli -check comprehensive

# Run simple health check
./build/healthcheck -mode cli -check simple

# Check readiness
./build/healthcheck -mode cli -check readiness

# Check liveness
./build/healthcheck -mode cli -check liveness

# Run health check with AI analysis
./build/healthcheck -mode cli -check analyzed
```

### Server Mode

```bash
# Start health check server on default port (8080)
./build/healthcheck -mode server

# Start on custom port
./build/healthcheck -mode server -port 9090
```

### Make Targets

```bash
# Run comprehensive health check
make check

# Run simple health check
make check-simple

# Run readiness check
make check-readiness

# Run liveness check
make check-liveness

# Run health check with AI analysis
make check-analyzed

# Start health server
make run
```

## API Endpoints (Server Mode)

- `GET /health` - Comprehensive health check
- `GET /health/simple` - Simple health check
- `GET /health/analyzed` - Health check with AI-powered analysis
- `GET /readiness` - Readiness probe
- `GET /liveness` - Liveness probe
- `GET /api/health` - Legacy endpoint (same as /health)

## Configuration

The health checker reads configuration from `config.yml` in the parent directory. It uses this to determine:
- Service endpoints to check
- Default model for Ollama generation test
- Timeout values

## Response Format

### Comprehensive Health Response
```json
{
  "status": "healthy|degraded|unhealthy",
  "timestamp": "2025-07-23T13:00:00Z",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "services": [...],
  "system_metrics": {...},
  "summary": {...}
}
```

### Simple Health Response
```json
{
  "status": "healthy",
  "timestamp": "2025-07-23T13:00:00Z",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "system": {
    "cpu_percent": 25.5,
    "memory_percent": 60.2
  }
}
```

### Analyzed Health Response
Includes all comprehensive health data plus:
```json
{
  "llm_analysis": {
    "available": true,
    "summary": "AI-generated analysis of system health...",
    "details": {
      "model": "phi3:mini",
      "health_status": "unhealthy",
      "services": 4
    },
    "timestamp": "2025-07-23T13:00:00Z"
  }
}
```

## Building

```bash
# Download dependencies
make deps

# Build for current platform
make build

# Clean build artifacts
make clean
```

## Integration

The health checker integrates with the main project Makefile:

```bash
# From project root
make health           # Comprehensive check
make health-simple    # Quick check
make health-analyzed  # AI-powered analysis
make health-server    # Run server
```

## AI-Powered Health Analysis

When Ollama is available, the health checker can provide intelligent analysis of system health:

1. **Automatic Problem Detection**: Identifies root causes of failures
2. **Performance Insights**: Analyzes resource usage patterns
3. **Actionable Recommendations**: Provides specific steps to resolve issues
4. **Context-Aware**: Understands relationships between services

Example use cases:
- Diagnose why services are failing
- Identify performance bottlenecks
- Get optimization suggestions
- Understand system resource constraints

## Migration from Python

This Go implementation replaces the Python `healthcheck.py` with improved performance and native compilation. It maintains API compatibility while adding:
- Concurrent service checks
- Better error handling
- Native system metrics collection
- AI-powered health analysis
- Smaller binary size
- No Python dependencies