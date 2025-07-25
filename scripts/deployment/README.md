# Deployment Scripts

This directory contains scripts for installation, setup, and deployment automation for the Ollama monitoring system.

## Scripts Overview

### 🚀 Quick Start
```bash
# One-command installation
./scripts/deployment/install.sh

# With custom options
./scripts/deployment/install.sh --dev --verbose
```

## Script Descriptions

### `install.sh` - Complete System Installer
**Purpose**: Automated installation and setup of the entire Ollama monitoring system.

**Features**:
- Dependency management and installation
- Python virtual environment setup
- Ollama installation and configuration
- Service configuration
- Container runtime setup (Docker/Podman)
- Prometheus configuration
- Initial system verification

**Usage**:
```bash
# Basic installation
./scripts/deployment/install.sh

# Development setup
./scripts/deployment/install.sh --dev

# Verbose output
./scripts/deployment/install.sh --verbose

# Skip prompts (CI/CD)
./scripts/deployment/install.sh --yes

# Custom installation path
./scripts/deployment/install.sh --prefix /opt/llama-metrics
```

**Installation Steps**:
1. **System Dependencies**: Python, pip, curl, git
2. **Ollama Installation**: Download and install Ollama
3. **Python Environment**: Virtual environment and packages
4. **Container Runtime**: Docker or Podman setup
5. **Configuration**: Service configs and environment
6. **Verification**: System health check
7. **Model Download**: Default model preparation

## Installation Options

### Command Line Arguments
- `--dev`: Install development dependencies and tools
- `--verbose`: Enable detailed output and logging
- `--yes`: Skip interactive prompts (automation mode)
- `--prefix PATH`: Custom installation directory
- `--no-ollama`: Skip Ollama installation
- `--no-containers`: Skip container runtime setup
- `--no-models`: Skip model download

### Environment Variables
```bash
# Installation configuration
export INSTALL_PREFIX="/opt/llama-metrics"
export PYTHON_VERSION="3.9"
export OLLAMA_VERSION="latest"

# Skip components
export SKIP_OLLAMA="false"
export SKIP_CONTAINERS="false"
export SKIP_MODELS="false"

# Development mode
export DEV_MODE="false"
```

## System Requirements

### Supported Platforms
- **macOS**: Intel and Apple Silicon
- **Linux**: Ubuntu 18.04+, CentOS 7+, RHEL 8+
- **Windows**: WSL2 with Ubuntu

### Hardware Requirements
- **CPU**: 4+ cores recommended
- **RAM**: 8GB minimum, 16GB recommended
- **Storage**: 10GB available space
- **GPU**: Optional, enhances Ollama performance

### Software Dependencies
- **Python**: 3.8+ (automatically installed if missing)
- **Git**: For repository operations
- **Curl**: For downloads
- **Container Runtime**: Docker or Podman

## Installation Workflows

### Fresh Installation
```bash
# 1. Clone repository
git clone https://github.com/atyronesmith/llama-metrics.git
cd llama-metrics

# 2. Run installer
./scripts/deployment/install.sh

# 3. Verify installation
make verify

# 4. Start services
make start
```

### Development Setup
```bash
# Development installation with extra tools
./scripts/deployment/install.sh --dev

# This installs:
# - Development dependencies
# - Code formatting tools
# - Testing frameworks
# - Debugging utilities
```

### CI/CD Pipeline
```bash
# Automated installation for CI/CD
./scripts/deployment/install.sh --yes --verbose

# For containers
./scripts/deployment/install.sh --yes --no-containers
```

### Offline Installation
```bash
# Download dependencies first
./scripts/deployment/install.sh --download-only

# Install offline
./scripts/deployment/install.sh --offline
```

## Post-Installation Tasks

### Service Verification
```bash
# Check all services
make verify

# Individual service checks
ollama list                    # Ollama
curl http://localhost:8001/health  # Proxy
curl http://localhost:3001/health  # Dashboard
```

### Initial Configuration
```bash
# Download default model
ollama pull phi3:mini

# Start monitoring stack
./scripts/monitoring/start_stack.sh

# Generate test traffic
./scripts/traffic/run.sh --quick
```

### Environment Setup
```bash
# Add to shell profile
echo 'export PATH="$PWD/scripts:$PATH"' >> ~/.bashrc
echo 'alias llama-start="make start"' >> ~/.bashrc
echo 'alias llama-stop="make stop"' >> ~/.bashrc
```

## Customization

### Custom Configuration
Create `config/local.yml` for local overrides:
```yaml
services:
  proxy:
    port: 11435
    metrics_port: 8001
  dashboard:
    port: 3001
  ollama:
    url: "http://localhost:11434"
```

### Custom Models
```bash
# Install specific models
ollama pull llama2:7b
ollama pull codellama:13b

# Update configuration
./scripts/deployment/configure_models.sh
```

### Custom Dashboards
```bash
# Copy dashboard templates
cp -r templates/dashboards config/

# Customize and restart
make restart-dashboard
```

## Troubleshooting

### Common Installation Issues

**Permission Denied**
```bash
# Fix script permissions
chmod +x scripts/deployment/install.sh

# Or run with explicit shell
bash scripts/deployment/install.sh
```

**Python Version Issues**
```bash
# Check Python version
python --version

# Install specific version (macOS)
brew install python@3.9

# Install specific version (Ubuntu)
sudo apt install python3.9
```

**Ollama Installation Fails**
```bash
# Manual Ollama installation
curl -fsSL https://ollama.ai/install.sh | sh

# Or skip Ollama during install
./scripts/deployment/install.sh --no-ollama
```

**Container Runtime Issues**
```bash
# Install Docker (Ubuntu)
sudo apt update && sudo apt install docker.io

# Install Podman (macOS)
brew install podman

# Or skip containers
./scripts/deployment/install.sh --no-containers
```

**Virtual Environment Issues**
```bash
# Clean and recreate venv
rm -rf venv
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Verification Commands
```bash
# Check installation status
./scripts/deployment/install.sh --check

# Verify services
make verify

# Run health checks
python scripts/monitoring/health_check.py

# Test traffic generation
./scripts/traffic/run.sh --quick
```

## Uninstallation

### Complete Removal
```bash
# Stop all services
make stop

# Remove virtual environment
rm -rf venv

# Remove downloaded models
ollama rm --all

# Remove Ollama (optional)
# See Ollama documentation for uninstall steps
```

### Partial Cleanup
```bash
# Clean build artifacts
make clean

# Reset configuration
git checkout -- config/

# Remove logs
rm -rf logs/
```

## Integration with CI/CD

### GitHub Actions Example
```yaml
- name: Install llama-metrics
  run: ./scripts/deployment/install.sh --yes --dev

- name: Verify installation
  run: make verify

- name: Run tests
  run: make test
```

### Docker Integration
```dockerfile
FROM ubuntu:22.04
COPY . /app
WORKDIR /app
RUN ./scripts/deployment/install.sh --yes --no-containers
```

## Best Practices

1. **Version Control**: Pin dependency versions for reproducible builds
2. **Testing**: Always run verification after installation
3. **Documentation**: Keep installation logs for troubleshooting
4. **Backup**: Backup configuration before updates
5. **Monitoring**: Set up alerts for installation failures

## Support

For installation issues:
1. Check this README for common solutions
2. Run `./scripts/deployment/install.sh --check` for diagnostics
3. Review installation logs in `logs/install.log`
4. Open an issue with installation details and error messages