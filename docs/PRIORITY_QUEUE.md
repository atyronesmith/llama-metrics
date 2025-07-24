# Priority Queue Implementation

The Ollama Monitoring Proxy now supports priority queuing with two levels: **normal** and **high** priority.

## How It Works

### Priority Levels
- **Normal Priority (0)**: Default priority for all requests
- **High Priority (1)**: Elevated priority for urgent requests

### Request Header
Priority is controlled via the `X-Priority` HTTP header:
```bash
# High priority request
curl -H "X-Priority: high" -H "Content-Type: application/json" \
  -d '{"model": "phi3:mini", "prompt": "Hello"}' \
  http://localhost:11435/api/generate

# Normal priority (default, header optional)
curl -H "Content-Type: application/json" \
  -d '{"model": "phi3:mini", "prompt": "Hello"}' \
  http://localhost:11435/api/generate
```

### Queue Behavior
1. High priority requests are always processed before normal priority requests
2. Within the same priority level, requests are processed in FIFO order
3. Priority doesn't affect requests already being processed
4. The queue respects the maximum queue size for all priorities combined

## OpenAI API Compatibility

The OpenAI API protocol doesn't have built-in priority support, so we've implemented it using a custom header approach:

- **No multiple ports needed**: Single proxy port handles all priorities
- **Backward compatible**: Requests without the header default to normal priority
- **Works with all endpoints**: `/api/generate`, `/api/chat`, and OpenAI-compatible endpoints

### Example with OpenAI Python Client
```python
import openai

# Configure client
openai.api_base = "http://localhost:11435/v1"
openai.api_key = "not-needed"

# High priority request
response = openai.ChatCompletion.create(
    model="phi3:mini",
    messages=[{"role": "user", "content": "Urgent: What is the status?"}],
    headers={"X-Priority": "high"}  # Note: Requires custom client modification
)
```

## Metrics

New Prometheus metrics for monitoring priority queues:

- `ollama_proxy_queue_high_priority_count`: Current number of high priority requests in queue
- `ollama_proxy_queue_normal_priority_count`: Current number of normal priority requests in queue
- `ollama_proxy_queue_high_priority_wait_time_seconds`: Wait time histogram for high priority requests
- `ollama_proxy_queue_normal_priority_wait_time_seconds`: Wait time histogram for normal priority requests
- `ollama_proxy_high_priority_request_duration_seconds`: Request latency histogram for high priority requests
- `ollama_proxy_normal_priority_request_duration_seconds`: Request latency histogram for normal priority requests

### Dashboard Display

The monitoring dashboard now shows:
- **Regular Latency Percentiles**: P50, P75, P95, P99 for all requests
- **High Priority Latency Percentiles**: P50, P75, P95, P99 specifically for high priority requests

This allows you to verify that high priority requests are indeed being processed faster than normal priority ones.

## Testing

Use the provided test script to verify priority queue behavior:

```bash
# Run the priority queue test
python scripts/test_priority_queue.py
```

This script will:
1. Send 10 normal priority requests
2. Wait briefly, then send 5 high priority requests
3. Show completion times demonstrating that high priority requests jump the queue

## Configuration

The priority queue uses the same configuration as the standard queue:
- `--max-queue-size`: Maximum total requests in queue (default: 100)
- `--max-concurrency`: Maximum concurrent request processors (default: 5)

## Use Cases

1. **Production Systems**: Prioritize user-facing requests over batch processing
2. **Multi-tenant**: Give premium users high priority access
3. **Emergency Responses**: Ensure critical queries get processed first
4. **Load Management**: Maintain responsiveness during high load
5. **Dashboard AI Summary**: The monitoring dashboard automatically uses high priority for AI-generated summaries to ensure quick status updates

## Implementation Details

The priority queue is implemented using Go's `container/heap` package:
- Efficient O(log n) insertion and removal
- Stable ordering within priority levels
- Thread-safe with mutex protection
- Integrated with existing metrics collection

## Best Practices

1. **Use sparingly**: Reserve high priority for truly urgent requests
2. **Monitor metrics**: Watch queue wait times for both priority levels
3. **Set limits**: Consider implementing rate limits per priority level
4. **Test thoroughly**: Verify your priority logic before production use