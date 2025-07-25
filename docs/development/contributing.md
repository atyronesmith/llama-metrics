# Contributing Guide

Thank you for your interest in contributing to llama-metrics! This guide will help you get started with contributing to the project.

## 🤝 Ways to Contribute

### 🐛 Bug Reports
- Search existing issues first
- Use the bug report template
- Include system information and steps to reproduce
- Provide logs and error messages

### ✨ Feature Requests
- Check if the feature already exists or is planned
- Use the feature request template
- Explain the use case and expected behavior
- Consider starting a discussion first

### 📝 Documentation
- Fix typos and improve clarity
- Add missing documentation
- Update outdated information
- Create tutorials and examples

### 💻 Code Contributions
- Fix bugs and implement features
- Improve performance and reliability
- Add tests and monitoring
- Refactor and clean up code

## 🚀 Getting Started

### 1. Fork and Clone

```bash
# Fork the repository on GitHub, then:
git clone https://github.com/YOUR-USERNAME/llama-metrics.git
cd llama-metrics
```

### 2. Set Up Development Environment

```bash
# Install dependencies
./scripts/deployment/install.sh --dev

# Verify setup
make verify

# Run tests
make test
```

### 3. Create a Branch

```bash
# Create a feature branch
git checkout -b feature/your-feature-name

# Or a bugfix branch
git checkout -b bugfix/issue-number-description
```

## 🛠️ Development Workflow

### Project Structure
```
llama-metrics/
├── services/           # Go services
│   ├── shared/        # Shared packages
│   ├── proxy/         # Monitoring proxy
│   ├── dashboard/     # Web dashboard
│   └── health/        # Health checker
├── scripts/           # Automation scripts
│   ├── traffic/       # Traffic generation
│   ├── monitoring/    # Monitoring tools
│   └── deployment/    # Installation scripts
├── config/            # Configuration files
├── docs/             # Documentation
└── test/             # Test files
```

### Development Guidelines

#### Go Services
- **Go Version**: Use Go 1.21+ (1.24.4 for dashboard)
- **Code Style**: Use `gofmt` and `golangci-lint`
- **Testing**: Write unit tests for new functionality
- **Documentation**: Include function/package comments

#### Scripts
- **Shell Scripts**: Must pass `shellcheck`
- **Python Scripts**: Use Python 3.8+, follow PEP 8
- **Documentation**: Include usage examples

#### Configuration
- **Format**: Use YAML for configuration files
- **Validation**: Add validation for new config options
- **Documentation**: Document all configuration options

### Code Quality Standards

#### Testing
```bash
# Run all tests
make test

# Run specific service tests
cd services/proxy && make test

# Run linting
make lint

# Check code formatting
make fmt
```

#### Required Checks
- ✅ All tests pass
- ✅ Linting passes
- ✅ Code is properly formatted
- ✅ Documentation is updated
- ✅ Configuration is valid

### Shared Packages
When working with shared packages:

1. **Update shared package first**
2. **Update services that use it**
3. **Test all affected services**
4. **Update documentation**

## 📋 Pull Request Process

### 1. Before Submitting

- [ ] Code follows project style guidelines
- [ ] Tests pass locally
- [ ] Documentation is updated
- [ ] Commit messages are clear
- [ ] Branch is up to date with main

### 2. Commit Guidelines

```bash
# Format: type(scope): description
git commit -m "feat(proxy): add request queue metrics"
git commit -m "fix(dashboard): resolve WebSocket connection issue"
git commit -m "docs(api): update metrics endpoint documentation"
```

**Types**:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Adding or fixing tests
- `chore`: Maintenance tasks

### 3. Pull Request Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation update
- [ ] Refactoring

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests pass
- [ ] Manual testing completed

## Checklist
- [ ] Code follows style guidelines
- [ ] Self-review completed
- [ ] Documentation updated
- [ ] No breaking changes (or documented)
```

### 4. Review Process

1. **Automated Checks**: CI/CD pipeline runs tests
2. **Code Review**: Maintainer reviews the code
3. **Discussion**: Address feedback and suggestions
4. **Approval**: Maintainer approves the changes
5. **Merge**: Changes are merged to main branch

## 🧪 Testing Guidelines

### Unit Tests
```bash
# Go services
cd services/proxy
go test ./...

# Python scripts
python -m pytest test/
```

### Integration Tests
```bash
# Full stack testing
make test-integration

# Service-specific testing
make test-proxy
make test-dashboard
make test-health
```

### Manual Testing
1. **Start Services**: `make start`
2. **Generate Traffic**: `make traffic-quick`
3. **Verify Metrics**: Check dashboard and Prometheus
4. **Test Features**: Validate new functionality

## 📚 Documentation Standards

### Code Documentation
- **Go**: Use GoDoc conventions
- **Python**: Use docstrings
- **Shell**: Include header comments with usage

### Markdown Documentation
- **Structure**: Use consistent headings
- **Examples**: Include working code examples
- **Links**: Use relative links for internal docs
- **Images**: Optimize images and use alt text

### API Documentation
- **Endpoints**: Document all endpoints
- **Parameters**: Include parameter descriptions
- **Examples**: Provide request/response examples
- **Errors**: Document error conditions

## 🎯 Development Tips

### Local Development
```bash
# Hot reload for Go services
cd services/proxy && make dev

# Background services
make start-background

# Clean restart
make clean && make start
```

### Debugging
```bash
# View logs
make logs

# Check service health
make verify

# Debug specific service
cd services/proxy && go run cmd/proxy/main.go --log-level debug
```

### Performance Testing
```bash
# Load testing
./scripts/traffic/scenarios.sh

# Stress testing
make traffic-stress

# Monitor resources
make monitor
```

## 🔧 Troubleshooting Development Issues

### Common Issues

**Import Errors**
```bash
# Update Go modules
go mod tidy
go mod download
```

**Port Conflicts**
```bash
# Kill processes on ports
pkill -f ollama
pkill -f prometheus
```

**Test Failures**
```bash
# Clean test cache
go clean -testcache

# Run with verbose output
go test -v ./...
```

### Getting Help

1. **Check Documentation**: Search existing docs
2. **Search Issues**: Look for similar problems
3. **Join Discussions**: Ask questions in GitHub Discussions
4. **Contact Maintainers**: Tag maintainers in issues

## 🏷️ Release Process

### Version Management
- **Semantic Versioning**: MAJOR.MINOR.PATCH
- **Tag Format**: `v1.0.0`
- **Changelog**: Update CHANGELOG.md

### Release Checklist
- [ ] All tests pass
- [ ] Documentation updated
- [ ] Version bumped
- [ ] Changelog updated
- [ ] Release notes prepared

## 🌟 Recognition

Contributors are recognized in:
- **CONTRIBUTORS.md**: All contributors listed
- **Release Notes**: Major contributions highlighted
- **GitHub**: Contributor badges and stats

## 📞 Contact

- **Issues**: [GitHub Issues](https://github.com/atyronesmith/llama-metrics/issues)
- **Discussions**: [GitHub Discussions](https://github.com/atyronesmith/llama-metrics/discussions)
- **Email**: [Project maintainers](mailto:maintainers@example.com)

---

Thank you for contributing to llama-metrics! 🎉