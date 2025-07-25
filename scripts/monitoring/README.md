# Monitoring Scripts

This directory contains scripts for system monitoring, health checks, and metrics collection for the Ollama monitoring system.

## Scripts Overview

### 🚀 Quick Start
```bash
# Start complete monitoring stack
./scripts/monitoring/start_stack.sh

# Run health check
python scripts/monitoring/health_check.py

# Start Mac metrics collection (macOS only)
sudo python scripts/monitoring/mac_metrics.py
```

## Script Descriptions

### `start_stack.sh` - Monitoring Stack Launcher
**Purpose**: Orchestrates the startup of the complete monitoring infrastructure.

**Features**:
- Starts Ollama monitoring proxy
- Launches Prometheus container
- Initializes dashboard service
- Starts health checker
- Configures log collection

**Usage**:
```bash
./scripts/monitoring/start_stack.sh
```

**What it starts**:
- Ollama Proxy (port 11435 → 11434)
- Metrics endpoint (port 8001)
- Prometheus (port 9090)
- Dashboard (port 3001)
- Health checker (port 8080)

### `health_check.py` - System Health Checker
**Purpose**: Comprehensive health checking and system status verification.

**Features**:
- Ollama service connectivity tests
- Model availability verification
- Resource usage monitoring
- Service dependency checks
- Automated issue detection
- Health status reporting

**Usage**:
```bash
# Basic health check
python scripts/monitoring/health_check.py

# Detailed analysis
python scripts/monitoring/health_check.py --verbose

# JSON output for automation
python scripts/monitoring/health_check.py --json

# Continuous monitoring
python scripts/monitoring/health_check.py --watch
```

**Health Checks Performed**:
- Ollama service availability
- Model loading status
- Proxy connectivity
- Prometheus scraping
- Dashboard accessibility
- System resource levels

### `mac_metrics.py` - macOS System Metrics Collector
**Purpose**: Collects macOS-specific system metrics that require elevated privileges.

**Features**:
- GPU utilization and temperature
- Power consumption monitoring
- Advanced memory statistics
- CPU core-level metrics
- Thermal throttling detection
- Battery status (laptops)

**Usage**:
```bash
# Start metrics collection (requires sudo)
sudo python scripts/monitoring/mac_metrics.py

# HTTP server mode (port 8002)
sudo python scripts/monitoring/mac_metrics.py --server

# One-shot metrics collection
sudo python scripts/monitoring/mac_metrics.py --once
```

**Metrics Collected**:
- GPU usage per core
- GPU memory utilization
- GPU temperature
- Power draw (watts)
- CPU frequency scaling
- Memory pressure
- Thermal state

## Prerequisites

### General Requirements
- **Python Virtual Environment**: `source venv/bin/activate`
- **Dependencies**: `pip install -r requirements.txt`
- **Ollama**: Running service for health checks

### macOS Specific (for mac_metrics.py)
- **Root Privileges**: Required for system metrics access
- **psutil**: Enhanced system information library
- **py-cpuinfo**: CPU information gathering

### Docker/Podman
- **Container Runtime**: For Prometheus deployment
- **Network Access**: Container networking for metrics scraping

## Configuration

### Environment Variables
```bash
# Service endpoints
export OLLAMA_URL="http://localhost:11434"
export PROXY_URL="http://localhost:11435"
export PROMETHEUS_URL="http://localhost:9090"
export DASHBOARD_URL="http://localhost:3001"

# Health check settings
export HEALTH_CHECK_INTERVAL=30
export HEALTH_CHECK_TIMEOUT=10

# Mac metrics settings
export MAC_METRICS_PORT=8002
export MAC_METRICS_INTERVAL=5
```

### Configuration Files
- `config.yml`: Service configuration
- `prometheus.yml`: Prometheus scraping configuration
- `ollama_alerts.yml`: Alert rules configuration

## Integration Points

### With Services
- **Proxy Service**: Health checks and metrics collection
- **Dashboard Service**: Status monitoring and alerts
- **Health Service**: Automated health analysis

### With Prometheus
- **Metrics Scraping**: Automatic discovery and collection
- **Alert Rules**: Based on health check results
- **Target Monitoring**: Service availability tracking

## Usage Patterns

### Development Workflow
```bash
# 1. Start monitoring stack
./scripts/monitoring/start_stack.sh

# 2. Verify all services are healthy
python scripts/monitoring/health_check.py

# 3. Generate some traffic
./scripts/traffic/run.sh --quick

# 4. Check metrics collection
curl http://localhost:8001/metrics
```

### Production Deployment
```bash
# 1. Start services with monitoring
./scripts/monitoring/start_stack.sh

# 2. Set up continuous health monitoring
python scripts/monitoring/health_check.py --watch &

# 3. Start Mac metrics collection (if on macOS)
sudo python scripts/monitoring/mac_metrics.py --server &

# 4. Configure alerts in Prometheus
```

### Troubleshooting Workflow
```bash
# 1. Check overall system health
python scripts/monitoring/health_check.py --verbose

# 2. Verify individual service status
curl http://localhost:11434/api/tags  # Ollama
curl http://localhost:8001/health     # Proxy
curl http://localhost:3001/api/health # Dashboard

# 3. Check Prometheus targets
open http://localhost:9090/targets
```

## Monitoring Outputs

### Health Check Results
```json
{
  "status": "healthy",
  "services": {
    "ollama": {"status": "healthy", "response_time": 45},
    "proxy": {"status": "healthy", "response_time": 12},
    "dashboard": {"status": "healthy", "response_time": 8}
  },
  "system": {
    "cpu_usage": 15.2,
    "memory_usage": 68.5,
    "disk_usage": 42.1
  }
}
```

### Mac Metrics Output
```json
{
  "gpu": {
    "utilization": 45.2,
    "temperature": 67,
    "power_draw": 15.8
  },
  "cpu": {
    "frequency": 2400,
    "thermal_state": "nominal"
  },
  "memory": {
    "pressure": "normal",
    "swap_used": 0
  }
}
```

## Troubleshooting

### Common Issues

**Permission Denied (mac_metrics.py)**
```bash
# Run with sudo
sudo python scripts/monitoring/mac_metrics.py

# Or fix permissions
sudo chown root scripts/monitoring/mac_metrics.py
sudo chmod +s scripts/monitoring/mac_metrics.py
```

**Service Not Found**
```bash
# Check if services are running
ps aux | grep ollama
ps aux | grep prometheus

# Restart monitoring stack
./scripts/monitoring/start_stack.sh
```

**Metrics Not Collecting**
```bash
# Verify endpoints
curl http://localhost:8001/metrics
curl http://localhost:8002/metrics  # Mac metrics

# Check Prometheus configuration
curl http://localhost:9090/config
```

**Health Check Failures**
```bash
# Run detailed health check
python scripts/monitoring/health_check.py --verbose --debug

# Check individual services
curl -I http://localhost:11434
curl -I http://localhost:11435
```

## Best Practices

1. **Regular Health Checks**: Run health checks before major operations
2. **Gradual Rollouts**: Start monitoring before deploying changes
3. **Resource Monitoring**: Watch system resources during load tests
4. **Log Aggregation**: Centralize logs for easier troubleshooting
5. **Alert Tuning**: Configure appropriate alert thresholds

## Integration with Services

These monitoring scripts work with:
- **Proxy Service**: For metrics collection
- **Dashboard Service**: For real-time visualization
- **Health Service**: For automated analysis
- **Prometheus**: For metrics storage and alerting