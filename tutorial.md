# AI-Assisted Rapid Prototyping Tutorial: Building an Ollama Monitoring Stack

This tutorial demonstrates how to use AI coding assistants to rapidly prototype a comprehensive AI application monitoring system. By following these step-by-step prompts, you'll build a production-ready monitoring solution for Ollama AI models.

## 🎯 What You'll Build

A complete monitoring stack featuring:
- **Traffic Generation**: 1000+ curated questions across 10 categories for load testing
- **Go Monitoring Services**: Proxy, Dashboard, and Health checker with shared libraries
- **Python Components**: Configuration management and containerization
- **Real-time Dashboard**: WebSocket-based monitoring with AI-powered insights
- **Prometheus Integration**: Metrics collection and visualization
- **Testing Framework**: Comprehensive unit, integration, and e2e tests

## 🏗️ Tutorial Overview

This tutorial is broken into **12 phases**, each building upon the previous. Each phase includes:
- **AI Prompt**: The exact prompt to give your AI assistant
- **Explanation**: How the prompt implements the step
- **Validation**: How to verify the step worked
- **Key Concepts**: What you learn from each phase

## Prerequisites

- **Ollama** installed and running
- **Go 1.24.4+** for the services
- **Python 3.8+** for traffic generation
- **Git** for version control
- **AI Coding Assistant** (Claude Code, GitHub Copilot, etc.)

---

## Phase 1: Project Foundation and Python Traffic Generator

**🎯 Goal**: Create the initial project structure with a Python traffic generator and basic monitoring.

### AI Prompt 1.1: Initialize Project Structure

```
Create a new project called "llama-metrics" for monitoring Ollama AI models. Set up the initial structure with:

1. A Python traffic generator that sends questions to Ollama models
2. A collection of 1000+ questions across 10 categories (100 questions each):
   - General Knowledge, Science, Technology, History, Geography
   - Sports, Entertainment, Literature, Philosophy, Food
3. A simple metrics server that exposes Prometheus metrics
4. Requirements files for different use cases
5. Basic README and gitignore files
6. A Makefile for automation

The traffic generator should:
- Load questions from JSON files with proper schema
- Send requests to Ollama API at localhost:11434
- Support configurable models, delays, and limits
- Provide real-time statistics and logging
- Handle errors gracefully

Create this as a working Python application with proper structure.
```

**💡 Explanation**: This prompt establishes the foundation by creating a working traffic generator with realistic test data. The AI will create the directory structure, implement the traffic generation logic, and set up the question database that will be used throughout the project.

**✅ Validation**:
```bash
# After AI creates the project
cd llama-metrics
python traffic_generator.py --max 5 --delay 2
```

### AI Prompt 1.2: Add Prometheus Integration

```
Add Prometheus metrics collection to the project:

1. Create a simple_metrics_server.py that exposes metrics on port 8001
2. Track key metrics:
   - Request count by model and status
   - Request duration histograms
   - Active request gauges
   - Question categories used
3. Create Prometheus configuration file
4. Add a script to run Prometheus in a container
5. Update the Makefile with targets for starting/stopping services
6. Add comprehensive documentation

The metrics server should integrate with the traffic generator to provide real monitoring data.
```

**💡 Explanation**: This prompt adds observability to the basic traffic generator. The AI will implement Prometheus metrics collection, which is essential for monitoring AI applications in production.

**✅ Validation**:
```bash
make start
curl http://localhost:8001/metrics | grep ollama
```

---

## Phase 2: Go Services Foundation

**🎯 Goal**: Build the core Go services with proper architecture and shared libraries.

### AI Prompt 2.1: Create Go Monitoring Proxy

```
Create a Go-based monitoring proxy service that sits between clients and Ollama:

1. Create a services/proxy directory with proper Go module structure
2. Implement a proxy that forwards requests from port 11435 to Ollama on 11434
3. Collect comprehensive metrics:
   - Request/response latency
   - Token generation rates
   - System metrics (CPU, memory, GPU on macOS)
   - Queue depth and processing rates
4. Add request queuing to handle load spikes
5. Support OpenAI-compatible API endpoints
6. Include health check endpoints
7. Use proper Go project structure with cmd/, internal/, pkg/ directories
8. Add Makefile for building and running

The proxy should be production-ready with proper error handling, logging, and graceful shutdown.
```

**💡 Explanation**: This prompt creates the core monitoring component using Go for better performance. The AI will implement enterprise-grade features like request queuing, comprehensive metrics, and proper Go project structure.

**✅ Validation**:
```bash
cd services/proxy
make build && make run
curl http://localhost:11435/v1/models
```

### AI Prompt 2.2: Create Real-time Dashboard Service

```
Create a Go-based real-time dashboard service:

1. Create services/dashboard directory with Go module
2. Implement a web dashboard with:
   - Real-time metrics via WebSockets
   - System monitoring graphs (CPU, memory, GPU, power on macOS)
   - Request queue analytics
   - Token generation performance charts
   - Latency percentiles (P50, P75, P95, P99)
   - AI-powered status analysis
3. Use HTML templates with modern CSS and JavaScript
4. Integrate with Prometheus for data sources
5. Add proper configuration management
6. Include health endpoints and graceful shutdown

The dashboard should provide comprehensive insights into Ollama performance with a professional UI.
```

**💡 Explanation**: This prompt creates a real-time monitoring interface. The AI will build a complete web application with data visualization, which is crucial for operational monitoring of AI systems.

**✅ Validation**:
```bash
cd services/dashboard
make build && make run
# Open http://localhost:3001 in browser
```

### AI Prompt 2.3: Create Health Checker Service

```
Create a health checking service with AI-powered analysis:

1. Create services/health directory with Go module
2. Implement health checks for:
   - Ollama API availability
   - Proxy service health
   - System resource usage
   - Queue health and performance
3. Add AI-powered health analysis that uses Ollama to:
   - Analyze system metrics
   - Provide actionable insights
   - Detect performance issues
   - Generate health reports
4. Support multiple modes: CLI, server, and API
5. Include comprehensive checks: simple, readiness, liveness, and analyzed
6. Add proper configuration and error handling

The health checker should act as an intelligent monitoring companion.
```

**💡 Explanation**: This prompt adds intelligent health monitoring using AI to analyze system state. This demonstrates how AI can be used not just as the monitored system, but as part of the monitoring solution itself.

**✅ Validation**:
```bash
cd services/health
make build
./build/healthcheck -mode cli -check comprehensive
```

---

## Phase 3: Shared Libraries and Go Workspace

**🎯 Goal**: Create shared libraries and organize Go services with a workspace.

### AI Prompt 3.1: Create Shared Go Packages

```
Create shared Go packages for common functionality across services:

1. Create services/shared directory with:
   - config/: Common configuration management
   - metrics/: Shared Prometheus metrics
   - models/: Common data structures
2. Move common code from services into shared packages
3. Update all services to use shared packages
4. Create proper Go module with versioning
5. Add comprehensive tests for shared functionality
6. Implement configuration loading from YAML and environment variables
7. Add standardized logging and error handling

Ensure all services build and run with the new shared architecture.
```

**💡 Explanation**: This prompt refactors the codebase to use shared libraries, demonstrating how AI can help restructure code for better maintainability and consistency across services.

**✅ Validation**:
```bash
cd services/shared && go test ./...
cd ../proxy && make build
cd ../dashboard && make build
```

### AI Prompt 3.2: Setup Go Workspace

```
Convert the Go services to use a Go workspace for better development experience:

1. Create go.work file that includes all service modules
2. Update all import paths to work with the workspace
3. Ensure cross-service dependencies work correctly
4. Update all Makefiles to work with the workspace structure
5. Add workspace synchronization commands
6. Update documentation for the new workspace setup
7. Test that all services build and run correctly

The workspace should simplify development while maintaining proper module boundaries.
```

**💡 Explanation**: This prompt sets up a modern Go workspace, which allows multiple related modules to be developed together efficiently. The AI will handle the complex dependency management.

**✅ Validation**:
```bash
go work sync
go list -m all
make build  # Should build all services
```

---

## Phase 4: Advanced Traffic Generation

**🎯 Goal**: Create sophisticated load testing capabilities.

### AI Prompt 4.1: High-Performance Load Tester

```
Create a high-performance load testing system:

1. Create scripts/traffic/high_performance.py with:
   - Multiple load patterns: constant, burst, chaos, ramp-up
   - Configurable concurrency and request rates
   - Support for different prompt lengths (short, medium, long)
   - Real-time performance metrics
   - Detailed latency analysis
   - Queue monitoring integration
2. Add load testing scenarios script that provides interactive menus
3. Create wrapper scripts for common test patterns
4. Add comprehensive performance reporting
5. Integrate with the monitoring proxy for accurate metrics

The load tester should be capable of stress testing production systems safely.
```

**💡 Explanation**: This prompt creates advanced load testing capabilities needed for performance validation of AI systems. The AI will implement sophisticated patterns that reveal system behavior under various conditions.

**✅ Validation**:
```bash
python scripts/traffic/high_performance.py --pattern constant --rps 5 --duration 60
```

### AI Prompt 4.2: Load Testing Integration

```
Integrate load testing with the monitoring system:

1. Update the Makefile with load testing targets:
   - Quick tests for validation
   - Demo tests for showcasing
   - Stress tests for performance limits
   - Queue stress tests for visualization
2. Add real-time monitoring of load tests through the dashboard
3. Create automated reporting of test results
4. Add safety limits and circuit breakers
5. Integrate with health checker for test validation
6. Add performance benchmarking capabilities

The integration should provide comprehensive insights into system performance under load.
```

**💡 Explanation**: This prompt integrates the load testing with the monitoring infrastructure, creating a complete performance testing suite that provides actionable insights.

**✅ Validation**:
```bash
make load-test-quick
make load-test-queue  # Watch dashboard at http://localhost:3001
```

---

## Phase 5: Testing Framework

**🎯 Goal**: Create a comprehensive testing framework for all components.

### AI Prompt 5.1: Testing Infrastructure

```
Create a comprehensive testing framework:

1. Create test/ directory with structure:
   - unit/: Unit tests for Python and Go components
   - integration/: Cross-service integration tests
   - e2e/: End-to-end system tests
   - fixtures/: Test data and configurations
   - mocks/: Mock services for testing
2. Create test Makefile with targets for different test types
3. Add Go tests for all services with proper coverage
4. Add Python tests for traffic generation and utilities
5. Create integration tests that verify service interactions
6. Add performance benchmarks
7. Include test data management and cleanup

The testing framework should ensure reliability and catch regressions.
```

**💡 Explanation**: This prompt creates a production-ready testing framework. The AI will implement various testing strategies needed for a complex distributed system.

**✅ Validation**:
```bash
cd test
make test-unit
make test-integration
make test-coverage
```

### AI Prompt 5.2: Continuous Integration Setup

```
Add CI/CD support and testing automation:

1. Create comprehensive test scripts that can run in CI
2. Add test health checking and validation
3. Create performance regression detection
4. Add automated dependency management
5. Create test reporting and metrics
6. Add test environment setup and teardown
7. Include linting and code quality checks

The CI setup should ensure code quality and prevent regressions.
```

**💡 Explanation**: This prompt adds automation and quality assurance to the development process, ensuring the system remains reliable as it evolves.

**✅ Validation**:
```bash
make test-ci
make lint
make validate
```

---

## Phase 6: Configuration Management

**🎯 Goal**: Centralize and standardize configuration across all components.

### AI Prompt 6.1: Centralized Configuration

```
Create a centralized configuration management system:

1. Create config/ directory with:
   - llama-metrics.yml: Main configuration file
   - services/: Service-specific configurations
   - prometheus/: Monitoring configurations
   - alerts/: Alert rules and notifications
2. Create Python configuration manager (python/shared/config_manager.py)
3. Update all Go services to use centralized config
4. Add environment variable overrides
5. Create configuration validation and defaults
6. Add configuration reloading capabilities
7. Document all configuration options

The configuration system should be flexible and environment-aware.
```

**💡 Explanation**: This prompt centralizes configuration management, which is crucial for managing complex systems across different environments (development, staging, production).

**✅ Validation**:
```bash
python -c "from python.shared.config_manager import get_config; print(get_config().server.proxy_port)"
```

---

## Phase 7: Organization and Structure

**🎯 Goal**: Organize the codebase for better maintainability and clarity.

### AI Prompt 7.1: Script Organization

```
Organize all scripts into logical categories:

1. Create scripts/ directory structure:
   - traffic/: All traffic generation scripts
   - monitoring/: Monitoring and health utilities
   - deployment/: Installation and setup scripts
2. Move existing scripts to appropriate directories
3. Add README files for each script category
4. Create wrapper scripts for common workflows
5. Update Makefile to reference new script locations
6. Add script documentation and usage examples
7. Ensure all scripts follow consistent patterns

The organization should make the project easy to navigate and use.
```

**💡 Explanation**: This prompt organizes the growing codebase into logical categories, making it easier for team members to find and use different components.

**✅ Validation**:
```bash
ls scripts/*/
make traffic  # Should work with new script organization
```

### AI Prompt 7.2: Python Package Organization

```
Organize Python components into a proper package structure:

1. Create python/ directory with:
   - shared/: Reusable utilities (config_manager.py, version.py)
   - container/: Docker and containerization files
   - requirements*.txt: All dependency files
2. Move all Python files to appropriate locations
3. Update Dockerfile to work with new structure
4. Update Makefile references to new paths
5. Ensure all imports work correctly
6. Add __init__.py files for proper packages
7. Update documentation to reflect new structure

The Python organization should follow best practices for package structure.
```

**💡 Explanation**: This prompt organizes Python components following standard packaging conventions, making the codebase more professional and maintainable.

**✅ Validation**:
```bash
python -c "import sys; sys.path.append('python/shared'); import version; print(version.__version__)"
make install  # Should work with new requirements paths
```

---

## Phase 8: Documentation and User Experience

**🎯 Goal**: Create comprehensive documentation and improve user experience.

### AI Prompt 8.1: Documentation Structure

```
Create comprehensive documentation:

1. Create docs/ directory structure:
   - setup/: Installation and quick start guides
   - architecture/: System design documentation
   - api/: API reference and integration guides
   - development/: Contributing and development guides
2. Add comprehensive README files
3. Create API documentation for all services
4. Add troubleshooting guides
5. Create performance tuning documentation
6. Add examples and tutorials
7. Include screenshots and diagrams

The documentation should enable new users to get started quickly and developers to contribute effectively.
```

**💡 Explanation**: This prompt creates professional-grade documentation that makes the project accessible to others and serves as a knowledge base for the system.

**✅ Validation**:
```bash
ls docs/*/
cat docs/setup/quick_start.md
```

### AI Prompt 8.2: Developer Experience

```
Create CLAUDE.md file for future AI coding assistant sessions:

1. Analyze the current codebase and architecture
2. Document the key components and their relationships
3. Include essential commands for development
4. Add configuration management guidance
5. Document the Go workspace structure
6. Include Python environment requirements
7. Add troubleshooting and debugging guidance
8. Include performance optimization notes

The CLAUDE.md should enable future AI sessions to be immediately productive.
```

**💡 Explanation**: This prompt creates documentation specifically for AI coding assistants, enabling consistent and productive future interactions with the codebase.

**✅ Validation**:
```bash
cat CLAUDE.md
# Should contain comprehensive guidance for AI assistants
```

---

## Phase 9: Container and Deployment

**🎯 Goal**: Add containerization and deployment capabilities.

### AI Prompt 9.1: Containerization

```
Add Docker containerization support:

1. Create optimized Dockerfile for Python components
2. Add docker-compose.yml for full stack deployment
3. Create container health checks
4. Add environment variable configuration
5. Optimize container size and security
6. Add multi-stage builds if beneficial
7. Create deployment scripts and documentation
8. Test container functionality

The containerization should enable easy deployment across different environments.
```

**💡 Explanation**: This prompt adds containerization capabilities, which are essential for modern deployment and scaling of applications.

**✅ Validation**:
```bash
docker build -t llama-metrics -f python/container/Dockerfile .
docker run --rm -p 8001:8001 llama-metrics
```

---

## Phase 10: Monitoring and Alerting

**🎯 Goal**: Add advanced monitoring and alerting capabilities.

### AI Prompt 10.1: Advanced Monitoring

```
Enhance the monitoring capabilities:

1. Add advanced Prometheus alert rules
2. Create monitoring dashboards for different user types
3. Add anomaly detection for AI model performance
4. Create automated reporting systems
5. Add integration with external monitoring systems
6. Create monitoring playbooks and runbooks
7. Add capacity planning metrics
8. Include cost monitoring if applicable

The monitoring should provide comprehensive observability for production use.
```

**💡 Explanation**: This prompt adds enterprise-grade monitoring capabilities that would be needed for production deployment of the system.

**✅ Validation**:
```bash
cat config/alerts/rules.yml
curl http://localhost:9090/alerts
```

---

## Phase 11: Performance Optimization

**🎯 Goal**: Optimize system performance and resource usage.

### AI Prompt 11.1: Performance Tuning

```
Optimize the system for production performance:

1. Add connection pooling and resource management
2. Optimize memory usage and garbage collection
3. Add caching where appropriate
4. Optimize database/storage access patterns
5. Add performance profiling capabilities
6. Create performance benchmarks and regression tests
7. Add auto-scaling capabilities where possible
8. Optimize for different deployment scenarios

The optimizations should improve throughput and reduce resource usage.
```

**💡 Explanation**: This prompt focuses on production readiness, ensuring the system can handle real-world loads efficiently.

**✅ Validation**:
```bash
make benchmark
make load-test-stress
```

---

## Phase 12: Final Integration and Polish

**🎯 Goal**: Complete the project with final integration and professional polish.

### AI Prompt 12.1: Final Integration

```
Complete the project integration and add final polish:

1. Ensure all components work together seamlessly
2. Add comprehensive end-to-end tests
3. Create demo scripts and showcase capabilities
4. Add version management and release processes
5. Create user onboarding flows
6. Add security best practices and hardening
7. Create backup and recovery procedures
8. Add final documentation polish

The final result should be a production-ready monitoring solution.
```

**💡 Explanation**: This final prompt ensures all components work together and adds the professional touches needed for a production system.

**✅ Validation**:
```bash
make demo  # Should run complete demonstration
make test  # All tests should pass
make validate  # System validation should succeed
```

---

## 🎉 Congratulations!

You've successfully built a comprehensive Ollama monitoring stack using AI-assisted development! Your system now includes:

- **Complete Monitoring**: Real-time dashboards, metrics, and health checking
- **Load Testing**: Sophisticated traffic generation and performance testing
- **Production Ready**: Proper configuration, testing, and documentation
- **Scalable Architecture**: Go services with shared libraries and proper structure
- **Developer Friendly**: Comprehensive documentation and development tools

## 🚀 Next Steps

1. **Deploy to Production**: Use the containerization and deployment guides
2. **Customize for Your Needs**: Modify configurations and add custom metrics
3. **Extend Functionality**: Add new monitoring capabilities or integrations
4. **Share and Contribute**: Use this as a template for other AI monitoring projects

## 📚 Key Learnings

This tutorial demonstrated:

- **Incremental Development**: Building complex systems step by step
- **AI Prompt Engineering**: Crafting effective prompts for different development phases
- **System Architecture**: Designing scalable and maintainable systems
- **Testing Strategy**: Implementing comprehensive testing at all levels
- **Documentation**: Creating maintainable and accessible documentation
- **Production Readiness**: Adding the capabilities needed for real-world deployment

## 🔧 Troubleshooting

If you encounter issues during any phase:

1. **Check Prerequisites**: Ensure all required tools are installed
2. **Verify Previous Steps**: Make sure earlier phases completed successfully
3. **Review Logs**: Check service logs for error messages
4. **Consult Documentation**: Refer to the generated docs/ directory
5. **Ask AI for Help**: Use follow-up prompts to resolve specific issues

Remember: The key to successful AI-assisted development is clear, specific prompts and iterative refinement!