//go:build darwin
// +build darwin

package collector

import (
	"context"
	"log"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// PowermeticsData is no longer used - we parse text output directly

// MacMetrics holds all the Mac system metrics
type MacMetrics struct {
	GPUUtilization  float64 `json:"gpu_utilization"`
	GPUPower        float64 `json:"gpu_power"`
	CPUPower        float64 `json:"cpu_power"`
	CPUTemperature  float64 `json:"cpu_temperature"`
	MemoryPressure  float64 `json:"memory_pressure"`
	ThermalPressure string  `json:"thermal_pressure"`
	Timestamp       float64 `json:"timestamp"`
}

// Collector manages the collection of Mac system metrics
type Collector struct {
	mu      sync.RWMutex
	metrics MacMetrics

	// Prometheus metrics
	gpuUtilization  prometheus.Gauge
	gpuPower        prometheus.Gauge
	cpuPower        prometheus.Gauge
	cpuTemperature  prometheus.Gauge
	memoryPressure  prometheus.Gauge
	thermalPressure prometheus.Gauge

	// Collection settings
	collectInterval time.Duration
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewCollector creates a new metrics collector
func NewCollector() *Collector {
	ctx, cancel := context.WithCancel(context.Background())

	c := &Collector{
		collectInterval: 5 * time.Second,
		ctx:             ctx,
		cancel:          cancel,
		gpuUtilization: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "mac_gpu_utilization_percent",
			Help: "GPU utilization percentage",
		}),
		gpuPower: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "mac_gpu_power_watts",
			Help: "GPU power consumption in watts",
		}),
		cpuPower: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "mac_cpu_power_watts",
			Help: "CPU power consumption in watts",
		}),
		cpuTemperature: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "mac_cpu_temperature_celsius",
			Help: "CPU temperature in Celsius",
		}),
		memoryPressure: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "mac_memory_pressure_percent",
			Help: "Memory pressure percentage",
		}),
		thermalPressure: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "mac_thermal_pressure_info",
			Help: "Thermal pressure state (0=nominal, 1=fair, 2=serious, 3=critical)",
		}),
	}

	// Initialize with default values
	c.metrics = MacMetrics{
		ThermalPressure: "nominal",
		Timestamp:       float64(time.Now().Unix()),
	}

	return c
}

// Start begins the metrics collection loop
func (c *Collector) Start() {
	go c.collectLoop()
}

// Stop stops the metrics collection
func (c *Collector) Stop() {
	c.cancel()
}

// GetMetrics returns the current metrics
func (c *Collector) GetMetrics() MacMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.metrics
}

// collectLoop runs the periodic collection of metrics
func (c *Collector) collectLoop() {
	ticker := time.NewTicker(c.collectInterval)
	defer ticker.Stop()

	// Collect once immediately
	c.collectMetrics()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.collectMetrics()
		}
	}
}

// collectMetrics collects all available metrics
func (c *Collector) collectMetrics() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update timestamp
	c.metrics.Timestamp = float64(time.Now().Unix())

	// Collect powermetrics (GPU and CPU power)
	c.collectPowermetrics()

	// Collect CPU temperature
	c.collectCPUTemperature()

	// Collect memory pressure
	c.collectMemoryPressure()

	// Update Prometheus metrics
	c.updatePrometheusMetrics()
}

// collectPowermetrics runs powermetrics to get GPU and CPU power data
func (c *Collector) collectPowermetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", "-n", "powermetrics",
		"--samplers", "gpu_power,cpu_power,thermal",
		"--sample-count", "1",
		"--sample-rate", "1000")

	output, err := cmd.Output()
	if err != nil {
		// Log debug info for powermetrics failures (expected without sudo)
		log.Printf("Powermetrics unavailable (requires sudo): %v", err)
		return
	}

	// Parse the text output
	outputStr := string(output)
	
	// Parse GPU metrics
	if gpuUtilization := c.parseGPUUtilization(outputStr); gpuUtilization >= 0 {
		c.metrics.GPUUtilization = gpuUtilization
	}
	
	if gpuPower := c.parseGPUPower(outputStr); gpuPower > 0 {
		c.metrics.GPUPower = gpuPower
	}

	// Parse CPU power
	if cpuPower := c.parseCPUPower(outputStr); cpuPower > 0 {
		c.metrics.CPUPower = cpuPower
	}

	// Parse thermal pressure
	if thermalPressure := c.parseThermalPressure(outputStr); thermalPressure != "" {
		c.metrics.ThermalPressure = thermalPressure
	}
}

// collectCPUTemperature gets CPU temperature using available tools
func (c *Collector) collectCPUTemperature() {
	// Try osx-cpu-temp first (if installed)
	if temp := c.runOSXCPUTemp(); temp > 0 {
		c.metrics.CPUTemperature = temp
		return
	}

	// Fallback to powermetrics SMC data
	c.collectTemperatureFromPowermetrics()
}

// runOSXCPUTemp tries to get temperature from osx-cpu-temp tool
func (c *Collector) runOSXCPUTemp() float64 {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "osx-cpu-temp")
	output, err := cmd.Output()
	if err != nil {
		// Tool not installed, this is expected
		return 0
	}

	// Parse output like "45.5°C"
	re := regexp.MustCompile(`([\d.]+)°C`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		if temp, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return temp
		}
	}

	return 0
}

// collectTemperatureFromPowermetrics gets temperature from powermetrics SMC data
func (c *Collector) collectTemperatureFromPowermetrics() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sudo", "-n", "powermetrics",
		"--samplers", "smc",
		"-n", "1",
		"-i", "1000")

	output, err := cmd.Output()
	if err != nil {
		// Expected without sudo
		return
	}

	// Look for CPU die temperature in the output
	re := regexp.MustCompile(`CPU die temperature:\s+([\d.]+)\s+C`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		if temp, err := strconv.ParseFloat(matches[1], 64); err == nil {
			c.metrics.CPUTemperature = temp
		}
	}
}

// collectMemoryPressure gets memory pressure statistics
func (c *Collector) collectMemoryPressure() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "memory_pressure")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error getting memory pressure: %v", err)
		return
	}

	// Parse memory pressure output
	re := regexp.MustCompile(`System-wide memory free percentage:\s+([\d.]+)%`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		if freePercent, err := strconv.ParseFloat(matches[1], 64); err == nil {
			c.metrics.MemoryPressure = 100 - freePercent
		}
	}
}

// updatePrometheusMetrics updates all Prometheus gauge metrics
func (c *Collector) updatePrometheusMetrics() {
	c.gpuUtilization.Set(c.metrics.GPUUtilization)
	c.gpuPower.Set(c.metrics.GPUPower)
	c.cpuPower.Set(c.metrics.CPUPower)
	c.cpuTemperature.Set(c.metrics.CPUTemperature)
	c.memoryPressure.Set(c.metrics.MemoryPressure)

	// Convert thermal pressure to numeric value
	thermalValue := 0.0
	switch strings.ToLower(c.metrics.ThermalPressure) {
	case "nominal":
		thermalValue = 0.0
	case "fair":
		thermalValue = 1.0
	case "serious":
		thermalValue = 2.0
	case "critical":
		thermalValue = 3.0
	}
	c.thermalPressure.Set(thermalValue)
}

// parseGPUUtilization extracts GPU utilization from powermetrics output
func (c *Collector) parseGPUUtilization(output string) float64 {
	// Look for "GPU idle residency: XX.XX%"
	re := regexp.MustCompile(`GPU idle residency:\s+([\d.]+)%`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		if idlePercent, err := strconv.ParseFloat(matches[1], 64); err == nil {
			// Convert idle percentage to utilization percentage
			return 100.0 - idlePercent
		}
	}
	return -1
}

// parseGPUPower extracts GPU power consumption from powermetrics output
func (c *Collector) parseGPUPower(output string) float64 {
	// Look for "GPU Power: XXX mW" or "GPU Power: X.XX W"
	re := regexp.MustCompile(`GPU Power:\s+([\d.]+)\s*(mW|W)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 2 {
		if power, err := strconv.ParseFloat(matches[1], 64); err == nil {
			// Convert mW to W if needed
			if matches[2] == "mW" {
				return power / 1000.0
			}
			return power
		}
	}
	return 0
}

// parseCPUPower extracts CPU power consumption from powermetrics output
func (c *Collector) parseCPUPower(output string) float64 {
	// Look for "Package Power: XXX mW" or similar CPU power indicators
	patterns := []string{
		`Package Power:\s+([\d.]+)\s*(mW|W)`,
		`CPU Power:\s+([\d.]+)\s*(mW|W)`,
		`E-Cluster Average Power:\s+([\d.]+)\s*(mW|W)`,
	}
	
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(output)
		if len(matches) > 2 {
			if power, err := strconv.ParseFloat(matches[1], 64); err == nil {
				// Convert mW to W if needed
				if matches[2] == "mW" {
					return power / 1000.0
				}
				return power
			}
		}
	}
	return 0
}

// parseThermalPressure extracts thermal pressure information from powermetrics output
func (c *Collector) parseThermalPressure(output string) string {
	// Look for thermal pressure indicators
	if strings.Contains(output, "Thermal pressure:") {
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Thermal pressure:") {
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					return strings.TrimSpace(parts[1])
				}
			}
		}
	}
	return ""
}

// HealthCheck returns the health status of the collector
func (c *Collector) HealthCheck() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := map[string]interface{}{
		"status":              "healthy",
		"last_collection":     time.Unix(int64(c.metrics.Timestamp), 0).Format(time.RFC3339),
		"metrics_available":   true,
		"powermetrics_access": c.metrics.GPUPower > 0 || c.metrics.CPUPower > 0,
	}

	// Check if we're getting any meaningful data
	if c.metrics.Timestamp == 0 {
		status["status"] = "degraded"
		status["metrics_available"] = false
	}

	return status
}