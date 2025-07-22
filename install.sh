#!/bin/bash

# Ollama Monitoring Stack - One-Command Installer for Mac M-Series
# Usage: curl -fsSL <raw-github-url>/install.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Ollama Monitoring Stack Installer${NC}"
echo -e "${BLUE}====================================${NC}"
echo ""

# Check if we're on macOS
if [[ "$(uname -s)" != "Darwin" ]]; then
    echo -e "${RED}❌ This installer is designed for macOS only${NC}"
    exit 1
fi

# Check for M-series Mac (optional warning)
if [[ "$(uname -m)" != "arm64" ]]; then
    echo -e "${YELLOW}⚠️  Warning: This installer is optimized for M-series Macs${NC}"
    echo -e "${YELLOW}   Your system: $(uname -m)${NC}"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check for required tools
echo -e "${BLUE}Checking system requirements...${NC}"
for cmd in git python3 curl; do
    if ! command -v $cmd >/dev/null 2>&1; then
        echo -e "${RED}❌ $cmd is required but not installed${NC}"
        if [[ "$cmd" == "git" ]]; then
            echo -e "${YELLOW}Please install Xcode Command Line Tools: xcode-select --install${NC}"
        elif [[ "$cmd" == "python3" ]]; then
            echo -e "${YELLOW}Please install Python 3 from: https://python.org${NC}"
        fi
        exit 1
    fi
done
echo -e "${GREEN}✅ System requirements met${NC}"

# Clone or update repository
REPO_URL="https://github.com/your-username/llamastack-prometheus.git"  # Update this
INSTALL_DIR="$HOME/llamastack-prometheus"

if [[ -d "$INSTALL_DIR" ]]; then
    echo -e "${YELLOW}Directory exists. Updating repository...${NC}"
    cd "$INSTALL_DIR"
    git pull
else
    echo -e "${BLUE}Cloning repository...${NC}"
    git clone "$REPO_URL" "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# Make setup executable
chmod +x load_test_scenarios.sh

echo ""
echo -e "${BLUE}Starting automated setup...${NC}"
echo ""

# Run setup
if make setup; then
    echo ""
    echo -e "${GREEN}🎉 Installation completed successfully!${NC}"
    echo ""
    echo -e "${BLUE}Starting services...${NC}"
    if make start; then
        echo ""
        echo -e "${GREEN}🚀 All services are now running!${NC}"
        echo ""
        echo -e "${BLUE}📊 Dashboard:${NC}        http://localhost:3001"
        echo -e "${BLUE}📈 Prometheus UI:${NC}    http://localhost:9090"
        echo -e "${BLUE}🔧 Metrics API:${NC}      http://localhost:8001/metrics"
        echo -e "${BLUE}🤖 Ollama API:${NC}       http://localhost:11434"
        echo ""
        echo -e "${YELLOW}💡 Next steps:${NC}"
        echo "   • Open http://localhost:3001 to view the dashboard"
        echo "   • Run 'cd $INSTALL_DIR && make traffic' to generate test traffic"
        echo "   • Run 'make help' to see all available commands"
        echo ""
        echo -e "${GREEN}✨ Enjoy monitoring your Ollama setup!${NC}"
    else
        echo -e "${RED}❌ Failed to start services${NC}"
        echo -e "${YELLOW}Try running: cd $INSTALL_DIR && make start${NC}"
        exit 1
    fi
else
    echo -e "${RED}❌ Setup failed${NC}"
    echo -e "${YELLOW}Check the error messages above and try running setup manually:${NC}"
    echo "   cd $INSTALL_DIR"
    echo "   make setup"
    exit 1
fi