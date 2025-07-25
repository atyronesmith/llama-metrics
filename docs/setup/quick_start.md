# Quick Start Guide

Get llama-metrics running in 5 minutes!

## 🚀 Prerequisites

- **Ollama**: [Install Ollama](https://ollama.ai/download)
- **Python**: 3.8+ installed
- **Git**: For cloning the repository
- **Docker/Podman**: For Prometheus (optional)

## ⚡ 1-Command Setup

```bash
# Clone and setup
git clone https://github.com/atyronesmith/llama-metrics.git
cd llama-metrics
./scripts/deployment/install.sh --yes
```

This automatically:
- ✅ Creates Python virtual environment
- ✅ Installs all dependencies
- ✅ Downloads default model (phi3:mini)
- ✅ Configures services

## 🏃‍♂️ Start Services

```bash
# Start Ollama (if not running)
ollama serve &

# Start monitoring stack
make start
```

This launches:
- **Ollama Proxy**: `http://localhost:11435` (metrics on `:8001`)
- **Dashboard**: `http://localhost:3001`
- **Health Checker**: `http://localhost:8080`
- **Prometheus**: `http://localhost:9090`

## 📊 Generate Test Traffic

```bash
# Quick 10-request test
make traffic-quick

# Interactive traffic generator
make traffic

# Continuous background traffic
make traffic-continuous &
```

## 🎯 View Results

1. **Dashboard**: http://localhost:3001
   - Real-time metrics and charts
   - AI-powered analysis
   - System monitoring

2. **Prometheus**: http://localhost:9090
   - Raw metrics data
   - Query interface
   - Alert rules

3. **Metrics Endpoint**: http://localhost:8001/metrics
   - Prometheus-format metrics
   - Direct data access

## ✅ Verify Everything Works

```bash
# Check all services
make verify

# Run health check
python scripts/monitoring/health_check.py

# Check metrics collection
curl http://localhost:8001/metrics | head -20
```

## 🎉 You're Ready!

Your monitoring system is now:
- 📈 **Collecting metrics** from Ollama requests
- 🎛️ **Displaying real-time data** in the dashboard
- 🔍 **Storing historical data** in Prometheus
- 🤖 **Analyzing performance** with AI insights

## 🔄 Next Steps

### Customize Your Setup
- **[Configuration Guide](configuration.md)**: Customize settings
- **[Performance Tuning](performance.md)**: Optimize for your hardware
- **[Traffic Generation](traffic_generation.md)**: Advanced testing scenarios

### Explore Features
- **[API Reference](../api/overview.md)**: Integrate with other tools
- **[Architecture Overview](../architecture/overview.md)**: Understand the system
- **[Development Guide](../development/contributing.md)**: Contribute improvements

## 🆘 Troubleshooting

### Common Issues

**Ollama not starting**
```bash
# Check if already running
ps aux | grep ollama

# Start manually
ollama serve
```

**Port conflicts**
```bash
# Check what's using ports
lsof -i :11434  # Ollama
lsof -i :11435  # Proxy
lsof -i :3001   # Dashboard
```

**Virtual environment issues**
```bash
# Recreate if needed
rm -rf venv
./scripts/deployment/install.sh
```

**No metrics showing**
```bash
# Verify proxy is collecting metrics
curl http://localhost:8001/metrics

# Check Prometheus targets
open http://localhost:9090/targets
```

### Get Help
- 📖 **[Troubleshooting Guide](../development/troubleshooting.md)**: Detailed solutions
- 🐛 **[GitHub Issues](https://github.com/atyronesmith/llama-metrics/issues)**: Report problems
- 💬 **[Discussions](https://github.com/atyronesmith/llama-metrics/discussions)**: Ask questions

---

**Need more detail?** See the [Complete Installation Guide](installation.md)