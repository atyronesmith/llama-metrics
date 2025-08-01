#!/bin/bash

# Prometheus Container Runner for Ollama Monitoring Stack
# Starts Prometheus in a container with proper configuration

set -euo pipefail

# Configuration
PROMETHEUS_VERSION="latest"
PROMETHEUS_PORT="9090"
CONFIG_FILE="$(pwd)/config/prometheus/prometheus.yml"
DATA_DIR="$(pwd)/prometheus-data"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Starting Prometheus container...${NC}"

# Ensure data directory exists
mkdir -p "$DATA_DIR"

# Check if config file exists
if [ ! -f "$CONFIG_FILE" ]; then
    echo -e "${RED}❌ Prometheus config file not found: $CONFIG_FILE${NC}"
    exit 1
fi

# Determine container runtime (prefer podman, fallback to docker)
if command -v podman >/dev/null 2>&1; then
    RUNTIME="podman"
elif command -v docker >/dev/null 2>&1; then
    RUNTIME="docker"
else
    echo -e "${RED}❌ Neither podman nor docker found. Please install one of them.${NC}"
    echo -e "${YELLOW}Install podman: brew install podman${NC}"
    echo -e "${YELLOW}Install docker: brew install --cask docker${NC}"
    exit 1
fi

echo -e "${BLUE}Using container runtime: $RUNTIME${NC}"

# Detect host IP for container communication
if [[ "$RUNTIME" == "podman" ]]; then
    # For podman on macOS, try to get the host IP
    HOST_IP=$(ifconfig | grep 'inet 192.168' | head -1 | awk '{print $2}' || echo "192.168.65.1")
    HOST_OPTION="--add-host host.containers.internal:$HOST_IP"
elif [[ "$RUNTIME" == "docker" ]]; then
    # For Docker Desktop on macOS
    HOST_OPTION="--add-host host.containers.internal:host-gateway"
else
    HOST_OPTION=""
fi

echo -e "${BLUE}Host networking: $HOST_OPTION${NC}"

# Stop existing Prometheus container if running
echo -e "${YELLOW}Stopping any existing Prometheus container...${NC}"
$RUNTIME stop prometheus 2>/dev/null || true
$RUNTIME rm prometheus 2>/dev/null || true

# Start Prometheus container
echo -e "${BLUE}Starting Prometheus container with:${NC}"
echo -e "  Config: $CONFIG_FILE"
echo -e "  Data:   $DATA_DIR"
echo -e "  Port:   $PROMETHEUS_PORT"

$RUNTIME run -d \
    --name prometheus \
    --publish "$PROMETHEUS_PORT:9090" \
    --volume "$CONFIG_FILE:/etc/prometheus/prometheus.yml:ro" \
    --volume "$DATA_DIR:/prometheus:Z" \
    $HOST_OPTION \
    prom/prometheus:$PROMETHEUS_VERSION \
    --config.file=/etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/prometheus \
    --web.console.libraries=/etc/prometheus/console_libraries \
    --web.console.templates=/etc/prometheus/consoles \
    --web.enable-lifecycle \
    --web.enable-admin-api

# Wait for Prometheus to start
echo -e "${YELLOW}Waiting for Prometheus to start...${NC}"
sleep 3

# Check if Prometheus is running
if $RUNTIME ps | grep -q prometheus; then
    echo -e "${GREEN}✅ Prometheus started successfully${NC}"
    echo -e "${BLUE}Prometheus UI:${NC} http://localhost:$PROMETHEUS_PORT"
    echo -e "${BLUE}Targets:${NC} http://localhost:$PROMETHEUS_PORT/targets"
    echo -e "${BLUE}Metrics:${NC} http://localhost:$PROMETHEUS_PORT/api/v1/label/__name__/values"
else
    echo -e "${RED}❌ Failed to start Prometheus${NC}"
    echo -e "${YELLOW}Container logs:${NC}"
    $RUNTIME logs prometheus
    exit 1
fi

echo -e "${GREEN}🚀 Prometheus is ready for monitoring!${NC}"