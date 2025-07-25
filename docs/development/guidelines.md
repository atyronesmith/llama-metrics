# Ollama Monitoring Stack - Important Rules and Requirements

## 🚨 CRITICAL: Python Virtual Environment Requirement

**ALL Python scripts MUST be run within the virtual environment**

### Why?
- Dependencies like `aiohttp`, `prometheus-client`, and others are installed in the virtual environment
- Running outside the venv will result in `ModuleNotFoundError` errors

### How to Run Python Scripts

#### Option 1: Activate the virtual environment first
```bash
source venv/bin/activate
python script_name.py
```

#### Option 2: Use the virtual environment's Python directly
```bash
./venv/bin/python script_name.py
```

#### Option 3: For background processes
```bash
# Using nohup
source venv/bin/activate && nohup python script_name.py > logfile.log 2>&1 &

# Or directly
nohup ./venv/bin/python script_name.py > logfile.log 2>&1 &
```

## 📋 Monitoring Stack Components

1. **ollama_monitoring_proxy.py** - MUST run in venv
   - Port: 11435 (proxy), 8001 (metrics)
   - Command: `./venv/bin/python ollama_monitoring_proxy.py`

2. **enhanced_metrics_server.py** - MUST run in venv
   - Port: 8000
   - Command: `./venv/bin/python enhanced_metrics_server.py`

3. **app.py** - MUST run in venv
   - Port: 8000 (conflicts with enhanced_metrics_server.py)
   - Command: `./venv/bin/python app.py`

4. **test_ollama_monitoring.py** - MUST run in venv
   - Command: `./venv/bin/python test_ollama_monitoring.py`

5. **traffic_generator.py** - MUST run in venv
   - Generates continuous traffic for monitoring
   - Easy commands:
     - `./generate_traffic.sh` - Simple continuous traffic
     - `./run_traffic_generator.sh` - Interactive mode with options
     - `./run_traffic_generator.sh --quick` - Quick 10 request test
     - `./run_traffic_generator.sh --demo` - 50 requests demo
     - `./run_traffic_generator.sh --stress` - 1000 request stress test

## 🐚 Shell Script Quality Requirements

### All shell scripts MUST pass shellcheck with zero errors

**Why?**
- Shellcheck catches common bash pitfalls and errors
- Ensures scripts are portable and reliable
- Prevents security vulnerabilities

### How to validate shell scripts
```bash
# Check all project shell scripts
shellcheck *.sh

# Check specific script
shellcheck script_name.sh
```

### Common shellcheck fixes applied
- Use `read -r` instead of `read` to handle backslashes properly
- Add shellcheck directives for sourcing files: `# shellcheck source=/dev/null`
- Quote variables to prevent word splitting and globbing
- Use proper array syntax and conditionals

### Scripts validated with shellcheck
- ✅ `direct_traffic_generator.sh`
- ✅ `generate_traffic.sh`
- ✅ `run_prometheus.sh`
- ✅ `run_traffic_generator.sh`
- ✅ `start_monitoring.sh`

## 🚦 Traffic Generator Usage

### Quick Start
```bash
# Simplest way - continuous traffic with defaults
./generate_traffic.sh

# Interactive mode - choose from presets
./run_traffic_generator.sh

# Quick test mode - 10 requests
./run_traffic_generator.sh --quick

# Demo mode - 50 requests with 2s delay
./run_traffic_generator.sh --demo

# Stress test - 1000 requests with 0.1s delay
./run_traffic_generator.sh --stress

# Direct command with custom options
./venv/bin/python traffic_generator.py --model phi3:mini --url http://localhost:11435 --max 100 --delay 1.5
```

### Traffic Generator Features
- Loads questions from `questions/` directory (1000+ diverse questions)
- Sends requests through monitoring proxy (port 11435)
- Tracks success rate and latency
- Supports various modes: quick test, demo, stress test, continuous
- Automatically uses virtual environment

## 🔧 Setup Checklist

- [ ] Virtual environment exists: `venv/`
- [ ] Virtual environment activated: `source venv/bin/activate`
- [ ] Required packages installed: `pip install -r requirements_monitoring.txt`
- [ ] Ollama is running: `ollama serve`
- [ ] Model is available: `ollama pull phi3:mini`
- [ ] All shell scripts pass shellcheck: `shellcheck *.sh`

## 🚀 Correct Startup Sequence

```bash
# 1. Ensure Ollama is running
ollama serve &

# 2. Start monitoring proxy (in venv)
./venv/bin/python ollama_monitoring_proxy.py &

# 3. Start enhanced metrics server (in venv)
./venv/bin/python enhanced_metrics_server.py &

# 4. Start Prometheus
prometheus --config.file=prometheus_config.yml &

# 5. Optional: Start Grafana
grafana-server &

# 6. Generate traffic for monitoring
./generate_traffic.sh &
```

## ⚠️ Common Errors and Solutions

### Error: `ModuleNotFoundError: No module named 'aiohttp'`
**Solution**: You're not in the virtual environment. Run:
```bash
source venv/bin/activate
```

### Error: `zsh: command not found: python`
**Solution**: Use `python3` or activate the virtual environment

### Error: Port already in use
**Solution**: Check running processes:
```bash
lsof -i :8000  # or other port number
pkill -f script_name.py
```

### Error: Shellcheck warnings/errors
**Solution**: Fix the specific issues reported by shellcheck:
```bash
shellcheck script_name.sh
# Fix issues, then verify:
shellcheck script_name.sh
```

## 📝 Environment Variables

When running scripts, ensure the virtual environment is active:
```bash
# Wrong
python script.py

# Correct
source venv/bin/activate && python script.py
# or
./venv/bin/python script.py
```

## 🔍 Verification Commands

```bash
# Check if virtual environment is active
echo $VIRTUAL_ENV

# Check Python path
which python

# Should show: /path/to/llamastack-prometheus/venv/bin/python

# Verify all shell scripts pass shellcheck
shellcheck *.sh && echo "✅ All shell scripts pass shellcheck"
```

## 📌 Remember

**ALWAYS use the virtual environment when running ANY Python script in this project!**
**ALWAYS validate shell scripts with shellcheck before committing!**