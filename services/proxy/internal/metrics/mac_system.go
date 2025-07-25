//go:build darwin
// +build darwin

package metrics

import (
	"bufio"
	"context"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// MacSystemCollector collects Mac-specific system metrics
type MacSystemCollector struct {
	metrics  *Collector
	interval time.Duration
}

// NewMacSystemCollector creates a new Mac system metrics collector
func NewMacSystemCollector(metrics *Collector, interval time.Duration) *MacSystemCollector {
	return &MacSystemCollector{
		metrics:  metrics,
		interval: interval,
	}
}

// Start begins collecting Mac system metrics in the background
func (m *MacSystemCollector) Start(ctx context.Context) {
	go m.collect(ctx)
}

func (m *MacSystemCollector) collect(ctx context.Context) {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// Collect immediately on start
	m.collectOnce()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.collectOnce()
		}
	}
}

func (m *MacSystemCollector) collectOnce() {
	// First try to get metrics from the helper service
	m.fetchMacMetricsFromHelper()

	// Collect GPU metrics using powermetrics (requires sudo)
	m.collectGPUMetrics()

	// Collect temperature using osx-cpu-temp if available
	m.collectTemperature()

	// Collect memory pressure
	m.collectMemoryPressure()

	// Collect disk I/O
	m.collectDiskIO()
}

func (m *MacSystemCollector) collectGPUMetrics() {
	// Try to get GPU metrics using ioreg (doesn't require sudo)
	cmd := exec.Command("ioreg", "-r", "-d", "1", "-w", "0", "-c", "IOAccelerator")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error collecting GPU metrics via ioreg: %v", err)
		return
	}

	// Parse output to find GPU utilization
	// This is a simplified approach - real parsing would be more complex
	outputStr := string(output)
	if strings.Contains(outputStr, "PerformanceStatistics") {
		// Try to extract GPU utilization
		// Note: This is a placeholder - actual parsing would depend on the exact format
		m.metrics.GPUUtilization.Set(0.0) // Default to 0 if we can't parse
	}

	// Alternative: Try using powermetrics if running with appropriate permissions
	m.tryPowerMetrics()
}

func (m *MacSystemCollector) tryPowerMetrics() {
	// This requires sudo permissions, so it might fail
	cmd := exec.Command("sudo", "powermetrics",
		"--samplers", "gpu_power,cpu_power",
		"--sample-count", "1")

	output, err := cmd.Output()
	if err != nil {
		// Log the error instead of silently failing
		log.Printf("Error running powermetrics: %v", err)
		return
	}

	// Parse text output line by line
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()

		// Look for GPU Power line
		if strings.Contains(line, "GPU Power:") {
			// Extract power value: "GPU Power: 7510 mW"
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "Power:" && i+1 < len(parts) {
					if powerStr := strings.TrimSpace(parts[i+1]); powerStr != "" {
						if power, err := strconv.ParseFloat(powerStr, 64); err == nil {
							m.metrics.GPUPower.Set(power)
						}
					}
					break
				}
			}
		}

		// Look for CPU/Package Power line
		if strings.Contains(line, "CPU Power:") || strings.Contains(line, "Package Power:") {
			// Extract power value
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "Power:" && i+1 < len(parts) {
					if powerStr := strings.TrimSpace(parts[i+1]); powerStr != "" {
						if power, err := strconv.ParseFloat(powerStr, 64); err == nil {
							m.metrics.CPUPower.Set(power)
						}
					}
					break
				}
			}
		}

		// Look for GPU active residency to calculate utilization
		if strings.Contains(line, "GPU HW active residency:") {
			// Extract percentage: "GPU HW active residency:  58.06%"
			if idx := strings.Index(line, ":"); idx != -1 {
				percentStr := strings.TrimSpace(line[idx+1:])
				percentStr = strings.TrimSuffix(percentStr, "%")
				// Remove any extra info in parentheses
				if parenIdx := strings.Index(percentStr, "("); parenIdx != -1 {
					percentStr = strings.TrimSpace(percentStr[:parenIdx])
				}
				if util, err := strconv.ParseFloat(percentStr, 64); err == nil {
					m.metrics.GPUUtilization.Set(util)
				}
			}
		}
	}
}

func (m *MacSystemCollector) collectTemperature() {
	// Try using osx-cpu-temp if installed
	cmd := exec.Command("osx-cpu-temp")
	output, err := cmd.Output()
	if err != nil {
		// Try alternative method using powermetrics
		m.collectTemperatureViaPowermetrics()
		return
	}

	// Parse output like "45.5°C"
	tempStr := strings.TrimSpace(string(output))
	tempStr = strings.TrimSuffix(tempStr, "°C")

	if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
		m.metrics.CPUTemperature.Set(temp)
	}
}

func (m *MacSystemCollector) collectTemperatureViaPowermetrics() {
	cmd := exec.Command("sudo", "-n", "powermetrics",
		"--samplers", "smc",
		"--sample-count", "1",
		"--sample-rate", "1000")

	output, err := cmd.Output()
	if err != nil {
		return
	}

	// Parse SMC output for temperature sensors
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "CPU die temperature") {
			// Extract temperature value
			parts := strings.Fields(line)
			for i, part := range parts {
				if strings.Contains(part, "C") && i > 0 {
					if temp, err := strconv.ParseFloat(parts[i-1], 64); err == nil {
						m.metrics.CPUTemperature.Set(temp)
						break
					}
				}
			}
		}
	}
}

func (m *MacSystemCollector) collectMemoryPressure() {
	cmd := exec.Command("memory_pressure")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error collecting memory pressure: %v", err)
		return
	}

	outputStr := string(output)

	// Parse memory pressure output
	if strings.Contains(outputStr, "System-wide memory free percentage:") {
		lines := strings.Split(outputStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "System-wide memory free percentage:") {
				// Extract percentage
				parts := strings.Fields(line)
				if len(parts) > 0 {
					percentStr := strings.TrimSuffix(parts[len(parts)-1], "%")
					if percent, err := strconv.ParseFloat(percentStr, 64); err == nil {
						m.metrics.MemoryPressure.Set(100 - percent) // Convert to used percentage
					}
				}
			}
		}
	}
}

func (m *MacSystemCollector) collectDiskIO() {
	cmd := exec.Command("iostat", "-c", "1")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error collecting disk I/O: %v", err)
		return
	}

	// Parse iostat output
	lines := strings.Split(string(output), "\n")
	if len(lines) > 2 {
		// Skip headers and get the data line
		dataLine := lines[len(lines)-2]
		fields := strings.Fields(dataLine)

		if len(fields) >= 3 {
			// KB/t (kilobytes per transfer)
			if kbt, err := strconv.ParseFloat(fields[0], 64); err == nil {
				m.metrics.DiskReadRate.Set(kbt * 1024) // Convert to bytes
			}

			// tps (transfers per second)
			if tps, err := strconv.ParseFloat(fields[1], 64); err == nil {
				m.metrics.DiskIOPS.Set(tps)
			}

			// MB/s
			if mbs, err := strconv.ParseFloat(fields[2], 64); err == nil {
				m.metrics.DiskWriteRate.Set(mbs * 1024 * 1024) // Convert to bytes/sec
			}
		}
	}
}