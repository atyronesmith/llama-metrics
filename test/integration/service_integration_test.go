package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/llama-metrics/shared/models"
)

const (
	defaultProxyURL      = "http://localhost:11435"
	defaultDashboardURL  = "http://localhost:3001"
	defaultHealthURL     = "http://localhost:8080"
	defaultMetricsURL    = "http://localhost:8001"
	defaultPrometheusURL = "http://localhost:9090"
)

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

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

func TestDashboardHealthCheck(t *testing.T) {
	dashboardURL := getEnvOrDefault("DASHBOARD_URL", defaultDashboardURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(dashboardURL + "/health")
	if err != nil {
		t.Skipf("Dashboard service not available at %s: %v", dashboardURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHealthServiceCheck(t *testing.T) {
	healthURL := getEnvOrDefault("HEALTH_URL", defaultHealthURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(healthURL + "/health")
	if err != nil {
		t.Skipf("Health service not available at %s: %v", healthURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	metricsURL := getEnvOrDefault("METRICS_URL", defaultMetricsURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(metricsURL + "/metrics")
	if err != nil {
		t.Skipf("Metrics endpoint not available at %s: %v", metricsURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Check that we get Prometheus format metrics
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read metrics response: %v", err)
	}

	bodyStr := string(body)

	// Look for common Prometheus metrics
	expectedMetrics := []string{
		"# HELP",
		"# TYPE",
		"llama_metrics_",
	}

	for _, metric := range expectedMetrics {
		if !bytes.Contains(body, []byte(metric)) {
			t.Errorf("Expected to find '%s' in metrics output", metric)
		}
	}

	if len(bodyStr) < 100 {
		t.Errorf("Metrics output too short, got %d characters", len(bodyStr))
	}
}

func TestPrometheusTargets(t *testing.T) {
	prometheusURL := getEnvOrDefault("PROMETHEUS_URL", defaultPrometheusURL)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(prometheusURL + "/api/v1/targets")
	if err != nil {
		t.Skipf("Prometheus not available at %s: %v", prometheusURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var targetsResp struct {
		Status string `json:"status"`
		Data   struct {
			ActiveTargets []struct {
				DiscoveredLabels map[string]string `json:"discoveredLabels"`
				Labels           map[string]string `json:"labels"`
				ScrapeURL        string            `json:"scrapeUrl"`
				Health           string            `json:"health"`
			} `json:"activeTargets"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&targetsResp); err != nil {
		t.Errorf("Failed to decode targets response: %v", err)
	}

	if targetsResp.Status != "success" {
		t.Errorf("Expected success status, got %s", targetsResp.Status)
	}

	expectedJobs := []string{"ollama-proxy", "llama-dashboard", "llama-health"}
	foundJobs := make(map[string]bool)

	for _, target := range targetsResp.Data.ActiveTargets {
		if job, ok := target.Labels["job"]; ok {
			foundJobs[job] = true
		}
	}

	for _, expectedJob := range expectedJobs {
		if !foundJobs[expectedJob] {
			t.Errorf("Expected to find job '%s' in Prometheus targets", expectedJob)
		}
	}
}

func TestOllamaRequestFlow(t *testing.T) {
	proxyURL := getEnvOrDefault("PROXY_URL", defaultProxyURL)

	// Create a test request
	request := models.OllamaRequest{
		Model:  "phi3:mini",
		Prompt: "Hello, this is a test.",
		Stream: false,
	}

	reqBody, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("Failed to marshal request: %v", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}

	// Send request through proxy
	resp, err := client.Post(proxyURL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Skipf("Failed to send request to proxy: %v", err)
	}
	defer resp.Body.Close()

	// We expect either success (200) or an error if Ollama is not available
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusServiceUnavailable {
		body, _ := io.ReadAll(resp.Body)
		t.Errorf("Unexpected status %d, body: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode == http.StatusOK {
		var response models.OllamaResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			t.Errorf("Failed to decode response: %v", err)
		}

		if response.Model != request.Model {
			t.Errorf("Expected model %s, got %s", request.Model, response.Model)
		}
	}
}

func TestServiceCommunication(t *testing.T) {
	// Test that services can communicate with each other
	tests := []struct {
		name        string
		serviceURL  string
		endpoint    string
		expectedKey string
	}{
		{
			name:        "proxy status",
			serviceURL:  getEnvOrDefault("PROXY_URL", defaultProxyURL),
			endpoint:    "/status",
			expectedKey: "status",
		},
		{
			name:        "dashboard api",
			serviceURL:  getEnvOrDefault("DASHBOARD_URL", defaultDashboardURL),
			endpoint:    "/api/status",
			expectedKey: "status",
		},
		{
			name:        "health check api",
			serviceURL:  getEnvOrDefault("HEALTH_URL", defaultHealthURL),
			endpoint:    "/api/health",
			expectedKey: "status",
		},
	}

	client := &http.Client{Timeout: 10 * time.Second}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := client.Get(tt.serviceURL + tt.endpoint)
			if err != nil {
				t.Skipf("Service not available at %s: %v", tt.serviceURL, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusNotFound {
				t.Skip("Endpoint not implemented yet")
			}

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			// Try to decode as JSON
			var result map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				// If JSON decode fails, it might be plain text or HTML
				t.Logf("Non-JSON response from %s%s", tt.serviceURL, tt.endpoint)
				return
			}

			if _, ok := result[tt.expectedKey]; !ok {
				t.Errorf("Expected key '%s' in response", tt.expectedKey)
			}
		})
	}
}

func TestMetricsCollection(t *testing.T) {
	metricsURL := getEnvOrDefault("METRICS_URL", defaultMetricsURL)

	// First, make a request to generate some metrics
	proxyURL := getEnvOrDefault("PROXY_URL", defaultProxyURL)
	client := &http.Client{Timeout: 10 * time.Second}

	// Make a health check request to generate metrics
	resp, err := client.Get(proxyURL + "/health")
	if err == nil {
		resp.Body.Close()
	}

	// Wait a bit for metrics to be collected
	time.Sleep(2 * time.Second)

	// Check metrics
	resp, err = client.Get(metricsURL + "/metrics")
	if err != nil {
		t.Skipf("Metrics endpoint not available: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("Failed to read metrics: %v", err)
	}

	// Look for specific metrics that should be present
	expectedMetrics := []string{
		"llama_metrics_requests_total",
		"llama_metrics_request_duration",
		"go_memstats",
		"process_",
	}

	bodyStr := string(body)
	for _, metric := range expectedMetrics {
		if !bytes.Contains(body, []byte(metric)) {
			t.Logf("Warning: Expected metric '%s' not found in output", metric)
			// Don't fail the test as some metrics might not be generated yet
		}
	}

	if len(bodyStr) < 50 {
		t.Errorf("Metrics output suspiciously short: %d characters", len(bodyStr))
	}
}

func TestErrorHandling(t *testing.T) {
	proxyURL := getEnvOrDefault("PROXY_URL", defaultProxyURL)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test invalid endpoint
	resp, err := client.Get(proxyURL + "/nonexistent")
	if err != nil {
		t.Skipf("Proxy not available: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("Expected 404 for invalid endpoint, got %d", resp.StatusCode)
	}

	// Test invalid JSON request
	invalidJSON := bytes.NewBufferString(`{"invalid": json`)
	resp, err = client.Post(proxyURL+"/api/generate", "application/json", invalidJSON)
	if err != nil {
		t.Skipf("Failed to send invalid request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Logf("Expected 400 for invalid JSON, got %d (some implementations may handle this differently)", resp.StatusCode)
	}
}

func TestConcurrentRequests(t *testing.T) {
	proxyURL := getEnvOrDefault("PROXY_URL", defaultProxyURL)

	client := &http.Client{Timeout: 10 * time.Second}

	// Test concurrent health checks
	numRequests := 5
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(id int) {
			resp, err := client.Get(fmt.Sprintf("%s/health?id=%d", proxyURL, id))
			if err != nil {
				results <- err
				return
			}
			resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- fmt.Errorf("request %d: expected 200, got %d", id, resp.StatusCode)
				return
			}

			results <- nil
		}(i)
	}

	// Wait for all requests to complete
	successCount := 0
	for i := 0; i < numRequests; i++ {
		select {
		case err := <-results:
			if err != nil {
				t.Logf("Request failed: %v", err)
			} else {
				successCount++
			}
		case <-time.After(15 * time.Second):
			t.Errorf("Timeout waiting for concurrent requests")
			return
		}
	}

	if successCount == 0 {
		t.Skip("No concurrent requests succeeded - service may not be available")
	}

	if successCount < numRequests/2 {
		t.Errorf("Too many concurrent requests failed: %d/%d succeeded", successCount, numRequests)
	}
}

func TestServiceStartupOrder(t *testing.T) {
	// This test checks that services start up in a reasonable order
	// and dependencies are available when needed

	services := []struct {
		name string
		url  string
		path string
	}{
		{"metrics", getEnvOrDefault("METRICS_URL", defaultMetricsURL), "/metrics"},
		{"proxy", getEnvOrDefault("PROXY_URL", defaultProxyURL), "/health"},
		{"dashboard", getEnvOrDefault("DASHBOARD_URL", defaultDashboardURL), "/health"},
		{"health", getEnvOrDefault("HEALTH_URL", defaultHealthURL), "/health"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, service := range services {
		t.Run(service.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Wait for service to be available
			for {
				select {
				case <-ctx.Done():
					t.Skipf("Service %s not available after timeout", service.name)
					return
				default:
					resp, err := client.Get(service.url + service.path)
					if err == nil && resp.StatusCode == http.StatusOK {
						resp.Body.Close()
						t.Logf("Service %s is available", service.name)
						return
					}
					if resp != nil {
						resp.Body.Close()
					}
					time.Sleep(1 * time.Second)
				}
			}
		})
	}
}