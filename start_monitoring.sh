#!/bin/bash

# Start Ollama Monitoring Stack
# This script starts all components for comprehensive Ollama monitoring

echo "🚀 Starting Ollama Monitoring Stack"
echo "=================================="

# Check if Ollama is running
if ! pgrep -x "ollama" > /dev/null; then
    echo "⚠️  Ollama is not running. Starting Ollama..."
    ollama serve &
    sleep 5
else
    echo "✅ Ollama is already running"
fi

# Check if model is available
echo "🔍 Checking for phi3:mini model..."
if ! ollama list | grep -q "phi3:mini"; then
    echo "📥 Pulling phi3:mini model..."
    ollama pull phi3:mini
else
    echo "✅ phi3:mini model is available"
fi

# Start the monitoring proxy
echo ""
echo "🔄 Starting Ollama Monitoring Proxy..."
if pgrep -f "ollama_monitoring_proxy.py" > /dev/null; then
    echo "⚠️  Monitoring proxy is already running"
else
    # Activate virtual environment if it exists
    if [ -d "venv" ]; then
        # shellcheck source=/dev/null
        source venv/bin/activate
        python ollama_monitoring_proxy_fixed.py &
    else
        python3 ollama_monitoring_proxy_fixed.py &
    fi
    PROXY_PID=$!
    echo "✅ Monitoring proxy started (PID: $PROXY_PID)"
    sleep 3
fi

# Start the enhanced metrics server
echo ""
echo "📊 Starting Enhanced Metrics Server..."
if pgrep -f "enhanced_metrics_server.py" > /dev/null; then
    echo "⚠️  Enhanced metrics server is already running"
else
    python enhanced_metrics_server.py &
    METRICS_PID=$!
    echo "✅ Enhanced metrics server started (PID: $METRICS_PID)"
    sleep 3
fi

# Start Prometheus
echo ""
echo "📈 Starting Prometheus..."
if pgrep -x "prometheus" > /dev/null; then
    echo "⚠️  Prometheus is already running"
else
    # Check if prometheus is installed
    if command -v prometheus &> /dev/null; then
        prometheus --config.file=prometheus_config.yml &
        PROM_PID=$!
        echo "✅ Prometheus started (PID: $PROM_PID)"
    else
        echo "❌ Prometheus not found. Please install it first:"
        echo "   brew install prometheus  # macOS"
        echo "   or download from https://prometheus.io/download/"
    fi
fi

# Optional: Start Grafana
echo ""
echo "📊 Grafana (optional)..."
if pgrep -x "grafana-server" > /dev/null; then
    echo "✅ Grafana is already running"
else
    if command -v grafana-server &> /dev/null; then
        echo "Would you like to start Grafana? (y/n)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            grafana-server &
            GRAFANA_PID=$!
            echo "✅ Grafana started (PID: $GRAFANA_PID)"
        fi
    else
        echo "ℹ️  Grafana not installed. Install with:"
        echo "   brew install grafana  # macOS"
    fi
fi

# Display status
echo ""
echo "🎉 Ollama Monitoring Stack Status:"
echo "=================================="
echo "✅ Ollama API: http://localhost:11434"
echo "✅ Monitoring Proxy: http://localhost:11435"
echo "✅ Proxy Metrics: http://localhost:8001/metrics"
echo "✅ Enhanced Metrics: http://localhost:8000/metrics"
echo "✅ Prometheus: http://localhost:9090"
[[ -n "$GRAFANA_PID" ]] && echo "✅ Grafana: http://localhost:3000"

echo ""
echo "📝 Usage:"
echo "  - Use http://localhost:11435 instead of http://localhost:11434 for monitored requests"
echo "  - View metrics at http://localhost:9090 (Prometheus)"
echo "  - Run 'python test_ollama_monitoring.py' to test the setup"

echo ""
echo "🛑 To stop all services, run:"
echo "  pkill -f ollama_monitoring_proxy.py"
echo "  pkill -f enhanced_metrics_server.py"
echo "  pkill prometheus"
[[ -n "$GRAFANA_PID" ]] && echo "  pkill grafana-server"

# Save PIDs to file for easy shutdown
echo "PROXY_PID=$PROXY_PID" > monitoring_pids.txt
echo "METRICS_PID=$METRICS_PID" >> monitoring_pids.txt
echo "PROM_PID=$PROM_PID" >> monitoring_pids.txt
[[ -n "$GRAFANA_PID" ]] && echo "GRAFANA_PID=$GRAFANA_PID" >> monitoring_pids.txt

echo ""
echo "✨ Monitoring stack is ready!"