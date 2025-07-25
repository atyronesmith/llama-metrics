# Ollama LLM Dashboard (Go)

A real-time monitoring dashboard for Ollama LLM performance metrics, rewritten in Go for improved performance and lower resource usage.

## Features

- **Real-time Metrics**: Live updates via WebSocket connection
- **AI-Generated Status**: Intelligent system status summaries using Ollama
- **Performance Charts**: Token generation, memory usage, GPU utilization, and power consumption
- **Responsive UI**: Bootstrap-based interface with dark mode support
- **Low Resource Usage**: Efficient Go implementation

## Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│   Dashboard     │────▶│   Prometheus     │────▶│ Monitoring  │
│  (Port 3001)    │     │  (Port 9099)     │     │   Proxy     │
└─────────────────┘     └──────────────────┘     └─────────────┘
        │                                                │
        └────────────────────────────────────────────────┘
                          WebSocket
```

## Prerequisites

- Go 1.21 or higher
- Prometheus running on port 9099
- Ollama running on port 11434
- Monitoring proxy collecting metrics

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd dashboard

# Install dependencies
make deps

# Build the application
make build
```

## Configuration

The dashboard can be configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `DASHBOARD_PORT` | 3001 | Port for the dashboard server |
| `DASHBOARD_ENV` | development | Environment (development/production) |
| `PROMETHEUS_URL` | http://localhost:9099 | Prometheus server URL |
| `OLLAMA_URL` | http://localhost:11434 | Ollama server URL |

## Usage

### Development Mode

```bash
# Run with hot reload
make dev

# Or run directly
make run-dev
```

### Production Mode

```bash
# Build and run
make build
./build/dashboard

# Or use make
make run-prod
```

### Other Commands

```bash
# Run tests
make test

# Format code
make fmt

# Run linter
make lint

# Clean build artifacts
make clean

# Build for multiple platforms
make build-all
```

## API Endpoints

- `GET /` - Main dashboard page
- `GET /ws` - WebSocket endpoint for real-time updates
- `GET /api/metrics` - Get all metrics
- `GET /api/metrics/summary` - Get summary metrics
- `GET /api/metrics/timeseries` - Get time series data for charts
- `GET /api/status` - Get AI-generated status
- `GET /api/health` - Health check endpoint

## WebSocket Protocol

The dashboard uses native WebSocket for real-time updates. Messages are JSON-formatted:

```json
{
    "summary": {
        "request_rate": 2.5,
        "avg_latency": 1.2,
        "tokens_per_second": 45.3,
        ...
    },
    "latency_percentiles": {
        "p50": 0.8,
        "p95": 2.1,
        ...
    },
    "ai_status": "System operating normally...",
    "is_ai_generated": true,
    "timestamp": "2024-01-15T10:30:00Z"
}
```

## Project Structure

```
dashboard/
├── cmd/
│   └── dashboard/
│       └── main.go           # Application entry point
├── internal/
│   ├── handlers/            # HTTP handlers
│   ├── metrics/            # Prometheus metrics collection
│   └── websocket/          # WebSocket hub and client management
├── pkg/
│   └── config/             # Configuration management
├── web/
│   ├── static/             # Static assets (CSS, JS)
│   └── templates/          # HTML templates
├── Makefile               # Build commands
├── go.mod                 # Go module definition
└── README.md             # This file
```

## Performance

The Go implementation offers several advantages over the Python version:

- **Lower Memory Usage**: ~10MB vs ~100MB for Python
- **Faster Startup**: < 1 second vs 3-5 seconds
- **Better Concurrency**: Native goroutines for handling multiple connections
- **Lower CPU Usage**: More efficient WebSocket handling

## Development

### Adding New Metrics

1. Update the Prometheus queries in `internal/metrics/collector.go`
2. Add the metric to the summary or time series methods
3. Update the frontend in `web/templates/dashboard.html` to display the new metric

### Modifying the UI

1. Edit `web/templates/dashboard.html`
2. Add any static assets to `web/static/`
3. The server automatically serves files from these directories

## Troubleshooting

### Connection Issues

- Verify Prometheus is running: `curl http://localhost:9099/-/healthy`
- Check Ollama is accessible: `curl http://localhost:11434/api/tags`
- Ensure the monitoring proxy is exposing metrics

### No Metrics Displayed

- Check browser console for errors
- Verify WebSocket connection is established
- Check Prometheus queries are returning data

### Build Issues

- Ensure Go 1.21+ is installed: `go version`
- Update dependencies: `go mod tidy`
- Clear module cache: `go clean -modcache`

## License

[Add your license here]