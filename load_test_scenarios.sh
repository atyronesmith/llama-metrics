#!/bin/bash

# Load Test Scenarios for Ollama Performance Testing
# =================================================

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOAD_TESTER="$SCRIPT_DIR/high_performance_load_tester.py"
PYTHON_CMD="$SCRIPT_DIR/venv/bin/python"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Ollama High-Performance Load Test Scenarios${NC}"
echo "=================================================="

# Check if load tester exists
if [ ! -f "$LOAD_TESTER" ]; then
    echo -e "${RED}❌ Load tester not found: $LOAD_TESTER${NC}"
    exit 1
fi

# Check if Python environment exists
if [ ! -f "$PYTHON_CMD" ]; then
    echo -e "${RED}❌ Python virtual environment not found${NC}"
    echo "Please run 'make setup' first"
    exit 1
fi

show_menu() {
    echo ""
    echo -e "${YELLOW}Available Load Test Scenarios:${NC}"
    echo "1. 🔥 Queue Stress Test - High concurrency to test queue visualization"
    echo "2. 💥 Burst Load Test - Periodic bursts of requests"
    echo "3. 📈 Ramp Up Test - Gradually increasing load"
    echo "4. ⚡ Spike Test - Sudden traffic spikes"
    echo "5. 🌀 Chaos Test - Random unpredictable load"
    echo "6. 🏃 Quick Test - Short high-intensity test"
    echo "7. 🔧 Custom Test - Specify your own parameters"
    echo "8. 📊 Monitor Only - Just watch the dashboard"
    echo "9. 🛑 Exit"
    echo ""
}

queue_stress_test() {
    echo -e "${GREEN}🔥 Starting Queue Stress Test${NC}"
    echo "This test will create high concurrency to stress the request queue"
    echo "Watch the dashboard at http://localhost:3001 to see queue metrics!"
    echo ""
    
    $PYTHON_CMD "$LOAD_TESTER" \
        --pattern constant \
        --rps 5.0 \
        --concurrent 5 \
        --requests 100 \
        --prompts short medium
}

burst_load_test() {
    echo -e "${GREEN}💥 Starting Burst Load Test${NC}"
    echo "This test sends bursts of 50 requests every 10 seconds"
    echo ""
    
    $PYTHON_CMD "$LOAD_TESTER" \
        --pattern burst \
        --rps 20.0 \
        --concurrent 5 \
        --requests 400 \
        --burst-size 50 \
        --burst-interval 10.0 \
        --prompts short medium long
}

ramp_up_test() {
    echo -e "${GREEN}📈 Starting Ramp Up Test${NC}"
    echo "This test gradually increases load from 0 to 30 RPS over 5 minutes"
    echo ""
    
    $PYTHON_CMD "$LOAD_TESTER" \
        --pattern ramp \
        --rps 30.0 \
        --concurrent 5 \
        --duration 300 \
        --prompts short medium
}

spike_test() {
    echo -e "${GREEN}⚡ Starting Spike Test${NC}"
    echo "This test alternates between normal load and sudden spikes"
    echo ""
    
    $PYTHON_CMD "$LOAD_TESTER" \
        --pattern spike \
        --rps 15.0 \
        --concurrent 5 \
        --requests 600 \
        --burst-size 40 \
        --prompts short medium long
}

chaos_test() {
    echo -e "${GREEN}🌀 Starting Chaos Test${NC}"
    echo "This test sends completely random load patterns"
    echo ""
    
    $PYTHON_CMD "$LOAD_TESTER" \
        --pattern chaos \
        --rps 20.0 \
        --concurrent 5 \
        --requests 500 \
        --burst-size 30 \
        --prompts short medium long
}

quick_test() {
    echo -e "${GREEN}🏃 Starting Quick Test${NC}"
    echo "Short 2-minute high-intensity test for immediate results"
    echo ""
    
    $PYTHON_CMD "$LOAD_TESTER" \
        --pattern constant \
        --rps 40.0 \
        --concurrent 5 \
        --duration 120 \
        --prompts short
}

custom_test() {
    echo -e "${GREEN}🔧 Custom Load Test Configuration${NC}"
    echo ""
    
    read -p "Requests per second (default: 20): " rps
    rps=${rps:-20}
    
    read -p "Max concurrent requests (default: 5): " concurrent
    concurrent=${concurrent:-5}
    
    read -p "Total requests (default: 300): " requests
    requests=${requests:-300}
    
    echo "Load patterns: constant, burst, ramp, spike, chaos"
    read -p "Load pattern (default: constant): " pattern
    pattern=${pattern:-constant}
    
    echo "Prompt types: short, medium, long"
    read -p "Prompt types (default: short medium): " prompts
    prompts=${prompts:-"short medium"}
    
    echo ""
    echo -e "${YELLOW}Starting custom test with:${NC}"
    echo "  RPS: $rps"
    echo "  Concurrent: $concurrent"
    echo "  Requests: $requests"
    echo "  Pattern: $pattern"
    echo "  Prompts: $prompts"
    echo ""
    
    $PYTHON_CMD "$LOAD_TESTER" \
        --pattern "$pattern" \
        --rps "$rps" \
        --concurrent "$concurrent" \
        --requests "$requests" \
        --prompts $prompts
}

monitor_only() {
    echo -e "${GREEN}📊 Monitor Mode${NC}"
    echo ""
    echo "Dashboard: http://localhost:3001"
    echo "Prometheus: http://localhost:9090"
    echo "Metrics: http://localhost:8001/metrics"
    echo ""
    echo -e "${YELLOW}Press Ctrl+C to return to menu${NC}"
    
    # Just wait and let user monitor
    trap 'echo -e "\n${GREEN}Returning to menu...${NC}"' INT
    sleep infinity
}

# Main menu loop
while true; do
    show_menu
    read -p "Select scenario (1-9): " choice
    
    case $choice in
        1)
            queue_stress_test
            ;;
        2)
            burst_load_test
            ;;
        3)
            ramp_up_test
            ;;
        4)
            spike_test
            ;;
        5)
            chaos_test
            ;;
        6)
            quick_test
            ;;
        7)
            custom_test
            ;;
        8)
            monitor_only
            ;;
        9)
            echo -e "${GREEN}👋 Goodbye!${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}❌ Invalid selection. Please choose 1-9.${NC}"
            ;;
    esac
    
    echo ""
    echo -e "${BLUE}Test completed. Press Enter to continue...${NC}"
    read
done