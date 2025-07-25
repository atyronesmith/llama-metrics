# AI-Oriented Metrics and OpenAI Compatibility Guide

The enhanced proxy now provides comprehensive AI-oriented metrics tracking and OpenAI API compatibility. This guide explains the features and how to use them.

## Features

### 1. AI-Oriented Metrics

The proxy tracks the following AI-specific metrics:

#### Token Metrics
- **`ollama_proxy_prompt_tokens_total`**: Total prompt tokens processed
- **`ollama_proxy_generated_tokens_total`**: Total tokens generated
- **`ollama_proxy_tokens_per_second`**: Token generation speed
- **`ollama_proxy_context_length`**: Context length distribution

#### Request Tracking
- **`ollama_proxy_request_by_id_total`**: Requests tracked by unique ID
- **`ollama_proxy_user_requests_total`**: Requests per user
- **`ollama_proxy_active_requests`**: Currently processing requests
- **`ollama_proxy_requests_total`**: Total request count

#### Performance Metrics
- **`ollama_proxy_request_duration_seconds`**: End-to-end request latency
- **`ollama_proxy_time_to_first_token_seconds`**: Time to first token (TTFT)
- **`ollama_proxy_model_load_duration_seconds`**: Model loading time

#### Cost Tracking
- **`ollama_proxy_token_cost_total`**: Estimated token costs in cents
- **`ollama_proxy_request_size_bytes`**: Request payload sizes
- **`ollama_proxy_response_size_bytes`**: Response payload sizes

### 2. OpenAI API Compatibility

The proxy now supports OpenAI-compatible endpoints, allowing you to use OpenAI SDKs and tools with Ollama.

#### Supported Endpoints

- **`POST /v1/chat/completions`**: OpenAI chat completions API
- **`POST /v1/completions`**: OpenAI completions API (legacy)
- **`GET /v1/models`**: List available models

#### Model Mapping

The proxy automatically maps OpenAI model names to Ollama equivalents:

| OpenAI Model | Ollama Model |
|-------------|--------------|
| gpt-4 | llama2:70b |
| gpt-4-turbo | llama2:70b |
| gpt-3.5-turbo | llama2:13b |
| gpt-3.5-turbo-16k | llama2:13b |
| text-davinci-003 | llama2:7b |
| text-davinci-002 | llama2:7b |
| code-davinci-002 | codellama:7b |
| text-embedding-ada-002 | nomic-embed-text |

You can also use Ollama model names directly.

## Usage Examples

### Using OpenAI Python SDK

```python
from openai import OpenAI

# Point to your proxy
client = OpenAI(
    api_key="not-needed",  # Ollama doesn't require API keys
    base_url="http://localhost:11434/v1"
)

# Chat completion
response = client.chat.completions.create(
    model="gpt-3.5-turbo",  # Will use llama2:13b
    messages=[
        {"role": "system", "content": "You are a helpful assistant."},
        {"role": "user", "content": "Hello!"}
    ],
    stream=True
)

for chunk in response:
    print(chunk.choices[0].delta.content, end="")
```

### Using cURL

```bash
# OpenAI-compatible request
curl http://localhost:11434/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": false
  }'

# Native Ollama request (still supported)
curl http://localhost:11434/api/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama2:13b",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ],
    "stream": false
  }'
```

## Metrics Dashboard

### Prometheus Queries

Here are some useful Prometheus queries for monitoring AI workloads:

```promql
# Token throughput by model
rate(ollama_proxy_generated_tokens_total[5m])

# Average tokens per second
avg by (model) (ollama_proxy_tokens_per_second)

# Request latency percentiles
histogram_quantile(0.95, rate(ollama_proxy_request_duration_seconds_bucket[5m]))

# Time to first token p95
histogram_quantile(0.95, rate(ollama_proxy_time_to_first_token_seconds_bucket[5m]))

# Cost per user
sum by (user) (rate(ollama_proxy_token_cost_total[1h]))

# Active requests
sum(ollama_proxy_active_requests)
```

### Grafana Dashboard

The proxy provides enhanced metrics that can be visualized in Grafana:

1. **Token Metrics Panel**
   - Token generation rate
   - Tokens per second by model
   - Context length distribution

2. **Latency Panel**
   - Request duration histogram
   - Time to first token
   - Model load times

3. **Cost Tracking Panel**
   - Token costs by user
   - Cost trends over time
   - Cost by model

4. **Request Analytics**
   - Requests per user
   - Request/response sizes
   - Error rates

## Configuration

### Environment Variables

- `OLLAMA_PROXY_PORT`: Proxy port (default: 11434)
- `OLLAMA_METRICS_PORT`: Metrics port (default: 9090)
- `OLLAMA_HOST`: Ollama backend host (default: localhost)
- `OLLAMA_PORT`: Ollama backend port (default: 11434)

### Token Cost Configuration

Token costs are configured in the metrics collector. To modify pricing, edit the `getTokenCost` function in `internal/metrics/metrics.go`.

## Best Practices

1. **Request IDs**: Each request is assigned a unique ID (returned in `X-Request-ID` header) for tracking
2. **User Tracking**: Use the `user` field in OpenAI requests to track usage per user
3. **Model Selection**: Use appropriate models for your use case to optimize cost/performance
4. **Streaming**: Enable streaming for better user experience and accurate TTFT metrics
5. **Monitoring**: Set up alerts for high latency, error rates, and cost thresholds

## Troubleshooting

### Common Issues

1. **Model not found**: Ensure the Ollama model is pulled locally
2. **Slow responses**: Check model load times and available resources
3. **High costs**: Review token usage patterns and optimize prompts

### Debug Headers

The proxy adds several headers for debugging:
- `X-Request-ID`: Unique request identifier
- `X-Model-Used`: Actual Ollama model used
- `X-Tokens-Prompt`: Prompt token count
- `X-Tokens-Generated`: Generated token count