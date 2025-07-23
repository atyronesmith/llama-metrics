package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/atyronesmith/llama-metrics/health/internal/models"
	"github.com/atyronesmith/llama-metrics/health/pkg/config"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// ServiceEndpoint represents a service to check
type ServiceEndpoint struct {
	Name     string
	URL      string
	Critical bool
	Timeout  time.Duration
}

// HealthChecker implements comprehensive health checking
type HealthChecker struct {
	config          *config.Config
	startTime       time.Time
	httpClient      *http.Client
	serviceEndpoints []ServiceEndpoint
	mu              sync.RWMutex
}

// NewHealthChecker creates a new health checker instance
func NewHealthChecker(cfg *config.Config) *HealthChecker {
	hc := &HealthChecker{
		config:    cfg,
		startTime: time.Now(),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}

	// Initialize service endpoints
	hc.serviceEndpoints = []ServiceEndpoint{
		{
			Name:     "ollama",
			URL:      fmt.Sprintf("%s/api/tags", cfg.Server.OllamaURL),
			Critical: true,
			Timeout:  5 * time.Second,
		},
		{
			Name:     "proxy",
			URL:      fmt.Sprintf("http://%s:%d/health", cfg.Server.MetricsHost, cfg.Server.MetricsPort),
			Critical: true,
			Timeout:  3 * time.Second,
		},
		{
			Name:     "metrics",
			URL:      fmt.Sprintf("http://%s:%d/metrics", cfg.Server.MetricsHost, cfg.Server.MetricsPort),
			Critical: false,
			Timeout:  3 * time.Second,
		},
		{
			Name:     "dashboard",
			URL:      fmt.Sprintf("http://%s:%d/api/status", cfg.Server.DashboardHost, cfg.Server.DashboardPort),
			Critical: false,
			Timeout:  3 * time.Second,
		},
	}

	return hc
}

// CheckOllamaGeneration performs comprehensive Ollama health check including generation
func (hc *HealthChecker) CheckOllamaGeneration(ctx context.Context) models.ServiceHealth {
	startTime := time.Now()

	// First, check if Ollama is listening
	resp, err := hc.httpClient.Get(fmt.Sprintf("%s/api/tags", hc.config.Server.OllamaURL))
	if err != nil {
		errStr := err.Error()
		return models.ServiceHealth{
			Name: "ollama",
			URL:  hc.config.Server.OllamaURL,
			Status: models.HealthStatus{
				Status:    "unhealthy",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Error:     &errStr,
			},
			Critical: true,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errStr := fmt.Sprintf("API endpoint not responding: HTTP %d", resp.StatusCode)
		return models.ServiceHealth{
			Name: "ollama",
			URL:  hc.config.Server.OllamaURL,
			Status: models.HealthStatus{
				Status:    "unhealthy",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Error:     &errStr,
			},
			Critical: true,
		}
	}

	// Test actual generation capability
	genStart := time.Now()

	// Create minimal generation request
	genReq := map[string]interface{}{
		"model":  hc.config.Models.DefaultModel,
		"prompt": "Hi",
		"stream": false,
		"options": map[string]interface{}{
			"num_predict": 1,
		},
	}

	reqBody, _ := json.Marshal(genReq)
	genResp, err := hc.httpClient.Post(
		fmt.Sprintf("%s/api/generate", hc.config.Server.OllamaURL),
		"application/json",
		bytes.NewReader(reqBody),
	)

	generationTime := time.Since(genStart).Milliseconds()
	totalTime := time.Since(startTime).Milliseconds()
	responseTimeMs := float64(totalTime)

	if err != nil {
		errStr := fmt.Sprintf("Generation failed: %v", err)
		return models.ServiceHealth{
			Name: "ollama",
			URL:  hc.config.Server.OllamaURL,
			Status: models.HealthStatus{
				Status:         "unhealthy",
				Timestamp:      time.Now().UTC().Format(time.RFC3339),
				ResponseTimeMs: &responseTimeMs,
				Error:          &errStr,
				Details: map[string]any{
					"generation_time_ms": generationTime,
				},
			},
			Critical: true,
		}
	}
	defer genResp.Body.Close()

	if genResp.StatusCode != http.StatusOK {
		errStr := fmt.Sprintf("Generation failed: HTTP %d", genResp.StatusCode)
		return models.ServiceHealth{
			Name: "ollama",
			URL:  hc.config.Server.OllamaURL,
			Status: models.HealthStatus{
				Status:         "unhealthy",
				Timestamp:      time.Now().UTC().Format(time.RFC3339),
				ResponseTimeMs: &responseTimeMs,
				Error:          &errStr,
				Details: map[string]any{
					"generation_time_ms": generationTime,
				},
			},
			Critical: true,
		}
	}

	// Check if we got a valid response
	var genData map[string]interface{}
	if err := json.NewDecoder(genResp.Body).Decode(&genData); err != nil {
		errStr := "Generation returned invalid JSON"
		return models.ServiceHealth{
			Name: "ollama",
			URL:  hc.config.Server.OllamaURL,
			Status: models.HealthStatus{
				Status:         "unhealthy",
				Timestamp:      time.Now().UTC().Format(time.RFC3339),
				ResponseTimeMs: &responseTimeMs,
				Error:          &errStr,
				Details: map[string]any{
					"generation_time_ms": generationTime,
				},
			},
			Critical: true,
		}
	}

	// All checks passed
	return models.ServiceHealth{
		Name: "ollama",
		URL:  hc.config.Server.OllamaURL,
		Status: models.HealthStatus{
			Status:         "healthy",
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			ResponseTimeMs: &responseTimeMs,
			Details: map[string]any{
				"generation_time_ms":  generationTime,
				"model":               hc.config.Models.DefaultModel,
				"generation_working": true,
			},
		},
		Critical: true,
	}
}

// CheckServiceHealth checks health of a single service
func (hc *HealthChecker) CheckServiceHealth(ctx context.Context, service ServiceEndpoint) models.ServiceHealth {
	// Special handling for Ollama
	if service.Name == "ollama" {
		return hc.CheckOllamaGeneration(ctx)
	}

	startTime := time.Now()

	// Create request with timeout
	req, err := http.NewRequestWithContext(ctx, "GET", service.URL, nil)
	if err != nil {
		errStr := err.Error()
		return models.ServiceHealth{
			Name: service.Name,
			URL:  service.URL,
			Status: models.HealthStatus{
				Status:    "unhealthy",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Error:     &errStr,
			},
			Critical: service.Critical,
		}
	}
	req.Header.Set("User-Agent", "HealthChecker/1.0")

	resp, err := hc.httpClient.Do(req)
	responseTime := time.Since(startTime).Milliseconds()
	responseTimeMs := float64(responseTime)

	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "timeout") {
			errStr = "Connection timeout"
		} else if strings.Contains(errStr, "refused") {
			errStr = "Connection refused"
		}

		return models.ServiceHealth{
			Name: service.Name,
			URL:  service.URL,
			Status: models.HealthStatus{
				Status:    "unhealthy",
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Error:     &errStr,
			},
			Critical: service.Critical,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return models.ServiceHealth{
			Name: service.Name,
			URL:  service.URL,
			Status: models.HealthStatus{
				Status:         "healthy",
				Timestamp:      time.Now().UTC().Format(time.RFC3339),
				ResponseTimeMs: &responseTimeMs,
			},
			Critical: service.Critical,
		}
	}

	errStr := fmt.Sprintf("HTTP %d", resp.StatusCode)
	return models.ServiceHealth{
		Name: service.Name,
		URL:  service.URL,
		Status: models.HealthStatus{
			Status:         "unhealthy",
			Timestamp:      time.Now().UTC().Format(time.RFC3339),
			ResponseTimeMs: &responseTimeMs,
			Error:          &errStr,
		},
		Critical: service.Critical,
	}
}

// GetSystemMetrics collects system metrics
func (hc *HealthChecker) GetSystemMetrics() models.SystemMetrics {
	metrics := models.SystemMetrics{}

	// CPU metrics
	cpuPercent, _ := cpu.Percent(100*time.Millisecond, false)
	if len(cpuPercent) > 0 {
		metrics.CPU.Percent = cpuPercent[0]
	}
	metrics.CPU.Count, _ = cpu.Counts(true)

	// Load average (Unix systems)
	if runtime.GOOS != "windows" {
		if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
			// Use loadavg file on Unix systems
			if loadAvg, err := getLoadAverage(); err == nil {
				metrics.CPU.LoadAvg = loadAvg
			}
		}
	}

	// Memory metrics
	if vm, err := mem.VirtualMemory(); err == nil {
		metrics.Memory.Percent = vm.UsedPercent
		metrics.Memory.TotalGB = float64(vm.Total) / (1024 * 1024 * 1024)
		metrics.Memory.AvailableGB = float64(vm.Available) / (1024 * 1024 * 1024)
		metrics.Memory.UsedGB = float64(vm.Used) / (1024 * 1024 * 1024)
	}

	// Disk metrics
	if d, err := disk.Usage("/"); err == nil {
		metrics.Disk.Percent = d.UsedPercent
		metrics.Disk.TotalGB = float64(d.Total) / (1024 * 1024 * 1024)
		metrics.Disk.FreeGB = float64(d.Free) / (1024 * 1024 * 1024)
		metrics.Disk.UsedGB = float64(d.Used) / (1024 * 1024 * 1024)
	}

	// Network metrics
	if n, err := net.IOCounters(false); err == nil && len(n) > 0 {
		metrics.Network.BytesSent = n[0].BytesSent
		metrics.Network.BytesRecv = n[0].BytesRecv
		metrics.Network.PacketsSent = n[0].PacketsSent
		metrics.Network.PacketsRecv = n[0].PacketsRecv
	}

	// macOS specific metrics
	if runtime.GOOS == "darwin" {
		// GPU metrics
		gpu := &models.GPUMetrics{Available: false}
		if output, err := exec.Command("system_profiler", "SPDisplaysDataType", "-json").Output(); err == nil {
			var data map[string]interface{}
			if json.Unmarshal(output, &data) == nil {
				gpu.Available = true
				if displays, ok := data["SPDisplaysDataType"].([]interface{}); ok {
					gpu.Data = displays
				}
			}
		}
		metrics.GPU = gpu

		// Power metrics
		power := &models.PowerMetrics{Available: false}
		if output, err := exec.Command("pmset", "-g", "ps").Output(); err == nil {
			power.Available = true
			power.BatteryInfo = string(output)
		}
		metrics.Power = power
	}

	return metrics
}

// getLoadAverage returns the system load average
func getLoadAverage() ([]float64, error) {
	// Try sysctl on macOS
	if runtime.GOOS == "darwin" {
		output, err := exec.Command("sysctl", "-n", "vm.loadavg").Output()
		if err == nil {
			var load1, load5, load15 float64
			if _, err := fmt.Sscanf(strings.TrimSpace(string(output)), "{ %f %f %f }", &load1, &load5, &load15); err == nil {
				return []float64{load1, load5, load15}, nil
			}
		}
	}

	// Try /proc/loadavg on Linux
	if runtime.GOOS == "linux" {
		data, err := os.ReadFile("/proc/loadavg")
		if err == nil {
			var load1, load5, load15 float64
			if _, err := fmt.Sscanf(string(data), "%f %f %f", &load1, &load5, &load15); err == nil {
				return []float64{load1, load5, load15}, nil
			}
		}
	}

	return nil, fmt.Errorf("load average not available")
}

// GetComprehensiveHealth returns comprehensive system health
func (hc *HealthChecker) GetComprehensiveHealth(ctx context.Context) models.SystemHealth {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	uptime := time.Since(hc.startTime).Seconds()

	// Check all services concurrently
	var wg sync.WaitGroup
	serviceChan := make(chan models.ServiceHealth, len(hc.serviceEndpoints))

	for _, service := range hc.serviceEndpoints {
		wg.Add(1)
		go func(svc ServiceEndpoint) {
			defer wg.Done()
			serviceChan <- hc.CheckServiceHealth(ctx, svc)
		}(service)
	}

	wg.Wait()
	close(serviceChan)

	// Collect results
	var services []models.ServiceHealth
	criticalFailures := 0
	totalFailures := 0

	for service := range serviceChan {
		services = append(services, service)
		if service.Status.Status != "healthy" {
			totalFailures++
			if service.Critical {
				criticalFailures++
			}
		}
	}

	// Determine overall status
	var overallStatus string
	if criticalFailures > 0 {
		overallStatus = "unhealthy"
	} else if totalFailures > 0 {
		overallStatus = "degraded"
	} else {
		overallStatus = "healthy"
	}

	// Get system metrics
	systemMetrics := hc.GetSystemMetrics()

	// Create summary
	healthyServices := len(services) - totalFailures
	summary := map[string]interface{}{
		"overall_status":    overallStatus,
		"services_healthy":  healthyServices,
		"services_total":    len(services),
		"critical_failures": criticalFailures,
		"uptime_seconds":    uptime,
		"version":           os.Getenv("VERSION"),
	}

	return models.SystemHealth{
		Status:        overallStatus,
		Timestamp:     timestamp,
		Version:       os.Getenv("VERSION"),
		UptimeSeconds: uptime,
		Services:      services,
		SystemMetrics: systemMetrics,
		Summary:       summary,
	}
}

// GetSimpleHealth returns simple health status
func (hc *HealthChecker) GetSimpleHealth() models.SimpleHealth {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	uptime := time.Since(hc.startTime).Seconds()

	// Quick system check
	cpuPercent, _ := cpu.Percent(0, false)
	var cpuPct float64
	if len(cpuPercent) > 0 {
		cpuPct = cpuPercent[0]
	}

	memInfo, _ := mem.VirtualMemory()

	return models.SimpleHealth{
		Status:        "healthy",
		Timestamp:     timestamp,
		Version:       os.Getenv("VERSION"),
		UptimeSeconds: uptime,
		System: map[string]interface{}{
			"cpu_percent":    cpuPct,
			"memory_percent": memInfo.UsedPercent,
		},
	}
}

// GetReadinessStatus returns readiness status
func (hc *HealthChecker) GetReadinessStatus() models.ReadinessStatus {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	components := make(map[string]string)
	ready := true

	// Check configuration
	if hc.config != nil {
		components["config"] = "ready"
	} else {
		components["config"] = "failed: no configuration"
		ready = false
	}

	// Check metrics collection
	if _, err := cpu.Percent(0, false); err != nil {
		components["metrics_collection"] = fmt.Sprintf("failed: %v", err)
		ready = false
	} else {
		components["metrics_collection"] = "ready"
	}

	return models.ReadinessStatus{
		Ready:      ready,
		Timestamp:  timestamp,
		Components: components,
	}
}

// GetLivenessStatus returns liveness status
func (hc *HealthChecker) GetLivenessStatus() models.LivenessStatus {
	timestamp := time.Now().UTC().Format(time.RFC3339)
	uptime := time.Since(hc.startTime).Seconds()

	return models.LivenessStatus{
		Alive:         true,
		Timestamp:     timestamp,
		UptimeSeconds: uptime,
	}
}

// GetAnalyzedHealth returns comprehensive health with LLM analysis
func (hc *HealthChecker) GetAnalyzedHealth(ctx context.Context) models.AnalyzedHealth {
	// First get the comprehensive health
	health := hc.GetComprehensiveHealth(ctx)

	// Create analyzed health
	analyzed := models.AnalyzedHealth{
		SystemHealth: health,
	}

	// Get LLM analysis if available
	analysis := hc.AnalyzeHealthWithLLM(ctx, health)
	analyzed.Analysis = &analysis

	return analyzed
}