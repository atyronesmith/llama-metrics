# 🧹 Llama Metrics - Project Cleanup and Restructuring Plan

## 📋 Executive Summary

This document outlines a comprehensive plan to clean up and restructure the Llama Metrics project, addressing issues with overlapping components, inconsistent organization, and outdated documentation.

## 🎯 Current Issues Identified

### 1. Multiple Overlapping Components
- Three separate Go services (proxy, dashboard, health) with duplicated code
- Missing Python monitoring components mentioned in docs but not in root
- Conflicting documentation about which components to use

### 2. Inconsistent Organization
- Mix of Go and Python code without clear separation
- Scripts scattered between root and `/scripts` directory
- Multiple requirement files with unclear purpose
- Documentation spread across multiple locations

### 3. Configuration Confusion
- Multiple Prometheus config files
- Unclear which components are actively used
- Conflicting port assignments mentioned in docs

### 4. Documentation Issues
- PROJECT_STRUCTURE.md references files that don't exist
- Multiple README files with overlapping information
- Claude conversation logs mixed with documentation

## 📁 Proposed Clean Structure

```
llama-metrics/
├── services/                    # All Go services
│   ├── proxy/                  # Main monitoring proxy
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── pkg/
│   │   ├── Makefile
│   │   └── README.md
│   ├── dashboard/              # Web dashboard
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── pkg/
│   │   ├── web/
│   │   ├── Makefile
│   │   └── README.md
│   ├── health/                 # Health checker
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── pkg/
│   │   ├── Makefile
│   │   └── README.md
│   └── shared/                 # Shared Go packages
│       ├── config/             # Unified configuration
│       ├── metrics/            # Common metrics definitions
│       └── models/             # Shared data models
│
├── scripts/                    # All executable scripts
│   ├── traffic/               # Traffic generation
│   │   ├── generator.py       # Main traffic generator
│   │   ├── high_performance.py # Performance tester
│   │   └── scenarios.sh       # Test scenarios
│   ├── monitoring/            # Monitoring scripts
│   │   ├── start_stack.sh     # Start all services
│   │   ├── health_check.sh    # System health checks
│   │   └── mac_metrics.py     # Mac system metrics
│   └── deployment/            # Deployment scripts
│       ├── install.sh         # Installation script
│       ├── docker_build.sh    # Container builds
│       └── prometheus_setup.sh # Prometheus setup
│
├── config/                     # All configuration files
│   ├── prometheus/
│   │   └── prometheus.yml     # Single Prometheus config
│   ├── services/
│   │   ├── proxy.yml         # Proxy service config
│   │   ├── dashboard.yml     # Dashboard config
│   │   └── health.yml        # Health service config
│   └── alerts/
│       └── rules.yml         # Alert rules
│
├── docs/                      # All documentation
│   ├── README.md             # Main project overview
│   ├── architecture/
│   │   ├── overview.md       # System architecture
│   │   ├── components.md     # Component details
│   │   └── data_flow.md      # Data flow diagrams
│   ├── setup/
│   │   ├── quick_start.md    # 5-minute setup
│   │   ├── installation.md   # Detailed install
│   │   └── configuration.md  # Config guide
│   ├── api/
│   │   ├── proxy.md         # Proxy API docs
│   │   ├── dashboard.md     # Dashboard API
│   │   └── metrics.md       # Metrics reference
│   └── development/
│       ├── contributing.md   # Contribution guide
│       ├── testing.md       # Testing guide
│       └── troubleshooting.md # Common issues
│
├── test/                      # All tests
│   ├── integration/          # Integration tests
│   ├── unit/                # Unit tests
│   ├── load/                # Load testing
│   └── data/                # Test data
│       └── questions/       # Question categories
│
├── build/                    # Build artifacts
│   └── .gitkeep
│
├── README.md                 # Main project README
├── Makefile                 # Root orchestration Makefile
├── go.work                  # Go workspace file
├── .gitignore
├── VERSION
└── Dockerfile
```

## 🚀 Implementation Phases

### Phase 1: Consolidate Services (Week 1)

#### 1.1 Create Shared Packages
**Objective**: Extract common code to avoid duplication

**Actions**:
- Create `services/shared/config` for unified configuration management
- Create `services/shared/metrics` for common Prometheus metrics
- Create `services/shared/models` for shared data structures
- Update import paths in all services
- Add go.work file for workspace management

**Shared Package Structure**:
```go
// services/shared/config/config.go
package config

type BaseConfig struct {
    Port        int
    MetricsPort int
    LogLevel    string
}

// services/shared/metrics/metrics.go
package metrics

var (
    RequestCounter *prometheus.CounterVec
    LatencyHistogram *prometheus.HistogramVec
)
```

#### 1.2 Standardize Service Structure
**Objective**: Consistent structure across all services

**Actions**:
- Standardize Makefile targets (build, run, test, clean)
- Use consistent port numbering scheme
- Implement graceful shutdown in all services
- Add health check endpoints

### Phase 2: Organize Scripts (Week 1)

#### 2.1 Categorize Scripts
**Objective**: Logical grouping of scripts by function

**Migration Map**:
```
Current Location → New Location
scripts/traffic_generator.py → scripts/traffic/generator.py
scripts/high_performance_load_tester.py → scripts/traffic/high_performance.py
scripts/load_test_scenarios.sh → scripts/traffic/scenarios.sh
scripts/run_traffic_generator.sh → scripts/traffic/run.sh
generate_traffic.sh → scripts/traffic/simple.sh
direct_traffic_generator.sh → scripts/traffic/direct.sh
scripts/start_monitoring_ollama.sh → scripts/monitoring/start_stack.sh
scripts/test_priority_queue.py → test/unit/test_priority_queue.py
install.sh → scripts/deployment/install.sh
```

#### 2.2 Remove Python Component References
**Objective**: Clean up references to non-existent Python monitoring components

**Files to Update**:
- Remove from PROJECT_STRUCTURE.md: ollama_monitoring_proxy_fixed.py, enhanced_metrics_server.py, app.py, dashboard.py
- Update MONITORING_RULES.md to reflect actual components
- Update Makefile to remove Python monitoring targets

### Phase 3: Configuration Management (Week 2)

#### 3.1 Consolidate Configurations
**Objective**: Single source of truth for each configuration type

**Actions**:
- Merge `docs/prometheus.yml` and `prometheus.yml` → `config/prometheus/prometheus.yml`
- Create service-specific YAML configs from embedded configurations
- Move `ollama_alerts.yml` → `config/alerts/rules.yml`
- Implement config validation in each service

**Configuration Schema**:
```yaml
# config/services/proxy.yml
service:
  name: ollama-proxy
  version: ${VERSION}

server:
  proxy_port: 11435
  metrics_port: 8001

ollama:
  url: http://localhost:11434
  timeout: 30s

monitoring:
  metrics_path: /metrics
  health_path: /health
```

### Phase 4: Documentation Restructure (Week 2)

#### 4.1 Consolidate Documentation
**Objective**: Single, well-organized documentation structure

**Migration Plan**:
1. Main README.md → Keep as project overview (simplified)
2. README_traffic_generator.md → docs/setup/traffic_generation.md
3. SETUP.md → docs/setup/installation.md
4. API_DOCUMENTATION.md → Split into docs/api/{proxy,dashboard,metrics}.md
5. MONITORING_RULES.md → docs/development/guidelines.md
6. PERFORMANCE_TUNING.md → docs/setup/performance.md
7. PROGRESS_GO_DASHBOARD.md → Archive or integrate into relevant docs
8. claude/* → Archive to docs/archive/claude/

#### 4.2 Create New Documentation
**Objective**: Fill gaps in current documentation

**New Documents**:
- `docs/architecture/overview.md` - System architecture with diagrams
- `docs/setup/quick_start.md` - 5-minute setup guide
- `docs/development/contributing.md` - Contribution guidelines
- `docs/api/metrics.md` - Complete metrics reference

### Phase 5: Testing Framework (Week 3)

#### 5.1 Organize Tests
**Objective**: Comprehensive, organized testing

**Structure**:
```
test/
├── integration/
│   ├── proxy_ollama_test.go      # Proxy-Ollama integration
│   ├── metrics_prometheus_test.go # Metrics-Prometheus integration
│   └── end_to_end_test.py        # Full stack test
├── unit/
│   ├── proxy/                    # Proxy unit tests
│   ├── dashboard/                # Dashboard unit tests
│   └── health/                   # Health unit tests
├── load/
│   ├── scenarios/               # Load test scenarios
│   ├── results/                # Test results (gitignored)
│   └── analysis/               # Result analysis scripts
└── data/
    ├── questions/              # Move from root questions/
    └── fixtures/              # Test fixtures
```

## 📝 Migration Steps

### Step 1: Preparation (Day 1)
```bash
# Create backup
git checkout -b cleanup-backup
git add -A && git commit -m "Backup before cleanup"

# Create cleanup branch
git checkout -b project-cleanup
```

### Step 2: Create New Structure (Day 1)
```bash
# Create all directories
mkdir -p services/{proxy,dashboard,health,shared/{config,metrics,models}}
mkdir -p scripts/{traffic,monitoring,deployment}
mkdir -p config/{prometheus,services,alerts}
mkdir -p docs/{architecture,setup,api,development,archive/claude}
mkdir -p test/{integration,unit/{proxy,dashboard,health},load/{scenarios,results,analysis},data/{questions,fixtures}}
mkdir -p build
touch build/.gitkeep
```

### Step 3: Move Services (Day 2)
```bash
# Move Go services
git mv proxy/* services/proxy/
git mv dashboard/* services/dashboard/
git mv health/* services/health/

# Remove empty directories
rmdir proxy dashboard health
```

### Step 4: Move Scripts (Day 2)
```bash
# Traffic scripts
git mv scripts/traffic_generator.py scripts/traffic/generator.py
git mv scripts/high_performance_load_tester.py scripts/traffic/high_performance.py
git mv scripts/load_test_scenarios.sh scripts/traffic/scenarios.sh
git mv scripts/run_traffic_generator.sh scripts/traffic/run.sh
git mv generate_traffic.sh scripts/traffic/simple.sh
git mv direct_traffic_generator.sh scripts/traffic/direct.sh

# Monitoring scripts
git mv scripts/start_monitoring_ollama.sh scripts/monitoring/start_stack.sh
git mv scripts/mac_metrics_helper.py scripts/monitoring/mac_metrics.py

# Deployment scripts
git mv install.sh scripts/deployment/install.sh

# Test scripts
git mv scripts/test_priority_queue.py test/unit/test_priority_queue.py
```

### Step 5: Move Configurations (Day 3)
```bash
# Prometheus configs
git mv prometheus.yml config/prometheus/
git rm docs/prometheus.yml  # Remove duplicate

# Alert rules
git mv ollama_alerts.yml config/alerts/rules.yml

# Service configs (extract from code)
# Create new YAML configs for each service
```

### Step 6: Reorganize Documentation (Day 3)
```bash
# Archive Claude conversations
git mv claude/* docs/archive/claude/
rmdir claude

# Move existing docs
git mv README_traffic_generator.md docs/setup/traffic_generation.md
git mv docs/SETUP.md docs/setup/installation.md
git mv docs/API_DOCUMENTATION.md docs/api/
git mv MONITORING_RULES.md docs/development/guidelines.md
git mv PERFORMANCE_TUNING.md docs/setup/performance.md

# Archive or remove
git mv PROGRESS_GO_DASHBOARD.md docs/archive/
git rm PROJECT_STRUCTURE.md  # Will be replaced by this document
```

### Step 7: Move Test Data (Day 4)
```bash
# Move questions
git mv questions/* test/data/questions/
rmdir questions

# Move any test data from data/
git mv data/* test/data/
rmdir data
```

### Step 8: Update Build Files (Day 4)
```bash
# Create Go workspace file
cat > go.work << 'EOF'
go 1.21

use (
    ./services/proxy
    ./services/dashboard
    ./services/health
)
EOF

# Update root Makefile for new structure
# Update service Makefiles for shared packages
```

### Step 9: Update Import Paths (Day 5)
```bash
# Update all Go import paths
find services -name "*.go" -type f -exec sed -i 's|github.com/user/llama-metrics/proxy|github.com/user/llama-metrics/services/proxy|g' {} +
# Repeat for dashboard and health
```

### Step 10: Clean Root Directory (Day 5)
```bash
# Remove old requirement files (consolidate if needed)
git rm requirements_*.txt
# Keep only requirements.txt with all dependencies

# Update .gitignore for new structure
# Update README.md to reflect new structure
```

## 🎯 Success Criteria

### Technical Criteria
- [ ] All services build and run successfully
- [ ] No duplicate code between services
- [ ] All tests pass
- [ ] Documentation is accurate and complete
- [ ] CI/CD pipelines work with new structure

### Organization Criteria
- [ ] Clear separation of concerns
- [ ] Intuitive directory structure
- [ ] No orphaned files
- [ ] Consistent naming conventions
- [ ] Easy to navigate for new developers

## 🚨 Rollback Plan

If issues arise during migration:

1. **Immediate Rollback**:
   ```bash
   git checkout main
   git branch -D project-cleanup
   ```

2. **Partial Rollback**:
   - Cherry-pick successful changes
   - Fix issues on cleanup branch
   - Re-attempt migration

3. **Incremental Approach**:
   - Implement one phase at a time
   - Test thoroughly between phases
   - Merge to main after each successful phase

## 📊 Maintenance Guidelines

### Post-Cleanup Rules

1. **Service Development**:
   - New services go in `services/` directory
   - Must use shared packages
   - Must include README and Makefile

2. **Script Management**:
   - Scripts must be categorized appropriately
   - Include usage documentation
   - Use consistent shebang and error handling

3. **Documentation**:
   - Keep docs in sync with code
   - Use relative links between docs
   - Archive old documentation instead of deleting

4. **Configuration**:
   - All configs in `config/` directory
   - Use YAML for consistency
   - Include schema validation

## 🔄 Next Steps

1. Review and approve this plan
2. Create project-cleanup branch
3. Execute migration steps
4. Test thoroughly
5. Update CI/CD configurations
6. Merge to main
7. Tag new version (e.g., v2.0.0)

---

*Document Version: 1.0*
*Created: [Current Date]*
*Status: Pending Approval*