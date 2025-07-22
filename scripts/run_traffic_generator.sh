#!/bin/bash

# Traffic Generator Runner Script
# This script makes it easy to run the Ollama traffic generator with monitoring

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
DEFAULT_MODEL="phi3:mini"
DEFAULT_URL="http://localhost:11435"  # Use the monitoring proxy!
DEFAULT_DELAY="2.0"
DEFAULT_MAX=""

echo -e "${BLUE}🚦 Ollama Traffic Generator${NC}"
echo "================================"

# Function to check if virtual environment exists
check_venv() {
    if [ ! -d "venv" ]; then
        echo -e "${RED}❌ Virtual environment not found!${NC}"
        echo "Please run: python3 -m venv venv && source venv/bin/activate && pip install -r requirements.txt"
        exit 1
    fi
}

# Function to check if Ollama is running
check_ollama() {
    if ! curl -s http://localhost:11434/api/tags > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Ollama doesn't seem to be running${NC}"
        echo "Would you like to start it? (y/n)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            echo "Starting Ollama..."
            ollama serve &
            sleep 5
        fi
    else
        echo -e "${GREEN}✅ Ollama is running${NC}"
    fi
}

# Function to check if monitoring proxy is running
check_proxy() {
    if ! curl -s http://localhost:11435/health > /dev/null 2>&1; then
        echo -e "${YELLOW}⚠️  Monitoring proxy is not running${NC}"
        echo "Would you like to start it? (y/n)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            echo "Starting monitoring proxy..."
            ../venv/bin/python ../ollama_monitoring_proxy.py > ../proxy.log 2>&1 &
            sleep 3
            echo -e "${GREEN}✅ Monitoring proxy started${NC}"
        else
            echo -e "${YELLOW}⚠️  Using direct Ollama connection (no monitoring)${NC}"
            DEFAULT_URL="http://localhost:11434"
        fi
    else
        echo -e "${GREEN}✅ Monitoring proxy is running${NC}"
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --quick)
            # Quick mode: 10 requests, no delay
            DEFAULT_MAX="10"
            DEFAULT_DELAY="0"
            shift
            ;;
        --demo)
            # Demo mode: 50 requests, 2s delay
            DEFAULT_MAX="50"
            DEFAULT_DELAY="2"
            shift
            ;;
        --stress)
            # Stress test: 1000 requests, 0.1s delay
            DEFAULT_MAX="1000"
            DEFAULT_DELAY="0.1"
            shift
            ;;
        --continuous)
            # Continuous mode: no limit, 1s delay
            DEFAULT_MAX=""
            DEFAULT_DELAY="1"
            shift
            ;;
        *)
            shift
            ;;
    esac
done

# Check prerequisites
check_venv
check_ollama
check_proxy

# Interactive mode if no preset was selected
if [ -z "$1" ]; then
    echo ""
    echo "Choose a traffic generation mode:"
    echo "  1) Quick test (10 requests, no delay)"
    echo "  2) Demo mode (50 requests, 2s delay)"
    echo "  3) Stress test (1000 requests, 0.1s delay)"
    echo "  4) Continuous (unlimited, 1s delay)"
    echo "  5) Custom settings"
    echo ""
    read -rp "Select mode (1-5): " mode

    case $mode in
        1)
            DEFAULT_MAX="10"
            DEFAULT_DELAY="0"
            ;;
        2)
            DEFAULT_MAX="50"
            DEFAULT_DELAY="2"
            ;;
        3)
            DEFAULT_MAX="1000"
            DEFAULT_DELAY="0.1"
            ;;
        4)
            DEFAULT_MAX=""
            DEFAULT_DELAY="1"
            ;;
        5)
            read -rp "Enter model name [$DEFAULT_MODEL]: " model
            DEFAULT_MODEL=${model:-$DEFAULT_MODEL}

            read -rp "Enter max requests (empty for unlimited): " max
            DEFAULT_MAX=$max

            read -rp "Enter delay between requests [$DEFAULT_DELAY]: " delay
            DEFAULT_DELAY=${delay:-$DEFAULT_DELAY}
            ;;
        *)
            echo "Invalid selection"
            exit 1
            ;;
    esac
fi

# Build the command
CMD="../venv/bin/python traffic_generator.py --model $DEFAULT_MODEL --url $DEFAULT_URL --delay $DEFAULT_DELAY"

if [ -n "$DEFAULT_MAX" ]; then
    CMD="$CMD --max $DEFAULT_MAX"
fi

# Display what we're about to run
echo ""
echo -e "${BLUE}🚀 Starting traffic generator with:${NC}"
echo "   Model: $DEFAULT_MODEL"
echo "   URL: $DEFAULT_URL"
echo "   Delay: ${DEFAULT_DELAY}s"
echo "   Max requests: ${DEFAULT_MAX:-unlimited}"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
echo ""

# Run the traffic generator
$CMD