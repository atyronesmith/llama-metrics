# 📁 Project Structure

```
llamastack-prometheus/
├── 🔧 Core Application Files
│   ├── CLAUDE.md                          # Claude Code context (MUST stay in root)
│   ├── Makefile                          # Main automation & build system
│   ├── install.sh                        # One-command installer script
│   ├── ollama_monitoring_proxy_fixed.py  # Main monitoring proxy
│   ├── dashboard.py                      # Real-time web dashboard
│   ├── app.py                           # Alternative metrics server
│   ├── enhanced_metrics_server.py       # Enhanced metrics server
│   └── test_ollama_monitoring.py        # Test suite
│
├── 📋 Requirements Files
│   ├── requirements.txt                  # Core dependencies
│   ├── requirements_monitoring.txt       # Monitoring dependencies
│   ├── requirements_traffic.txt         # Traffic generation dependencies
│   └── requirements_dashboard.txt       # Dashboard dependencies
│
├── 🚀 scripts/                          # All executable scripts
│   ├── load_test_scenarios.sh           # Interactive load testing scenarios
│   ├── high_performance_load_tester.py  # High-performance load testing
│   ├── traffic_generator.py             # Basic traffic generation
│   ├── run_traffic_generator.sh         # Traffic generator wrapper
│   └── run_prometheus.sh               # Prometheus container runner
│
├── 📖 docs/                             # Documentation & configs
│   ├── SETUP.md                        # User setup guide
│   ├── prometheus.yml                  # Prometheus config (original)
│   └── prometheus_config.yml           # Prometheus config (current)
│
├── 🤖 claude/                           # Claude-specific files
│   ├── conv1.md                        # Conversation log 1
│   ├── convo2.md                       # Conversation log 2
│   └── dashboard_enhancement_plan.md    # Dashboard enhancement plans
│
├── 📊 templates/                        # Web templates
│   └── dashboard.html                   # Dashboard template
│
├── 🗂️ questions/                        # Test question categories
│   ├── coding/                         # Programming questions
│   ├── creative/                       # Creative writing prompts
│   ├── analysis/                       # Analysis questions
│   └── [8 more categories...]          # Additional test categories
│
├── 📝 logs/                            # Log files & runtime data
│   ├── monitoring_pids.txt             # Process ID tracking
│   ├── dashboard.log                   # Dashboard logs
│   ├── proxy.log                       # Proxy logs
│   └── ollama.log                      # Ollama logs
│
├── 🔗 venv/                            # Python virtual environment
│   └── [Python environment files]
│
└── 🗃️ Other Files
    ├── .ansible/                       # Ansible configuration
    ├── .gitignore                      # Git ignore rules
    └── README.md                       # Main project documentation
```

## 🏗️ Architecture Overview

### Core Components
- **ollama_monitoring_proxy_fixed.py**: Main proxy server (port 11435 → 11434)
- **dashboard.py**: Real-time monitoring dashboard (port 3001)
- **scripts/**: All executable scripts organized by function
- **templates/**: Web UI templates for the dashboard

### Data Flow
```
[scripts/traffic_generator.py] → [proxy:11435] → [Ollama:11434]
                                      ↓
[Dashboard:3001] ← [Prometheus:9090] ← [Metrics:8001]
```

### Configuration Files
- **docs/prometheus_config.yml**: Prometheus scraping configuration
- **requirements_*.txt**: Dependency management by component
- **CLAUDE.md**: Project context for Claude Code (must remain in root)

## 🎯 Key Design Decisions

1. **CLAUDE.md in Root**: Required for Claude Code context initialization
2. **scripts/ Directory**: All executable scripts for clean organization
3. **docs/ Directory**: User documentation and configuration files
4. **claude/ Directory**: Development conversation logs and plans
5. **logs/ Directory**: Runtime logs and process tracking
6. **Modular Requirements**: Separate dependency files by component