# Llama Metrics - Ollama Monitoring with Prometheus

A comprehensive monitoring solution for Ollama AI models using Prometheus metrics collection and traffic generation for performance testing.

## 🎯 Overview

This project provides a complete monitoring stack for Ollama AI models, including:

- **Traffic Generator**: Sends diverse questions to Ollama models for load testing
- **Metrics Server**: Exposes Prometheus metrics for monitoring AI model performance
- **Question Database**: 1000+ curated questions across 10 categories
- **Prometheus Integration**: Ready-to-use Prometheus configuration

## 🏗️ Architecture

```
[Traffic Generator] → [Ollama API] → [Metrics Server] → [Prometheus] → [Dashboard]
        ↓                ↓              ↓              ↓           ↓
   1000 Questions    AI Responses    Metrics Collection  Scraping   Visualization
```

## ✨ Example Application Architecture

The following diagrams illustrate how the core components of Llama Metrics—the proxy, monitoring server, and dashboard—work together to create a complete, observable AI application.

### Component Interaction Flow

This diagram shows how a user request travels through the system, from the dashboard to the Ollama backend and back.

```mermaid
graph TD
    subgraph "User Interaction"
        User([User])
    end

    subgraph "Application Services (Go)"
        Dashboard[Web Dashboard]
        Proxy[Monitoring Proxy]
    end

    subgraph "AI Backend"
        Ollama[Ollama API]
    end

    User -- "1. Sends request" --> Dashboard
    Dashboard -- "2. Forwards to Proxy" --> Proxy
    Proxy -- "3. Manages queue & forwards to Ollama" --> Ollama
    Ollama -- "4. Generates response" --> Proxy
    Proxy -- "5. Returns response" --> Dashboard
    Dashboard -- "6. Displays result" --> User

    style User fill:#cde4ff
    style Dashboard fill:#d1ecf1
    style Proxy fill:#d1ecf1
    style Ollama fill:#d4edda
```

### Monitoring Data Flow

This diagram illustrates how monitoring data is generated, collected, and displayed.

```mermaid
graph TD
    subgraph "Application Services (Go)"
        Proxy[Monitoring Proxy]
    end

    subgraph "Monitoring Stack"
        Prometheus[Prometheus]
        Dashboard[Web Dashboard]
    end

    subgraph "User Interaction"
        Operator([Operator/User])
    end

    Proxy -- "1. Exposes /metrics endpoint" --> Prometheus
    Prometheus -- "2. Scrapes metrics periodically" --> Proxy
    Dashboard -- "3. Queries metrics from Prometheus (PromQL)" --> Prometheus
    Prometheus -- "4. Returns time-series data" --> Dashboard
    Operator -- "5. Views real-time dashboard" --> Dashboard

    style Proxy fill:#cde4ff
    style Prometheus fill:#f8d7da
    style Dashboard fill:#d1ecf1
    style Operator fill:#cde4ff
```

## 📊 Dashboard

The real-time dashboard provides comprehensive monitoring of your Ollama LLM performance:

![Ollama LLM Dashboard](docs/images/dashboard-screenshot.png)

### Dashboard Features:
- **Real-time Metrics**: Request rate, success rate, and token generation speed
- **System Monitoring**: GPU usage, power consumption, and memory utilization
- **Queue Analytics**: Request queue size, processing rate, and efficiency
- **Latency Tracking**: P50, P75, P95, and P99 percentiles
- **Time Series Graphs**: Historical view of token generation and memory usage
- **AI Status**: Automatic health analysis with actionable insights

## 📁 Project Structure

```
├── services/                     # Go services
│   ├── shared/                   # Shared packages
│   ├── proxy/                    # Monitoring proxy
│   ├── dashboard/                # Web dashboard
│   └── health/                   # Health checker
├── scripts/                      # Automation scripts
│   ├── traffic/                  # Traffic generation
│   ├── monitoring/               # Monitoring tools
│   └── deployment/               # Installation scripts
├── config/                       # Configuration files
│   ├── prometheus/
│   │   └── prometheus.yml        # Prometheus configuration
│   ├── services/                 # Service-specific configs
│   └── alerts/                   # Alert rules
├── docs/                         # Documentation
│   ├── setup/                    # Installation guides
│   ├── architecture/             # System design
│   ├── api/                      # API reference
│   └── development/              # Contributing guides
├── test/                         # Test files and data
│   ├── unit/                     # Unit tests
│   └── data/                     # Test data
│       └── questions/            # Question categories (10 files)
└── README.md                     # This file
```

## 🚀 Quick Start

### Prerequisites

- **Ollama** installed and running
- **Python 3.8+**
- **Podman** or **Docker** (for Prometheus)

### 1. Install Dependencies

```bash
# For traffic generator only (lightweight)
pip install -r requirements_traffic.txt

# OR for full LlamaIndex features
pip install -r requirements_app.txt
```

### 2. Start Ollama

```bash
ollama serve
ollama pull phi3:mini  # or your preferred model
```

### 3. Start Metrics Server

```bash
# Simple metrics server (recommended)
python simple_metrics_server.py

# OR LlamaIndex-based server (advanced)
python app.py
```

### 4. Start Prometheus

```bash
./run_prometheus.sh
```

### 5. Generate Traffic

```bash
# Basic usage
python traffic_generator.py

# With options
python traffic_generator.py --max 50 --delay 2
```

### 6. View Monitoring

- **Prometheus Dashboard**: http://localhost:9090
- **Raw Metrics**: http://localhost:8000/metrics

## 🎯 Traffic Generator

### Features

- **1000+ Questions** across 10 diverse categories
- **Random Selection** for varied testing
- **Serialized Processing** (waits for each response)
- **Configurable Parameters** (model, delay, limits)
- **Real-time Statistics** and logging

### Usage Examples

```bash
# Default usage (unlimited questions, 1s delay)
python traffic_generator.py

# Limited test run
python traffic_generator.py --max 10 --delay 3

# Fast load testing
python traffic_generator.py --delay 0.5

# Different model
python traffic_generator.py --model llama2
```

### Question Categories

- **General Knowledge** (100) - Basic facts and trivia
- **Science** (100) - Physics, chemistry, biology
- **Technology** (100) - Computing, software, modern tech
- **History** (100) - World events, civilizations
- **Geography** (100) - Countries, capitals, physical features
- **Sports** (100) - Various sports, rules, athletes
- **Entertainment** (100) - Movies, music, TV, celebrities
- **Literature** (100) - Books, authors, characters
- **Philosophy** (100) - Concepts, thinkers, theories
- **Food** (100) - Cuisine, cooking, ingredients

## 📊 Metrics Collected

### Ollama Metrics

- `ollama_requests_total{model, status}` - Total requests by model and status
- `ollama_request_duration_seconds{model}` - Request latency histogram
- `ollama_active_requests{model}` - Current active requests
- `ollama_model_info` - Available model information

### Traffic Metrics

- `traffic_questions_asked_total{category}` - Questions by category
- Standard Python metrics (GC, memory, etc.)

## 📖 Documentation

For detailed documentation, visit the [`docs/`](docs/) directory:

- **[Quick Start](docs/setup/quick_start.md)**: Get up and running in 5 minutes
- **[Installation Guide](docs/setup/installation.md)**: Complete setup instructions
- **[Architecture Overview](docs/architecture/overview.md)**: System design and components
- **[API Reference](docs/api/overview.md)**: API documentation and integration
- **[Contributing Guide](docs/development/contributing.md)**: Development guidelines

## 🛠️ Configuration

Configuration is organized in the [`config/`](config/) directory:

- **[Services](config/services/)**: Service-specific configurations
- **[Prometheus](config/prometheus/)**: Monitoring configuration
- **[Alerts](config/alerts/)**: Alert rules and notifications

### Quick Configuration

The main configuration file is [`config/llama-metrics.yml`](config/llama-metrics.yml) with settings for:

- Service ports and endpoints
- Model configurations
- Monitoring thresholds
- Performance tuning options

## 🔧 Advanced Usage

### Custom Question Categories

Add new question files to the `questions/` directory:

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "category": "Your Category",
  "questions": [
    "Your question 1?",
    "Your question 2?",
    ...
  ]
}
```

### Prometheus Queries

Example queries for the Prometheus dashboard:

```promql
# Request rate
rate(ollama_requests_total[5m])

# Average latency
rate(ollama_request_duration_seconds_sum[5m]) / rate(ollama_request_duration_seconds_count[5m])

# Questions by category
rate(traffic_questions_asked_total[5m])

# Error rate
rate(ollama_requests_total{status="error"}[5m])
```

## ⚡ Performance Tuning

### Preventing Ollama Overload

To avoid resource contention and multiple model runner processes:

```bash
# Set before starting Ollama
export OLLAMA_NUM_PARALLEL=2      # Limit parallel model instances
export OLLAMA_MAX_LOADED_MODELS=2 # Limit loaded models

# Or use the Makefile targets which include these settings
make start-ollama
```

### Proxy Configuration

The monitoring proxy includes a request queue to handle load spikes:

```bash
# Default settings (applied automatically with make start-proxy)
--max-concurrency 4    # Worker threads (matches Ollama capacity)
--max-queue-size 100   # Request buffer size

# Custom settings
export MAX_CONCURRENCY=6
export MAX_QUEUE_SIZE=200
make start-proxy
```

### Quick Start with Optimized Settings

```bash
make setup          # Install dependencies
make start          # Starts all services with optimized settings
```

## 🐛 Troubleshooting

### Traffic Generator Issues

- **"Cannot connect to Ollama"**: Ensure Ollama is running (`ollama serve`)
- **"Model not found"**: Check available models (`ollama list`)
- **High error rates**: Reduce request frequency with `--delay`

### Metrics Server Issues

- **Port 8000 in use**: Stop existing processes or change port
- **Import errors**: Install dependencies (`pip install -r requirements_traffic.txt`)

### Prometheus Issues

- **Target down**: Ensure metrics server is running on port 8000
- **Container conflicts**: Stop existing containers (`podman stop prometheus`)

## 📝 Development

### Adding New Metrics

Edit `simple_metrics_server.py` to add custom metrics:

```python
NEW_METRIC = Counter('new_metric_total', 'Description')
NEW_METRIC.inc()  # Increment the metric
```

### Extending Question Categories

1. Create new JSON file in `questions/` directory
2. Follow the existing schema format
3. Traffic generator will automatically load it

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Add your improvements
4. Submit a pull request

## 📄 License

This project is open source. Feel free to use and modify as needed.

## 🔗 Related Projects

- [Ollama](https://ollama.ai/) - Local AI model runner
- [Prometheus](https://prometheus.io/) - Monitoring system
- [LlamaIndex](https://www.llamaindex.ai/) - Data framework for AI apps

---

**Happy Monitoring!** 🚀📊