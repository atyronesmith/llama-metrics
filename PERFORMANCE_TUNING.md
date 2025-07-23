# Performance Tuning Guide

This guide provides optimal configuration settings to prevent system overload and ensure smooth operation.

## Issue: Ollama Resource Contention

When Ollama receives too many concurrent requests, it may spawn multiple model runner processes, leading to:
- High memory usage (2-3GB per runner)
- Resource contention
- System unresponsiveness
- Request timeouts

## Recommended Settings

### 1. Ollama Configuration

```bash
# Limit parallel processing
export OLLAMA_NUM_PARALLEL=2         # Max concurrent model evaluations
export OLLAMA_MAX_LOADED_MODELS=2    # Max models in memory
export OLLAMA_MAX_QUEUE=10           # Request queue size

# Start Ollama with these settings
ollama serve
```

### 2. Proxy Configuration

The monitoring proxy includes a request queue to buffer incoming requests:

```bash
# Default settings (optimized for Ollama)
./proxy/build/ollama-proxy \
    --max-concurrency 4 \     # Worker threads
    --max-queue-size 100      # Request buffer
```

### 3. Environment Variables

```bash
# Set in your shell profile or .env file
export OLLAMA_NUM_PARALLEL=2
export OLLAMA_MAX_LOADED_MODELS=2
export MAX_CONCURRENCY=4
export MAX_QUEUE_SIZE=100
```

### 4. Using Make Commands

The Makefile includes these optimizations automatically:

```bash
make start          # Starts all services with optimal settings
make start-ollama   # Starts Ollama with OLLAMA_NUM_PARALLEL=2
make start-proxy    # Starts proxy with --max-concurrency 4
```

## Load Testing Guidelines

When running load tests:

```bash
# Moderate load (recommended)
make traffic-demo   # 10 requests, 1s delay

# Higher load (monitor carefully)
make load-test-constant   # 10 RPS, 10 concurrent

# Stress testing (may cause issues)
make load-test-queue      # 25 RPS, 100 concurrent
```

## Monitoring Performance

Watch these metrics in the dashboard (http://localhost:3001):

1. **Queue Size**: Should stay below 20 under normal load
2. **Active Requests**: Should not exceed concurrency limit (4)
3. **Memory Usage**: Monitor for sudden spikes
4. **Processing Rate**: Should match request rate

## Troubleshooting

### Symptoms of Overload
- Multiple `ollama runner` processes in `ps aux`
- High memory usage (>10GB)
- Requests timing out (503 errors)
- Queue size growing continuously

### Recovery Steps
1. Stop traffic: `make stop-traffic`
2. Kill runners: `pkill -f "ollama runner"`
3. Restart Ollama: `make stop-ollama && make start-ollama`
4. Clear proxy queue: `make stop-proxy && make start-proxy`

### Prevention
- Use recommended settings above
- Monitor dashboard during load tests
- Gradually increase load
- Set up alerts for high queue size

## Advanced Tuning

### For High-Performance Systems
```bash
# If you have ample resources (32GB+ RAM)
export OLLAMA_NUM_PARALLEL=4
export MAX_CONCURRENCY=8
```

### For Resource-Constrained Systems
```bash
# For systems with limited RAM (<16GB)
export OLLAMA_NUM_PARALLEL=1
export MAX_CONCURRENCY=2
```

## Best Practices

1. **Start Conservative**: Begin with default settings
2. **Monitor Metrics**: Watch dashboard during operation
3. **Gradual Scaling**: Increase limits incrementally
4. **Test Thoroughly**: Run load tests before production
5. **Document Changes**: Keep track of what works for your system