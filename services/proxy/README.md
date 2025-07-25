# Ollama Monitoring Proxy

A high-performance, AI-oriented monitoring proxy for Ollama with comprehensive metrics tracking and OpenAI API compatibility.

## Features

### AI-Oriented Metrics
- **Token Tracking**: Monitor prompt tokens, generated tokens, and tokens per second
- **Request Analytics**: Track requests by ID, user, and model
- **Performance Metrics**: Measure latency, time to first token, and model load times
- **Cost Estimation**: Track token usage costs per user and model
- **Request/Response Size**: Monitor payload sizes for optimization

### OpenAI API Compatibility
- **Drop-in Replacement**: Use OpenAI SDKs and tools with Ollama
- **Endpoint Support**: `/v1/chat/completions`, `/v1/completions`, `/v1/models`
- **Automatic Model Mapping**: Maps OpenAI model names to Ollama equivalents
- **Streaming Support**: Full support for streaming responses

### System Monitoring
- Cross-platform system metrics (CPU, memory, disk I/O)
- Mac-specific metrics (GPU utilization, power consumption, temperature)
- Real-time performance tracking

## Quick Start

### Build and Run

```bash
# Build the proxy
make build

# Run the proxy
./proxy

# Or use make
make run
```

### Using OpenAI SDK

```python
from openai import OpenAI

client = OpenAI(
    api_key="not-needed",
    base_url="http://localhost:11434/v1"
)

response = client.chat.completions.create(
    model="gpt-3.5-turbo",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### Configuration

Environment variables:
- `OLLAMA_PROXY_PORT`: Proxy port (default: 11434)
- `OLLAMA_METRICS_PORT`: Metrics port (default: 9090)
- `OLLAMA_HOST`: Ollama backend host (default: localhost)
- `OLLAMA_PORT`: Ollama backend port (default: 11434)

## Metrics

Access Prometheus metrics at `http://localhost:9090/metrics`

Key metrics include:
- `ollama_proxy_requests_total`: Total requests by method, endpoint, model, and status
- `ollama_proxy_request_duration_seconds`: Request latency distribution
- `ollama_proxy_tokens_per_second`: Token generation speed
- `ollama_proxy_time_to_first_token_seconds`: Time to first token
- `ollama_proxy_token_cost_total`: Estimated token costs

## Architecture

```
┌─────────────┐     ┌─────────────────┐     ┌──────────┐
│   Client    │────▶│  Proxy (11434)  │────▶│  Ollama  │
│ (OpenAI SDK)│     │                 │     │          │
└─────────────┘     │  - Metrics      │     └──────────┘
                    │  - Transform     │
                    │  - Track         │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │ Prometheus      │
                    │ (9090)          │
                    └─────────────────┘
```

## Documentation

- [AI Metrics Guide](docs/AI_METRICS_GUIDE.md) - Detailed guide on AI-oriented features
- [Dashboard Integration](docs/DASHBOARD_INTEGRATION.md) - Grafana dashboard setup

## Development

```bash
# Run tests
make test

# Format code
make fmt

# Lint code
make lint
```

## License

MIT# To prevent Ollama from spawning too many runners, set:
export OLLAMA_NUM_PARALLEL=2
# Before starting ollama serve
