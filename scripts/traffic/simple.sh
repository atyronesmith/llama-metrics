#!/bin/bash
# Quick script to generate traffic with default settings

echo "🚦 Starting Ollama traffic generation..."
echo "This will send continuous requests through the monitoring proxy"
echo "Press Ctrl+C to stop"
echo ""

# Get the script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Run with virtual environment Python
"$PROJECT_DIR/venv/bin/python" "$SCRIPT_DIR/generator.py" \
    --model phi3:mini \
    --url http://localhost:11435 \
    --delay 2.0