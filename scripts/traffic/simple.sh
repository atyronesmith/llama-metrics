#!/bin/bash
# Quick script to generate traffic with default settings

echo "🚦 Starting Ollama traffic generation..."
echo "This will send continuous requests through the monitoring proxy"
echo "Press Ctrl+C to stop"
echo ""

# Run with virtual environment Python
./venv/bin/python scripts/traffic/generator.py \
    --model phi3:mini \
    --url http://localhost:11435 \
    --delay 2.0