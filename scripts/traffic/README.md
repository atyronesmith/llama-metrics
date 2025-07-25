# Traffic Generation Scripts

This directory contains scripts for generating traffic and load testing the Ollama monitoring system.

## Scripts Overview

### 🚀 Quick Start
```bash
# Interactive traffic generation
./scripts/traffic/run.sh

# Quick 10-request test
./scripts/traffic/run.sh --quick

# Simple continuous traffic
./scripts/traffic/simple.sh
```

## Script Descriptions

### `generator.py` - Main Traffic Generator
**Purpose**: Core Python script that generates diverse questions to test Ollama models.

**Features**:
- 1000+ curated questions across 10 categories
- Configurable model, URL, delay, and request limits
- Real-time statistics and logging
- Support for both proxy and direct Ollama connections

**Usage**:
```bash
# Basic usage
python scripts/traffic/generator.py

# Custom configuration
python scripts/traffic/generator.py --model phi3:mini --url http://localhost:11435 --delay 1.0 --max 50
```

**Arguments**:
- `--model`: Ollama model to use (default: phi3:mini)
- `--url`: Ollama API URL (default: http://localhost:11434)
- `--delay`: Delay between requests in seconds (default: 1.0)
- `--max`: Maximum number of requests (default: unlimited)

### `run.sh` - Interactive Traffic Runner
**Purpose**: User-friendly shell script with preset configurations and interactive prompts.

**Features**:
- Multiple preset modes (quick, demo, stress)
- Automatic proxy detection
- Interactive configuration
- Colored output and progress indicators

**Usage**:
```bash
# Interactive mode
./scripts/traffic/run.sh

# Preset modes
./scripts/traffic/run.sh --quick    # 10 requests
./scripts/traffic/run.sh --demo     # 50 requests
./scripts/traffic/run.sh --stress   # 1000 requests
```

### `simple.sh` - Simple Continuous Traffic
**Purpose**: Minimal script for basic continuous traffic generation.

**Features**:
- No configuration prompts
- Uses sensible defaults
- Connects through monitoring proxy
- Easy to modify for custom needs

**Usage**:
```bash
./scripts/traffic/simple.sh
```

### `direct.sh` - Direct Ollama Traffic
**Purpose**: Generate traffic directly to Ollama, bypassing the monitoring proxy.

**Use Cases**:
- Testing Ollama performance without proxy overhead
- Comparing proxy vs direct performance
- Debugging proxy issues

**Usage**:
```bash
./scripts/traffic/direct.sh
```

### `high_performance.py` - Advanced Load Tester
**Purpose**: High-performance, concurrent load testing with advanced metrics.

**Features**:
- Concurrent request handling
- Detailed performance metrics
- Configurable concurrency levels
- Advanced statistics and reporting

**Usage**:
```bash
python scripts/traffic/high_performance.py --threads 10 --requests 1000
```

### `scenarios.sh` - Load Test Scenarios
**Purpose**: Pre-configured load testing scenarios for different use cases.

**Features**:
- Multiple test scenarios (light, medium, heavy)
- Interactive scenario selection
- Automated test execution
- Results reporting

**Usage**:
```bash
./scripts/traffic/scenarios.sh
```

## Prerequisites

- **Ollama**: Must be running (`ollama serve`)
- **Python Virtual Environment**: Activate with `source venv/bin/activate`
- **Model**: Download required model (`ollama pull phi3:mini`)

## Configuration

### Environment Variables
- `OLLAMA_HOST`: Ollama server host (default: localhost)
- `OLLAMA_PORT`: Ollama server port (default: 11434)
- `PROXY_PORT`: Monitoring proxy port (default: 11435)

### Question Categories
Questions are loaded from `test/data/questions/` directory:
- General Knowledge, Science, Technology
- History, Geography, Sports
- Entertainment, Literature, Philosophy, Food

## Monitoring Integration

### With Proxy (Recommended)
- URL: `http://localhost:11435`
- Metrics: Automatically collected
- Dashboard: Real-time visualization

### Direct to Ollama
- URL: `http://localhost:11434`
- Metrics: Not collected
- Use for performance comparison

## Troubleshooting

### Common Issues

**"Connection refused"**
```bash
# Check if Ollama is running
ollama list

# Start Ollama if needed
ollama serve
```

**"Model not found"**
```bash
# Pull the required model
ollama pull phi3:mini
```

**"Virtual environment not found"**
```bash
# Create and activate virtual environment
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

**High error rates**
- Reduce request frequency with higher `--delay` values
- Check Ollama resource usage
- Verify model is loaded properly

## Best Practices

1. **Start Small**: Begin with `--quick` mode to verify setup
2. **Monitor Resources**: Watch CPU/GPU usage during load tests
3. **Use Proxy**: Route through monitoring proxy for metrics collection
4. **Gradual Load**: Increase load gradually to avoid overwhelming Ollama
5. **Clean Shutdown**: Use Ctrl+C to stop scripts gracefully

## Integration with Makefile

These scripts are integrated with the main project Makefile:

```bash
make traffic          # Interactive traffic generation
make traffic-quick    # Quick test (10 requests)
make traffic-demo     # Demo (50 requests)
make traffic-stress   # Stress test (1000 requests)
make traffic-continuous # Continuous simple traffic
```