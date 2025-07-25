# Testing Guide

This guide covers testing strategies, tools, and best practices for the llama-metrics project.

## 🧪 Testing Philosophy

Our testing approach follows the **testing pyramid** principle:

```
    /\     E2E Tests
   /  \    (Few, High-Value)
  /____\
 /      \  Integration Tests
/________\  (Some, Critical Paths)
/          \
/____________\ Unit Tests
               (Many, Fast, Isolated)
```

## 🏗️ Test Structure

### Directory Organization
```
test/
├── unit/                    # Unit tests
│   ├── config_test.py      # Python unit tests
│   └── test_priority_queue.py
├── integration/             # Integration tests
│   ├── service_integration_test.go
│   └── openai_compatibility_test.py
├── e2e/                     # End-to-end tests
│   └── (planned)
├── fixtures/                # Test data and fixtures
│   └── questions/          # Test question sets
├── mocks/                   # Mock objects and stubs
└── Makefile                # Test automation
```

### Service Tests
```
services/
├── shared/
│   ├── config/
│   │   └── config_test.go  # Shared config tests
│   └── models/
│       └── models_test.go  # Shared model tests
├── proxy/
│   └── (service-specific tests)
├── dashboard/
│   └── (service-specific tests)
└── health/
    └── (service-specific tests)
```

## 🚀 Quick Start

### Running All Tests
```bash
# From project root
cd test
make test

# Or from any service directory
cd services/proxy
make test
```

### Running Specific Test Types
```bash
# Unit tests only
make test-unit

# Integration tests only
make test-integration

# With coverage
make test-coverage

# With race detection
make test-race
```

## 🔧 Unit Tests

Unit tests verify individual components in isolation.

### Go Unit Tests

**Location**: `services/*/` directories
**Naming**: `*_test.go`
**Command**: `go test ./...`

**Example Test Structure**:
```go
func TestLoadEnvString(t *testing.T) {
    tests := []struct {
        name         string
        key          string
        defaultValue string
        envValue     string
        expected     string
    }{
        {
            name:         "environment variable exists",
            key:          "TEST_STRING",
            defaultValue: "default",
            envValue:     "custom",
            expected:     "custom",
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test setup
            os.Unsetenv(tt.key)
            if tt.envValue != "" {
                os.Setenv(tt.key, tt.envValue)
            }

            // Execute
            result := LoadEnvString(tt.key, tt.defaultValue)

            // Assert
            if result != tt.expected {
                t.Errorf("LoadEnvString() = %v, want %v", result, tt.expected)
            }

            // Cleanup
            os.Unsetenv(tt.key)
        })
    }
}
```

### Python Unit Tests

**Location**: `test/unit/`
**Naming**: `test_*.py`
**Command**: `python -m pytest test/unit/`

**Example Test Structure**:
```python
import pytest
from scripts.monitoring.health_check import HealthChecker

class TestHealthChecker:
    def test_check_service_health_success(self):
        # Arrange
        checker = HealthChecker()

        # Act
        result = checker.check_service_health("http://localhost:8080")

        # Assert
        assert result.status == "healthy"
        assert result.response_time < 1000
```

### Unit Test Best Practices

1. **Test Structure**: Arrange, Act, Assert (AAA)
2. **Naming**: Descriptive test names explaining the scenario
3. **Isolation**: Each test should be independent
4. **Cleanup**: Always clean up test artifacts
5. **Fast**: Unit tests should run quickly (< 100ms each)

## 🔗 Integration Tests

Integration tests verify that services work together correctly.

### Service Integration Tests

**Location**: `test/integration/`
**Purpose**: Test service-to-service communication
**Requirements**: Running services

**Example**:
```go
func TestProxyHealthCheck(t *testing.T) {
    proxyURL := getEnvOrDefault("PROXY_URL", defaultProxyURL)

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Get(proxyURL + "/health")
    if err != nil {
        t.Skipf("Proxy service not available at %s: %v", proxyURL, err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected status 200, got %d", resp.StatusCode)
    }

    var health models.HealthStatus
    if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
        t.Errorf("Failed to decode health response: %v", err)
    }

    if health.Status != models.StatusHealthy {
        t.Errorf("Expected healthy status, got %s", health.Status)
    }
}
```

### Running Integration Tests

**Prerequisites**:
1. All services must be running
2. Prometheus must be accessible
3. Test data must be available

**Commands**:
```bash
# Start all services first
make start

# Run integration tests
cd test
make test-integration

# Or with specific configuration
PROXY_URL=http://localhost:11435 \
DASHBOARD_URL=http://localhost:3001 \
make test-integration
```

## 🌐 End-to-End Tests

E2E tests verify complete user workflows.

### Planned E2E Tests

1. **Complete Request Flow**
   - Send request through proxy
   - Verify Ollama processing
   - Check metrics collection
   - Validate dashboard display

2. **Error Recovery**
   - Service failure scenarios
   - Network interruption handling
   - Data consistency verification

3. **Performance Scenarios**
   - Load testing
   - Concurrent request handling
   - Resource usage validation

## 📊 Test Coverage

### Generating Coverage Reports

```bash
# Coverage for all modules
cd test
make test-coverage

# Coverage for specific service
cd services/proxy
go test -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Targets

- **Unit Tests**: 80%+ coverage
- **Integration Tests**: Critical paths covered
- **Overall Project**: 70%+ coverage

### Viewing Coverage

```bash
# HTML report
go tool cover -html=coverage.out

# Terminal summary
go tool cover -func=coverage.out
```

## 🏃‍♂️ Test Automation

### Make Commands

| Command | Description | Use Case |
|---------|-------------|----------|
| `make test` | Run all tests | Development |
| `make test-unit` | Unit tests only | Quick feedback |
| `make test-integration` | Integration tests | Pre-deployment |
| `make test-coverage` | Tests with coverage | Code quality |
| `make test-race` | Race condition detection | Concurrency issues |
| `make test-bench` | Benchmark tests | Performance analysis |
| `make test-ci` | CI-optimized tests | Automated pipelines |

### Continuous Integration

**GitHub Actions Example**:
```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: 1.21

    - name: Run Unit Tests
      run: |
        cd test
        make test-unit

    - name: Start Services
      run: make start-background

    - name: Run Integration Tests
      run: |
        cd test
        make test-integration

    - name: Upload Coverage
      uses: codecov/codecov-action@v3
```

## 🐛 Debugging Tests

### Test Failures

**Common Issues**:
1. **Service not available**: Check if services are running
2. **Environment variables**: Verify test configuration
3. **Race conditions**: Run with `-race` flag
4. **Timeouts**: Increase timeout for slow tests

**Debugging Commands**:
```bash
# Verbose test output
go test -v ./...

# Run specific test
go test -v -run TestSpecificFunction ./...

# Debug with race detection
go test -race -v ./...

# Test with coverage and race detection
go test -race -coverprofile=coverage.out ./...
```

### Environment Setup

**Test Environment Variables**:
```bash
export PROXY_URL="http://localhost:11435"
export DASHBOARD_URL="http://localhost:3001"
export HEALTH_URL="http://localhost:8080"
export METRICS_URL="http://localhost:8001"
export PROMETHEUS_URL="http://localhost:9090"
export TEST_TIMEOUT="30m"
```

## 📝 Writing Good Tests

### Test Naming Conventions

**Go Tests**:
- `TestFunctionName_Scenario_ExpectedBehavior`
- `TestLoadConfig_InvalidFile_ReturnsError`
- `TestHealthCheck_ServiceUp_ReturnsHealthy`

**Python Tests**:
- `test_function_name_scenario_expected_behavior`
- `test_load_config_invalid_file_returns_error`
- `test_health_check_service_up_returns_healthy`

### Test Data Management

**Fixtures**:
```go
// Use test fixtures for consistent data
func loadTestFixture(filename string) []byte {
    data, err := ioutil.ReadFile(filepath.Join("fixtures", filename))
    if err != nil {
        panic(err)
    }
    return data
}
```

**Environment Isolation**:
```go
func TestWithCleanEnvironment(t *testing.T) {
    // Save original environment
    originalValue := os.Getenv("TEST_VAR")
    defer os.Setenv("TEST_VAR", originalValue)

    // Set test value
    os.Setenv("TEST_VAR", "test-value")

    // Run test...
}
```

### Mock Objects

**Interface-based Mocking**:
```go
type MockOllamaClient struct {
    responses map[string]models.OllamaResponse
}

func (m *MockOllamaClient) Generate(req models.OllamaRequest) (models.OllamaResponse, error) {
    if resp, ok := m.responses[req.Model]; ok {
        return resp, nil
    }
    return models.OllamaResponse{}, errors.New("model not found")
}
```

## 🚀 Performance Testing

### Benchmark Tests

```go
func BenchmarkConfigLoading(b *testing.B) {
    for i := 0; i < b.N; i++ {
        config := DefaultBaseConfig("test-service")
        _ = config.Validate()
    }
}

func BenchmarkMetricsCollection(b *testing.B) {
    collector := NewMetricsCollector()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        collector.RecordRequest("test", time.Now(), 200)
    }
}
```

**Running Benchmarks**:
```bash
go test -bench=. -benchmem ./...
go test -bench=BenchmarkSpecific -count=5 ./...
```

### Load Testing

**Using Scripts**:
```bash
# Generate load for testing
./scripts/traffic/simple.sh --concurrent 10 --requests 100

# Monitor during load testing
make monitor
```

## 🔍 Test Monitoring

### Test Metrics

Track test health over time:
- Test execution time
- Failure rates
- Coverage trends
- Flaky test detection

### Test Reporting

**Generate Test Reports**:
```bash
# JUnit XML for CI systems
go test -v ./... | go-junit-report > test-report.xml

# Coverage badge generation
go test -coverprofile=coverage.out ./...
gocov convert coverage.out | gocov-xml > coverage.xml
```

## 🛠️ Tools and Dependencies

### Go Testing Tools

- **Testing Framework**: Built-in `testing` package
- **Assertions**: Standard library + custom helpers
- **Mocking**: Interface-based mocking
- **Coverage**: Built-in coverage tools
- **Benchmarking**: Built-in benchmark support

### Python Testing Tools

- **Framework**: pytest
- **Coverage**: pytest-cov
- **Mocking**: unittest.mock
- **Fixtures**: pytest fixtures

### Additional Tools

- **Race Detection**: Built into Go toolchain
- **Memory Profiling**: `go tool pprof`
- **Load Testing**: Custom scripts + external tools
- **CI Integration**: GitHub Actions, GitLab CI

## 📚 Best Practices Summary

### Do's ✅

1. **Write tests first** for new features (TDD)
2. **Keep tests simple** and focused
3. **Use descriptive names** for tests and variables
4. **Clean up** after tests (reset environment)
5. **Test edge cases** and error conditions
6. **Mock external dependencies** in unit tests
7. **Use table-driven tests** for multiple scenarios
8. **Run tests frequently** during development

### Don'ts ❌

1. **Don't test implementation details** - test behavior
2. **Don't write flaky tests** - ensure consistency
3. **Don't skip test cleanup** - avoid test pollution
4. **Don't hardcode values** - use constants and fixtures
5. **Don't ignore failing tests** - fix or investigate
6. **Don't write slow unit tests** - keep them fast
7. **Don't duplicate test logic** - extract common helpers

## 🆘 Getting Help

### Test Issues

1. **Check test logs**: Look for specific error messages
2. **Verify environment**: Ensure services are running
3. **Run tests in isolation**: Identify specific problems
4. **Use debugging tools**: Add logging, use debugger
5. **Ask for help**: Create issue with test failure details

### Resources

- **Go Testing**: https://golang.org/pkg/testing/
- **pytest Documentation**: https://docs.pytest.org/
- **Testing Best Practices**: Internal team guidelines
- **Project Issues**: GitHub issue tracker

---

Remember: **Good tests are the foundation of reliable software!** 🏗️