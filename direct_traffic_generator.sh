#!/bin/bash
# Direct traffic generator - bypasses proxy and connects directly to Ollama

echo "🚦 Starting Direct Ollama Traffic Generation..."
echo "This will send requests directly to Ollama (no monitoring proxy)"
echo "Press Ctrl+C to stop"
echo ""

# Run with virtual environment Python, connecting directly to Ollama
./venv/bin/python traffic_generator.py \
    --model phi3:mini \
    --url http://localhost:11434 \
    --delay 2.0