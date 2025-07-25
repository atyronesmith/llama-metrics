# System Architecture Overview

This document provides a comprehensive overview of the llama-metrics system architecture, including components, data flow, and design decisions.

## 🏗️ High-Level Architecture

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────┐
│   Dashboard     │────▶│   Prometheus     │────▶│  Proxy      │
│  (Port 3001)    │     │  (Port 9090)     │     │ (Port 8001) │
└─────────────────┘     └──────────────────┘     └─────────────┘
                                                          │
                              ┌───────────────────────────┴───────┐
                              │                                   │
                        ┌─────▼─────┐                    ┌───────▼──────┐
                        │  Ollama   │                    │ Health Check │
                        │(Port 11434)│                    │ (Port 8080)  │
                        └───────────┘                    └──────────────┘
```

## 🔧 Core Components

### 1. Ollama Monitoring Proxy
**Location**: `services/proxy/`
**Ports**: 11435 (proxy), 8001 (metrics)

**Purpose**: Intercepts and monitors all requests to Ollama while forwarding them transparently.

**Key Features**:
- Request/response logging and metrics collection
- Queue management for concurrent requests
- Token counting and performance tracking
- System resource monitoring (CPU, memory, GPU)
- Health check endpoints

**Data Collected**:
- Request latency and throughput
- Token generation rates
- Model loading times
- Error rates and types
- System resource utilization

### 2. Real-Time Dashboard
**Location**: `services/dashboard/`
**Port**: 3001

**Purpose**: Provides real-time visualization of Ollama performance and system metrics.

**Key Features**:
- Live metric charts and graphs
- WebSocket-based real-time updates
- AI-powered performance analysis
- Historical trend visualization
- System health indicators

**Technologies**:
- Go-based backend with Gin framework
- WebSocket for real-time communication
- HTML/CSS/JavaScript frontend
- Prometheus query integration

### 3. Health Monitoring Service
**Location**: `services/health/`
**Port**: 8080

**Purpose**: Continuous health monitoring and automated issue detection.

**Key Features**:
- Service availability monitoring
- Resource threshold checking
- AI-powered health analysis
- Automated alerting
- Performance degradation detection

**Monitored Services**:
- Ollama API availability
- Proxy service health
- Dashboard responsiveness
- Prometheus data collection

### 4. Prometheus Metrics Storage
**Port**: 9090

**Purpose**: Time-series database for metrics storage and querying.

**Configuration**:
- Multiple scrape jobs for different services
- Optimized retention policies
- Alert rule integration
- Service discovery ready

## 📊 Data Flow Architecture

### Request Flow
```
Client Request → Proxy (11435) → Ollama (11434) → Response
      │                │                │            │
      │                ▼                │            │
      │         Metrics Collection      │            │
      │                │                │            │
      │                ▼                │            │
      │         Prometheus (8001)       │            │
      │                │                │            │
      │                ▼                │            │
      └─────────→ Dashboard (3001) ◄────┘            │
                       │                              │
                       ▼                              │
               Real-time Updates ◄───────────────────┘
```

### Metrics Collection Flow
```
1. Proxy intercepts request
2. Starts timer and queue tracking
3. Forwards to Ollama
4. Ollama processes request
5. Proxy receives response
6. Calculates metrics (latency, tokens, etc.)
7. Updates Prometheus metrics
8. Dashboard queries Prometheus
9. Real-time updates pushed to UI
```

## 🏛️ Design Principles

### 1. Non-Intrusive Monitoring
- **Transparent Proxy**: Clients see no difference in API behavior
- **Minimal Overhead**: < 5ms added latency per request
- **Fail-Safe**: If proxy fails, direct Ollama access still works

### 2. Scalable Architecture
- **Service Separation**: Each component can scale independently
- **Stateless Design**: Services can be replicated without coordination
- **Queue Management**: Handles traffic spikes gracefully

### 3. Real-Time Insights
- **WebSocket Updates**: Sub-second dashboard updates
- **Streaming Metrics**: Continuous data flow
- **Live Analysis**: AI-powered insights in real-time

### 4. Comprehensive Monitoring
- **Request-Level**: Individual request tracking
- **System-Level**: Resource utilization monitoring
- **Service-Level**: Health and availability checks

## 🔗 Service Interactions

### Proxy ↔ Ollama
- **Protocol**: HTTP/JSON
- **Connection**: Direct TCP connection
- **Timeout**: Configurable (default: 30s)
- **Retry Logic**: Automatic retry with backoff

### Dashboard ↔ Prometheus
- **Protocol**: HTTP/REST
- **Queries**: PromQL for metrics aggregation
- **Refresh**: Configurable intervals (default: 5s)
- **Caching**: Client-side caching for performance

### Health ↔ All Services
- **Protocol**: HTTP health checks
- **Frequency**: Configurable (default: 30s)
- **Timeouts**: Service-specific timeouts
- **Alerting**: Multiple notification channels

## 📈 Performance Characteristics

### Latency
- **Proxy Overhead**: < 5ms average
- **Dashboard Updates**: < 1s from event
- **Health Checks**: < 10s detection time

### Throughput
- **Request Handling**: 100+ req/s sustained
- **Metrics Collection**: 1000+ metrics/s
- **Data Retention**: 30 days default

### Resource Usage
- **Memory**: ~100MB per service
- **CPU**: < 5% under normal load
- **Storage**: ~1GB/month metrics data

## 🔒 Security Considerations

### Network Security
- **Local Deployment**: All services run locally by default
- **No External Dependencies**: Self-contained system
- **Configurable Binding**: Services can bind to specific interfaces

### Data Privacy
- **No External Transmission**: All data stays local
- **Request Logging**: Optional and configurable
- **Metrics Anonymization**: No sensitive data in metrics

## 🎛️ Configuration Management

### Hierarchical Configuration
```
1. Default values (in code)
2. Service configuration files (config/services/*.yml)
3. Environment variables
4. Command-line arguments
```

### Service Discovery
- **Static Configuration**: Fixed ports and hosts
- **Health Check Integration**: Automatic service detection
- **Prometheus Discovery**: Service registration ready

## 🔄 Deployment Patterns

### Development
- **Single Machine**: All services on localhost
- **Docker Compose**: Container-based development
- **Hot Reload**: Development-time file watching

### Production
- **Containerized**: Docker/Podman deployment
- **Load Balancing**: Multiple proxy instances
- **High Availability**: Service redundancy

## 🚀 Extensibility Points

### Adding New Services
1. Create service in `services/` directory
2. Add Prometheus scrape job
3. Update dashboard queries
4. Add health check endpoint

### Custom Metrics
1. Define in shared metrics package
2. Implement in relevant service
3. Add to Prometheus config
4. Create dashboard visualization

### Integration Points
- **Webhook Support**: For external notifications
- **API Extensions**: RESTful endpoints for integration
- **Plugin Architecture**: Planned for future releases

---

For detailed component information, see [Components Documentation](components.md).
For data flow details, see [Data Flow Documentation](data_flow.md).