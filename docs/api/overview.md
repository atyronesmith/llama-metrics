# Ollama Monitoring Stack API Documentation

## Overview

The Ollama Monitoring Stack provides several APIs for monitoring, metrics collection, and dashboard functionality. This document describes all available endpoints, their parameters, and expected responses.

## Base URLs

- **Monitoring Proxy**: `http://localhost:11435`
- **Metrics Server**: `http://localhost:8001`
- **Dashboard Server**: `http://localhost:3001`
- **Ollama API** (proxied): `http://localhost:11434`

## Authentication

Currently, no authentication is required for API access. This is suitable for development and internal monitoring setups. For production deployments, consider implementing API key authentication.

## Content Types

- **Request**: `application/json`
- **Response**: `application/json` or `text/plain` (for metrics)

---

## Monitoring Proxy API

The monitoring proxy runs on port `11435` and forwards requests to Ollama while collecting metrics.

### Chat Completion

Proxies chat completion requests to Ollama while collecting performance metrics.

**Endpoint:** `POST /api/chat`

**Headers:**
- `Content-Type: application/json`
- `Accept: application/json` (optional, for JSON responses)
- `Accept: text/event-stream` (optional, for streaming responses)

**Request Body:**
```json
{
  "model": "phi3:mini",
  "messages": [
    {
      "role": "user",
      "content": "What is machine learning?"
    }
  ],
  "stream": false,
  "options": {
    "temperature": 0.7,
    "top_p": 0.9
  }
}
```

**Response (Non-streaming):**
```json
{
  "model": "phi3:mini",
  "created_at": "2023-12-07T09:00:00.000Z",
  "message": {
    "role": "assistant",
    "content": "Machine learning is a subset of artificial intelligence..."
  },
  "done": true,
  "total_duration": 5000000000,
  "load_duration": 500000000,
  "prompt_eval_count": 15,
  "prompt_eval_duration": 1000000000,
  "eval_count": 150,
  "eval_duration": 3500000000
}
```

**Response (Streaming):**
```
data: {"model":"phi3:mini","created_at":"2023-12-07T09:00:00.000Z","message":{"role":"assistant","content":"Machine"},"done":false}

data: {"model":"phi3:mini","created_at":"2023-12-07T09:00:00.000Z","message":{"role":"assistant","content":" learning"},"done":false}

data: {"model":"phi3:mini","created_at":"2023-12-07T09:00:00.000Z","message":{"role":"assistant","content":""},"done":true,"total_duration":5000000000,"eval_count":150}
```

### Generate Completion

**Endpoint:** `POST /api/generate`

**Request Body:**
```json
{
  "model": "phi3:mini",
  "prompt": "What is the capital of France?",
  "stream": false,
  "options": {
    "temperature": 0.7
  }
}
```

**Response:**
```json
{
  "model": "phi3:mini",
  "created_at": "2023-12-07T09:00:00.000Z",
  "response": "The capital of France is Paris.",
  "done": true,
  "total_duration": 3000000000,
  "load_duration": 200000000,
  "prompt_eval_count": 8,
  "prompt_eval_duration": 500000000,
  "eval_count": 7,
  "eval_duration": 2300000000
}
```

### List Models

**Endpoint:** `GET /api/tags`

**Response:**
```json
{
  "models": [
    {
      "name": "phi3:mini",
      "model": "phi3:mini",
      "modified_at": "2023-12-07T09:00:00.000Z",
      "size": 2300000000,
      "digest": "sha256:abc123...",
      "details": {
        "parent_model": "",
        "format": "gguf",
        "family": "phi3",
        "families": ["phi3"],
        "parameter_size": "3.8B",
        "quantization_level": "Q4_K_M"
      }
    }
  ]
}
```

---

## Metrics API

The metrics server runs on port `8001` and provides Prometheus-compatible metrics.

### Prometheus Metrics

**Endpoint:** `GET /metrics`

**Response Format:** Prometheus exposition format

**Sample Response:**
```
# HELP ollama_proxy_requests_total Total number of requests processed
# TYPE ollama_proxy_requests_total counter
ollama_proxy_requests_total{method="POST",endpoint="/api/chat",model="phi3:mini",status="200"} 42.0

# HELP ollama_proxy_request_duration_seconds Request duration in seconds
# TYPE ollama_proxy_request_duration_seconds histogram
ollama_proxy_request_duration_seconds_bucket{method="POST",endpoint="/api/chat",model="phi3:mini",le="0.1"} 5.0
ollama_proxy_request_duration_seconds_bucket{method="POST",endpoint="/api/chat",model="phi3:mini",le="0.5"} 15.0
ollama_proxy_request_duration_seconds_bucket{method="POST",endpoint="/api/chat",model="phi3:mini",le="1.0"} 25.0
ollama_proxy_request_duration_seconds_bucket{method="POST",endpoint="/api/chat",model="phi3:mini",le="5.0"} 40.0
ollama_proxy_request_duration_seconds_bucket{method="POST",endpoint="/api/chat",model="phi3:mini",le="10.0"} 42.0
ollama_proxy_request_duration_seconds_bucket{method="POST",endpoint="/api/chat",model="phi3:mini",le="+Inf"} 42.0
ollama_proxy_request_duration_seconds_sum{method="POST",endpoint="/api/chat",model="phi3:mini"} 85.5
ollama_proxy_request_duration_seconds_count{method="POST",endpoint="/api/chat",model="phi3:mini"} 42.0

# HELP ollama_proxy_active_requests Current number of active requests
# TYPE ollama_proxy_active_requests gauge
ollama_proxy_active_requests 3.0

# HELP ollama_proxy_queue_size Current queue size
# TYPE ollama_proxy_queue_size gauge
ollama_proxy_queue_size 2.0

# HELP ollama_proxy_tokens_per_second Tokens generated per second
# TYPE ollama_proxy_tokens_per_second gauge
ollama_proxy_tokens_per_second{model="phi3:mini"} 45.2

# HELP ollama_proxy_system_cpu_percent CPU usage percentage
# TYPE ollama_proxy_system_cpu_percent gauge
ollama_proxy_system_cpu_percent 23.5

# HELP ollama_proxy_system_memory_percent Memory usage percentage
# TYPE ollama_proxy_system_memory_percent gauge
ollama_proxy_system_memory_percent 67.8

# HELP ollama_proxy_system_memory_bytes Memory usage in bytes
# TYPE ollama_proxy_system_memory_bytes gauge
ollama_proxy_system_memory_bytes{type="available"} 8589934592.0
ollama_proxy_system_memory_bytes{type="used"} 5825781760.0

# HELP ollama_proxy_gpu_utilization_percent GPU utilization percentage (macOS Metal)
# TYPE ollama_proxy_gpu_utilization_percent gauge
ollama_proxy_gpu_utilization_percent 15.2

# HELP ollama_proxy_power_watts Power consumption in watts (macOS)
# TYPE ollama_proxy_power_watts gauge
ollama_proxy_power_watts{component="cpu"} 12.5
ollama_proxy_power_watts{component="gpu"} 8.3
```

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2023-12-07T09:00:00.000Z",
  "version": "1.0.0",
  "services": {
    "ollama": {
      "status": "healthy",
      "url": "http://localhost:11434",
      "response_time_ms": 15
    },
    "proxy": {
      "status": "healthy",
      "active_requests": 2,
      "queue_size": 0
    },
    "metrics": {
      "status": "healthy",
      "total_requests": 42,
      "avg_response_time": 2.1
    }
  },
  "system": {
    "cpu_percent": 23.5,
    "memory_percent": 67.8,
    "disk_usage_percent": 45.2
  }
}
```

### Metrics Summary

**Endpoint:** `GET /api/metrics/summary`

**Response:**
```json
{
  "timestamp": "2023-12-07T09:00:00.000Z",
  "summary": {
    "total_requests": 42,
    "active_requests": 2,
    "queue_size": 0,
    "avg_latency": 2.1,
    "tokens_per_second": 45.2,
    "success_rate": 0.95,
    "error_rate": 0.05
  },
  "system": {
    "cpu_percent": 23.5,
    "memory_percent": 67.8,
    "memory_available_gb": 8.0,
    "memory_used_gb": 5.4,
    "gpu_utilization_percent": 15.2,
    "power_cpu_watts": 12.5,
    "power_gpu_watts": 8.3
  },
  "models": {
    "phi3:mini": {
      "requests": 42,
      "avg_latency": 2.1,
      "tokens_per_second": 45.2,
      "last_request": "2023-12-07T08:58:30.000Z"
    }
  }
}
```

---

## Dashboard API

The dashboard server runs on port `3001` and provides real-time monitoring interface.

### WebSocket Connection

**Endpoint:** `WebSocket /socket.io/`

**Events:**

#### Client to Server:
- `connect`: Establish connection
- `disconnect`: Close connection
- `request_update`: Request immediate metrics update

#### Server to Client:
- `metrics_update`: Real-time metrics data
- `ai_status_update`: AI-generated status summary
- `system_alert`: System alerts and warnings

**Sample metrics_update payload:**
```json
{
  "timestamp": "2023-12-07T09:00:00.000Z",
  "metrics": {
    "active_requests": 2,
    "queue_size": 0,
    "avg_latency": 2.1,
    "tokens_per_second": 45.2,
    "cpu_percent": 23.5,
    "memory_percent": 67.8,
    "gpu_percent": 15.2,
    "power_watts": 20.8
  },
  "ollama_status": "healthy"
}
```

### Dashboard Status

**Endpoint:** `GET /api/status`

**Response:**
```json
{
  "status": "online",
  "version": "1.0.0",
  "uptime_seconds": 3600,
  "connected_clients": 2,
  "last_update": "2023-12-07T09:00:00.000Z",
  "ai_status_enabled": true,
  "monitoring_active": true
}
```

---

## Error Responses

All APIs follow a consistent error response format:

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Invalid request format",
    "details": "Missing required field 'model'",
    "timestamp": "2023-12-07T09:00:00.000Z"
  }
}
```

**Common Error Codes:**

- `INVALID_REQUEST`: Request format or parameters invalid
- `MODEL_NOT_FOUND`: Specified model not available
- `SERVICE_UNAVAILABLE`: Backend service (Ollama) not available
- `RATE_LIMIT_EXCEEDED`: Too many requests
- `INTERNAL_ERROR`: Server internal error
- `TIMEOUT`: Request timeout

**HTTP Status Codes:**

- `200`: Success
- `400`: Bad Request
- `404`: Not Found
- `429`: Too Many Requests
- `500`: Internal Server Error
- `502`: Bad Gateway (Ollama unavailable)
- `503`: Service Unavailable
- `504`: Gateway Timeout

---

## Rate Limiting

Default rate limits (configurable in `config.yml`):

- **Monitoring Proxy**: 60 requests/minute per client
- **Metrics API**: 100 requests/minute per client
- **Dashboard API**: No limit (WebSocket-based)

Rate limit headers are included in responses:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
X-RateLimit-Reset: 1701936000
```

---

## Configuration

API behavior can be configured via `config.yml`:

```yaml
server:
  proxy_port: 11435
  metrics_port: 8001
  dashboard_port: 3001
  request_timeout: 30

monitoring:
  rate_limit_per_minute: 60
  max_concurrent_requests: 10
  max_queue_size: 50

security:
  require_api_key: false
  rate_limiting:
    enabled: true
    requests_per_minute: 100
```

---

## SDK Examples

### Python

```python
import requests
import json

# Chat completion
def chat_completion(message, model="phi3:mini"):
    url = "http://localhost:11435/api/chat"
    payload = {
        "model": model,
        "messages": [{"role": "user", "content": message}],
        "stream": False
    }
    
    response = requests.post(url, json=payload)
    return response.json()

# Get metrics
def get_metrics():
    response = requests.get("http://localhost:8001/metrics")
    return response.text

# Get health status
def get_health():
    response = requests.get("http://localhost:8001/health")
    return response.json()

# Example usage
result = chat_completion("What is Python?")
print(result["message"]["content"])

health = get_health()
print(f"Status: {health['status']}")
```

### JavaScript/Node.js

```javascript
const axios = require('axios');

// Chat completion
async function chatCompletion(message, model = 'phi3:mini') {
  const response = await axios.post('http://localhost:11435/api/chat', {
    model: model,
    messages: [{ role: 'user', content: message }],
    stream: false
  });
  return response.data;
}

// Get metrics summary
async function getMetricsSummary() {
  const response = await axios.get('http://localhost:8001/api/metrics/summary');
  return response.data;
}

// Example usage
chatCompletion('What is JavaScript?')
  .then(result => console.log(result.message.content))
  .catch(err => console.error(err));
```

### curl Examples

```bash
# Chat completion
curl -X POST http://localhost:11435/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "phi3:mini",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": false
  }'

# Get Prometheus metrics
curl http://localhost:8001/metrics

# Get health status
curl http://localhost:8001/health

# Get metrics summary
curl http://localhost:8001/api/metrics/summary
```

---

## Monitoring and Debugging

### Enable Debug Logging

Set environment variable:
```bash
export LOG_LEVEL=DEBUG
```

Or modify `config.yml`:
```yaml
logging:
  level: "DEBUG"
```

### Common Issues

1. **Connection refused**: Check if services are running with `make status`
2. **Model not found**: Ensure model is pulled with `ollama pull phi3:mini`
3. **High latency**: Check system resources and reduce concurrent requests
4. **Rate limiting**: Adjust limits in configuration or implement backoff

### Performance Tips

- Use streaming for long responses to reduce perceived latency
- Monitor queue size and adjust `max_concurrent_requests` accordingly
- Enable compression for metrics endpoint in production
- Use WebSocket connection for real-time dashboard updates

---

## Changelog

- **v1.0.0**: Initial API documentation
- Added comprehensive endpoint documentation
- Added error handling guidelines
- Added SDK examples and curl commands