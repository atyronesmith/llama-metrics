#!/bin/bash
# Direct traffic generator - bypasses proxy and connects directly to Ollama

echo "🚦 Starting Direct Ollama Traffic Generation..."
echo "This will send requests directly to Ollama (no monitoring proxy)"
echo "Press Ctrl+C to stop"
echo ""

# Get the script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Run with virtual environment Python, connecting directly to Ollama
"$PROJECT_DIR/venv/bin/python" "$SCRIPT_DIR/generator.py" \
    --model phi3:mini \
    --url http://localhost:11434 \
    --delay 2.0