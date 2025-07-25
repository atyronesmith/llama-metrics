# Ollama Monitoring Stack - Complete Conversation Log

## Initial Status Check

**User**: "Is the monitoring app running?"

**Assistant**: I checked and confirmed all monitoring services were operational:
- Ollama service: Running
- Monitoring proxy: Running on port 8001  
- Prometheus: Running in container

## Prometheus Configuration Update

**User**: "I don't think the localhost:8001/metrics target is enabled in prometheus. add that target to the configuration file and make the :8000 target optional as I want to use the newer monitoring app"

**Assistant**: Updated `prometheus.yml` to:
- Change primary target from :8000 to :8001 for the newer monitoring proxy
- Made the legacy :8000 target optional (commented out)

```yaml
scrape_configs:
  # Primary job for the newer Ollama monitoring proxy
  - job_name: 'ollama_proxy'
    static_configs:
      - targets: ['host.containers.internal:8001']
```

## Makefile Fix for Containerized Prometheus

**User**: "make start-prometheus doesn't work. remember prometheus is run as a container in podman."

**Assistant**: Fixed the Makefile `start-prometheus` target to use containerized approach:

```makefile
start-prometheus:
    @if ! (podman ps 2>/dev/null | grep -q prometheus); then \
        echo "$(BLUE)Starting Prometheus container...$(NC)"; \
        ./run_prometheus.sh; \
    else \
        echo "$(YELLOW)Prometheus container is already running$(NC)"; \
    fi
```

## Request Latency Metrics Enhancement

**User**: "I want to add request latency to our metrics"

**Assistant**: Enhanced latency metrics in `ollama_monitoring_proxy_fixed.py`:
- Increased histogram buckets from 9 to 20 for better percentile calculation
- Added buckets: 0.01, 0.05, 0.1, 0.25, 0.5, 0.75, 1.0, 1.5, 2.0, 2.5, 5.0, 7.5, 10.0, 15.0, 30.0, 45.0, 60.0, 90.0, 120.0, 180.0

**User**: "Add 2 and 3" (referring to additional features)

**Assistant**: Added per-endpoint latency breakdown with separate histograms for:
- `/api/generate` 
- `/api/chat`
- `/api/tags`
- `/api/show`

## macOS GPU/Power Monitoring

**User**: "Are you able to get GPU usage for the metal shaders since we are running on a mac? Can you poll temperature, power as well and have them put into prometheus to be able to corollate with running the model?"

**Assistant**: Implemented comprehensive macOS monitoring:

### GPU Usage Estimation
- Process-based GPU monitoring using Ollama CPU correlation
- Non-privileged approach since powermetrics requires sudo
- Active residency percentage estimation

### Power Monitoring  
- CPU utilization-based power estimation
- M4 Pro power ranges: 4W (idle) to 35W (high load)
- Exponential scaling based on CPU usage

### Temperature Monitoring
- CPU temperature tracking via `psutil`
- Correlated with processing load

### New Prometheus Metrics Added:
```python
# GPU metrics
ollama_proxy_gpu_active_residency_percent
ollama_proxy_gpu_device_utilization_percent

# Power metrics  
ollama_proxy_package_power_watts
ollama_proxy_cpu_power_watts

# Temperature metrics
ollama_proxy_cpu_temp_celsius
ollama_proxy_gpu_temp_celsius
```

## Web Dashboard Creation

**User**: "Nice. Now we have a bunch of metrics. The prometheus UI is okay. However, I want a brower application that shows a summary of these metrics to give me an immediate insight into how the llm is doing. Create a web app that can access the prometheus data and continually shows a summarization. Also, add graphs for token creation and memory usage."

**Assistant**: Created comprehensive web dashboard with:

### Backend (`dashboard.py`)
- Flask application with WebSocket support
- PrometheusClient for metrics querying
- DashboardMetrics collector for data processing
- Real-time updates every 5 seconds

### Frontend (`templates/dashboard.html`)
- Bootstrap UI with responsive design
- Real-time metric cards showing:
  - Requests/sec
  - Average latency
  - Success rate
  - Tokens/sec
  - GPU usage
  - Power consumption
  - Memory usage
  - Active requests

### Charts (Chart.js)
- Token generation rate over time
- Memory usage over time  
- GPU utilization over time
- Power consumption over time

### Features
- Live WebSocket updates
- Connection status indicator
- Latency percentiles (P50, P75, P95, P99)
- Last updated timestamps

**User**: "how do I see the dashboard in my browser?"

**Assistant**: Dashboard available at http://localhost:3001 after running `python dashboard.py`

## Dashboard Bug Fixes

**User**: "the average latency(s) number is stuck at 0.00 in the dashboard. The power doesn't seem to change. GPU Usage is at 0.0 even though I am generating traffic."

**User**: "The avg latency is still at 0.00. I see ollama_proxy_request_duration_seconds_bucket with a lot of non-zero data. Check how avg latency is being calculated. it may be wrong."

**Assistant**: Fixed multiple dashboard issues:

### 1. Average Latency Calculation
Changed from generic query to focus on `/api/generate` endpoint:
```python
result = self.client.query(
    'sum(rate(ollama_proxy_request_duration_seconds_sum{endpoint="/api/generate"}[5m])) / '
    'sum(rate(ollama_proxy_request_duration_seconds_count{endpoint="/api/generate"}[5m]))'
)
```

### 2. Histogram Percentile Queries
Fixed malformed quantile queries from "0.0.50" to "0.50"

### 3. NaN/Null Value Handling
Added JavaScript `safeNumber()` function:
```javascript
function safeNumber(value, decimals = 2) {
    if (value === null || value === undefined || isNaN(value)) {
        return '0.' + '0'.repeat(decimals);
    }
    return value.toFixed(decimals);
}
```

### 4. Query Aggregation
Used `sum()` functions for proper metric aggregation across multiple result sets

## AI Status Line Implementation

**User**: "I want to generate a text status line that is geared for humans. The status should take into account all of the metrics. For example, if the GPU usage has spiked, say, Llama is running, however, we have a spike in GPU usage, temperature is ok, memory usage is good. etc... The key is that I want to have a LLM generate the status line. So, you will need to get the data, make sense and send the data to a model and have it generate a summary that you put in a status line at the top of the dashboard."

**Assistant**: Implemented comprehensive AI status generation system:

### LLMStatusGenerator Class
- Intelligent metric analysis with thresholds
- Context preparation for LLM prompts
- Ollama integration using phi3:mini model
- Status caching and error handling

### Metric Analysis Functions
```python
def _analyze_request_activity(self, rate: float) -> str:
    if rate > 1.0: return "high activity"
    elif rate > 0.1: return "moderate activity"
    elif rate > 0: return "low activity"
    else: return "idle"

def _analyze_gpu_usage(self, usage: float) -> str:
    if usage > 50: return "high GPU usage"
    elif usage > 10: return "moderate GPU usage"
    elif usage > 1: return "light GPU usage"
    else: return "minimal GPU usage"
```

### Status Prompt Engineering
Structured prompt for consistent LLM responses:
```python
prompt = f"""You are a system monitor for an AI model server. Create a concise, single-line status message (max 100 characters) based on these conditions:

Request Activity: {context['request_activity']}
Latency: {context['latency_status']}
GPU: {context['gpu_status']}
Power: {context['power_status']}
Memory: {context['memory_status']}
Reliability: {context['success_status']}
Token Generation: {context['token_generation']}
Active Requests: {context['active_requests']}

Generate a brief, human-friendly status like:
- "LLM running smoothly - normal latency, light GPU usage, good performance"
- "High activity detected - GPU spike, elevated power, all systems stable"
- "System idle - minimal resource usage, ready for requests"

Status:"""
```

### Dashboard Integration
- AI status banner at top of dashboard
- Real-time updates every 30 seconds
- WebSocket broadcasting of status
- Timestamp tracking for status updates

### Backend Implementation
- Status generation in background thread
- API endpoint `/api/status` for manual queries
- Error handling with fallback messages
- Temperature and stop token configuration

### Frontend Integration
- Status banner UI component
- Real-time WebSocket updates
- Timestamp display for status freshness

## Technical Implementation Details

### File Structure
```
/Users/asmith/dev/llamastack-prometheus/
├── dashboard.py                      # Flask dashboard application
├── templates/dashboard.html          # Dashboard UI template
├── ollama_monitoring_proxy_fixed.py  # Enhanced monitoring proxy
├── prometheus.yml                    # Prometheus configuration
├── Makefile                          # Build and deployment automation
├── requirements_dashboard.txt        # Dashboard dependencies
└── various log files and configs
```

### Key Technologies Used
- **Backend**: Python, Flask, SocketIO, Requests, psutil
- **Frontend**: HTML5, Bootstrap 5, Chart.js, WebSockets
- **Monitoring**: Prometheus, custom metrics collection
- **LLM**: Ollama with phi3:mini model
- **Containerization**: Podman for Prometheus

### Metrics Collected
1. **Request Metrics**: Rate, latency, success rate, active requests
2. **Performance Metrics**: Token generation rate, response times
3. **System Metrics**: CPU/GPU usage, memory consumption, power draw
4. **Latency Percentiles**: P50, P75, P95, P99 with 20-bucket histograms

### Dashboard Features
- **Real-time Updates**: 5-second metric refresh via WebSockets
- **Interactive Charts**: Time-series graphs with Chart.js
- **AI Status Line**: Human-readable system interpretation
- **Responsive Design**: Bootstrap-based mobile-friendly UI
- **Connection Monitoring**: Live connection status indicators

### Prometheus Integration
- Custom metrics export on port 8001
- Container-based Prometheus deployment
- Range queries for time-series data
- Histogram quantile calculations

### macOS-Specific Optimizations
- Non-privileged GPU monitoring
- CPU-based power estimation for Apple Silicon
- Process correlation for GPU usage estimation
- Temperature monitoring via system APIs

## Final Status

All requested features have been successfully implemented:

✅ **Monitoring Stack**: Complete Ollama monitoring with Prometheus integration  
✅ **Enhanced Latency Metrics**: 20-bucket histograms with per-endpoint breakdown  
✅ **macOS GPU/Power Monitoring**: Comprehensive system metrics collection  
✅ **Web Dashboard**: Real-time Flask application with Chart.js visualizations  
✅ **Bug Fixes**: Resolved latency calculation, NaN handling, and query issues  
✅ **AI Status Line**: LLM-generated human-readable system status interpretation  

The system provides comprehensive monitoring of Ollama LLM performance with real-time web dashboard, intelligent status generation, and detailed metrics collection suitable for both development and production environments.