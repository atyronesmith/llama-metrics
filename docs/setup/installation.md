# 🚀 Quick Setup Guide for Mac M-Series

**One-command setup for complete Ollama monitoring stack!**

## Prerequisites

- **Mac with M-series chip** (M1, M2, M3, M4)
- **macOS 12.0+**
- **Python 3.8+** (usually pre-installed)

## 🎯 One-Command Install

```bash
git clone <repository-url>
cd llamastack-prometheus
make setup && make start
```

**That's it!** This will:

✅ **Auto-detect** your Mac M-series system  
✅ **Install Ollama** if not present  
✅ **Install Prometheus** via Homebrew  
✅ **Download phi3:mini model** automatically  
✅ **Setup Python environment** with all dependencies  
✅ **Start all services** (Ollama, Proxy, Prometheus, Dashboard)  

## 🎛️ What You Get

After setup completes, you'll have:

- **📊 Real-time Dashboard**: http://localhost:3001
- **📈 Prometheus UI**: http://localhost:9090  
- **🔧 Metrics API**: http://localhost:8001/metrics
- **🤖 Ollama API**: http://localhost:11434

## 🧪 Test Your Setup

Generate some test traffic:

```bash
make traffic        # Interactive traffic generator
make traffic-quick  # Quick 10-request test
make load-test      # Advanced load testing scenarios
```

## 📱 Dashboard Features

- **Real-time LLM monitoring** with AI-generated status summaries
- **Performance metrics**: latency, tokens/sec, GPU usage, power consumption
- **Queue visualization** and load balancing insights
- **Ollama health monitoring** with response times
- **High-load mode detection** and timeout handling

## 🛠️ Manual Install Steps

If automatic setup fails, follow these steps:

### 1. Install Ollama
```bash
# Download from https://ollama.ai or use:
curl -fsSL https://ollama.ai/install.sh | sh
```

### 2. Install Prometheus
```bash
# Install Homebrew first if needed:
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"

# Then install Prometheus:
brew install prometheus
```

### 3. Run Setup
```bash
make quick-setup && make start
```

## ⚡ Quick Commands

```bash
make help           # Show all available commands
make setup          # Full automated setup
make start          # Start all services
make stop           # Stop all services  
make restart        # Restart all services
make status         # Check service status
make dashboard      # Start just the dashboard
make traffic        # Generate test traffic
make logs           # View service logs
```

## 🔧 Troubleshooting

**Port conflicts?**
```bash
make stop && make start
```

**Ollama not responding?**
```bash
make restart-ollama
```

**Dashboard not loading?**
```bash
make stop-dashboard && make start-dashboard
```

**Check service status:**
```bash
make status
```

## 🎯 Performance Optimizations for M-Series

- **GPU acceleration** automatically enabled for phi3:mini
- **Memory-efficient** settings optimized for Apple Silicon
- **Low-power mode** detection during idle periods
- **Concurrent request limiting** to prevent overload (max 5 concurrent)

## 🚦 Service Architecture

```
[Traffic Generator] → [Monitoring Proxy:11435] → [Ollama:11434]
                            ↓
[Dashboard:3001] ← [Prometheus:9090] ← [Metrics:8001]
```

## 💡 Pro Tips

- Use `make traffic-stress` for performance testing
- Monitor GPU usage in Activity Monitor during load tests
- Dashboard shows "High Load Mode" during stress tests
- Ollama status appears in the top navigation bar

---

**Need help?** Check the main README or open an issue!