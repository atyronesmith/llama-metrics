#!/bin/bash

# Start LiteLLM proxy with priority queues for Ollama monitoring
echo "🚀 Starting LiteLLM proxy with priority queues..."

# Set configuration path
CONFIG_FILE="${PWD}/litellm_config.yaml"
if [[ ! -f "$CONFIG_FILE" ]]; then
    echo "❌ Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Check if port 8000 is already in use
if lsof -ti:8000 > /dev/null 2>&1; then
    echo "⚠️  Port 8000 already in use. Stopping existing process..."
    lsof -ti:8000 | xargs kill -9 2>/dev/null || true
    sleep 2
fi

# Activate virtual environment
if [[ -f "venv/bin/activate" ]]; then
    source venv/bin/activate
    echo "✅ Virtual environment activated"
else
    echo "❌ Virtual environment not found"
    exit 1
fi

# Set environment variables
export LITELLM_LOG="INFO"
export OLLAMA_BASE_URL="http://localhost:11434"
export LITELLM_DROP_PARAMS="true"

# Start LiteLLM proxy
echo "🔧 Starting LiteLLM proxy on port 8000..."
echo "📝 Config file: $CONFIG_FILE"
echo "🎯 Ollama backend: $OLLAMA_BASE_URL"

nohup litellm --config "$CONFIG_FILE" --port 8000 --host 0.0.0.0 > litellm.log 2>&1 &
LITELLM_PID=$!

# Wait for the service to start
echo "⏳ Waiting for LiteLLM proxy to start..."
sleep 5

# Test if the service is responding
MAX_RETRIES=10
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:8000/health > /dev/null 2>&1; then
        echo "✅ LiteLLM proxy is running on port 8000"
        echo "🎉 LiteLLM setup complete!"
        echo "💡 PID: $LITELLM_PID"
        echo "💡 Port: 8000"
        echo "💡 Config: $CONFIG_FILE"
        echo "💡 Logs: litellm.log"
        echo ""
        echo "📋 API Endpoints:"
        echo "  🔗 Health: http://localhost:8000/health"
        echo "  🔗 Models: http://localhost:8000/v1/models"
        echo "  🔗 Chat: http://localhost:8000/v1/chat/completions"
        echo "  🔗 Metrics: http://localhost:8000/metrics"
        echo ""
        echo "🔑 API Key: sk-1234567890abcdef"
        exit 0
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "⏳ Waiting for service... (attempt $RETRY_COUNT/$MAX_RETRIES)"
    sleep 2
done

echo "❌ Failed to start LiteLLM proxy"
echo "📋 Check logs with: tail -f litellm.log"
exit 1