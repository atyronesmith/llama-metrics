#!/bin/bash

# Start dedicated Ollama instance for monitoring/AI status generation
# This runs on port 11435 and is separate from the main Ollama instance (11434)

echo "🤖 Starting dedicated monitoring Ollama instance..."

# Set environment variables for the monitoring instance
export OLLAMA_HOST="127.0.0.1:11435"
export OLLAMA_KEEP_ALIVE="30m"  # Keep models loaded longer for monitoring
export OLLAMA_MAX_LOADED_MODELS="1"  # Only need one model for monitoring
export OLLAMA_NUM_PARALLEL="1"  # Minimal parallel requests
export OLLAMA_MAX_QUEUE="5"  # Small queue for monitoring requests

# Check if port 11435 is already in use
if lsof -ti:11435 > /dev/null 2>&1; then
    echo "⚠️  Port 11435 already in use. Stopping existing process..."
    lsof -ti:11435 | xargs kill -9 2>/dev/null || true
    sleep 2
fi

# Start Ollama in background
echo "🚀 Starting Ollama monitoring instance on port 11435..."
nohup ollama serve > /dev/null 2>&1 &
OLLAMA_PID=$!

# Wait for the service to start
echo "⏳ Waiting for Ollama monitoring instance to start..."
sleep 5

# Test if the service is responding
MAX_RETRIES=10
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
    if curl -s http://localhost:11435/api/tags > /dev/null 2>&1; then
        echo "✅ Ollama monitoring instance is running on port 11435"
        
        # Pull the model if not already available
        echo "📥 Ensuring phi3:mini model is available..."
        OLLAMA_HOST="127.0.0.1:11435" ollama pull phi3:mini
        
        echo "🎉 Monitoring Ollama setup complete!"
        echo "💡 PID: $OLLAMA_PID"
        echo "💡 Port: 11435"
        echo "💡 URL: http://localhost:11435"
        exit 0
    fi
    
    RETRY_COUNT=$((RETRY_COUNT + 1))
    echo "⏳ Waiting for service... (attempt $RETRY_COUNT/$MAX_RETRIES)"
    sleep 2
done

echo "❌ Failed to start Ollama monitoring instance"
exit 1