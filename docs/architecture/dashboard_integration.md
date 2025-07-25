# Dashboard Integration Guide

This guide explains how to integrate the Go-based Ollama monitoring proxy with the Python dashboard for real-time metrics visualization.

## Architecture Overview

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│   Dashboard     │────▶│   Prometheus     │────▶│  Go Proxy   │
│  (Port 3001)    │     │  (Port 9090)     │     │ (Port 8001) │
└─────────────────┘     └──────────────────┘     └─────────────┘
                                                          │
                              ┌───────────────────────────┴───────┐
                              │                                   │
                        ┌─────▼─────┐                    ┌───────▼──────┐
                        │  Ollama   │                    │ Mac Metrics  │
                        │(Port 11434)│                    │ Helper:8002  │
                        └───────────┘                    └──────────────┘
```

## Components

### 1. Go Proxy (Port 11435 & 8001)
- **Proxy Port (11435)**: Intercepts and forwards requests to Ollama
- **Metrics Port (8001)**: Exposes Prometheus metrics at `/metrics`
- Collects request metrics, token counts, performance data
- On macOS: Also collects system metrics (GPU, temperature, etc.)

### 2. Mac Metrics Helper (Port 8002) - Optional
- Python service that collects privileged Mac system metrics
- Required for GPU utilization, power consumption, temperature
- Runs with sudo permissions
- Provides JSON API for the Go proxy

### 3. Dashboard (Port 3001)
- Flask/SocketIO web application
- Queries Prometheus for metrics
- Real-time updates via WebSocket
- Visualizes LLM and system performance

## Quick Start

### 1. Start the Mac Metrics Helper (macOS only, optional)

```bash
# Requires sudo for accessing system metrics
cd proxy/scripts
sudo python3 mac_metrics_helper.py
```

### 2. Start the Go Proxy

```bash
cd proxy
make run

# Or with the helper (macOS)
make run-with-helper
```

### 3. Configure Prometheus

Ensure `prometheus.yml` includes:

```yaml
scrape_configs:
  - job_name: 'ollama_proxy'
    static_configs:
      - targets: ['localhost:8001']
    metric_relabel_configs:
      - source_labels: [__name__]
        regex: 'ollama_proxy_.*'
        action: keep
```

### 4. Start Prometheus

```bash
prometheus --config.file=prometheus.yml
```

### 5. Start the Dashboard

```bash
python dashboard.py
```

## Available Metrics

### LLM Performance Metrics
- `ollama_proxy_requests_total` - Total requests by model, method, status
- `ollama_proxy_request_duration_seconds` - Request latency histogram
- `ollama_proxy_active_requests` - Currently processing requests
- `ollama_proxy_prompt_tokens_total` - Total prompt tokens
- `ollama_proxy_generated_tokens_total` - Total generated tokens
- `ollama_proxy_tokens_per_second` - Token generation rate
- `ollama_proxy_time_to_first_token_seconds` - TTFT histogram
- `ollama_proxy_model_load_duration_seconds` - Model loading time

### System Metrics (All Platforms)
- `ollama_proxy_cpu_usage_percent` - CPU usage percentage
- `ollama_proxy_memory_usage_bytes` - Memory usage in bytes

### Mac-Specific Metrics
- `ollama_proxy_gpu_active_residency_percent` - GPU utilization
- `ollama_proxy_gpu_power_milliwatts` - GPU power consumption
- `ollama_proxy_cpu_power_milliwatts` - CPU package power
- `ollama_proxy_cpu_temperature_celsius` - CPU temperature
- `ollama_proxy_memory_pressure_percent` - Memory pressure
- `ollama_proxy_disk_read_bytes_per_second` - Disk read rate
- `ollama_proxy_disk_write_bytes_per_second` - Disk write rate
- `ollama_proxy_disk_iops` - Disk I/O operations per second

## Dashboard Configuration

The dashboard expects these metrics from Prometheus. Key queries:

```promql
# GPU Utilization
ollama_proxy_gpu_active_residency_percent

# Memory Usage (MB)
ollama_proxy_memory_usage_bytes / 1024 / 1024

# Active Requests
sum(ollama_proxy_active_requests)

# Request Rate
rate(ollama_proxy_requests_total[1m])

# Average Latency
rate(ollama_proxy_request_duration_seconds_sum[1m]) /
rate(ollama_proxy_request_duration_seconds_count[1m])

# Token Generation Rate
rate(ollama_proxy_generated_tokens_total[1m])
```

## Troubleshooting

### No GPU Metrics on macOS
1. Check if the metrics helper is running: `curl http://localhost:8002/metrics`
2. Ensure running with sudo: `sudo python3 mac_metrics_helper.py`
3. Install `osx-cpu-temp` for temperature: `brew install osx-cpu-temp`

### Missing Metrics in Dashboard
1. Verify Go proxy is exposing metrics: `curl http://localhost:8001/metrics`
2. Check Prometheus is scraping: Visit http://localhost:9090/targets
3. Ensure metric names match between proxy and dashboard queries

### High CPU Usage
- The Mac metrics helper uses `powermetrics` which can be CPU intensive
- Adjust collection interval in `mac_metrics_helper.py` (default: 5 seconds)
- Disable helper if not needed for GPU/power metrics

## Development Tips

### Adding New Metrics

1. **In Go Proxy** (`proxy/internal/metrics/metrics.go`):
```go
NewMetric: promauto.NewGauge(
    prometheus.GaugeOpts{
        Name: "ollama_proxy_new_metric",
        Help: "Description of metric",
    },
),
```

2. **In Dashboard** (`dashboard.py`):
```python
result = self.client.query('ollama_proxy_new_metric')
if result['data']['result']:
    value = float(result['data']['result'][0]['value'][1])
```

### Testing Integration

```bash
# Test metric collection
curl http://localhost:8001/metrics | grep ollama_proxy

# Test Prometheus query
curl 'http://localhost:9090/api/v1/query?query=ollama_proxy_gpu_active_residency_percent'

# Test dashboard data endpoint
curl http://localhost:3001/api/metrics
```

## Performance Considerations

- The Go proxy adds < 1ms latency to requests
- Mac metrics collection runs every 10 seconds (configurable)
- Helper service updates every 5 seconds (configurable)
- Dashboard polls Prometheus every 5 seconds
- Use WebSocket for real-time updates to reduce polling

## Security Notes

- The metrics helper requires sudo for system metrics
- Consider running helper as a system service with limited permissions
- Metrics endpoints have no authentication - use firewall rules in production
- Don't expose ports 8001, 8002, 9090 to public networks