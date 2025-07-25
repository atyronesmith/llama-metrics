//go:build darwin
// +build darwin

package metrics

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"
)

// MacMetricsResponse represents the response from mac_metrics_helper.py
type MacMetricsResponse struct {
	GPUUtilization  float64 `json:"gpu_utilization"`
	GPUPower        float64 `json:"gpu_power"`
	CPUPower        float64 `json:"cpu_power"`
	CPUTemperature  float64 `json:"cpu_temperature"`
	MemoryPressure  float64 `json:"memory_pressure"`
	ThermalPressure string  `json:"thermal_pressure"`
	Timestamp       float64 `json:"timestamp"`
}

// fetchMacMetricsFromHelper fetches metrics from the Python helper service
func (m *MacSystemCollector) fetchMacMetricsFromHelper() {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	resp, err := client.Get("http://localhost:8002/metrics")
	if err != nil {
		// Helper not running, this is OK
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading helper response: %v", err)
		return
	}

	var metrics MacMetricsResponse
	if err := json.Unmarshal(body, &metrics); err != nil {
		log.Printf("Error parsing helper metrics: %v", err)
		return
	}

	// Update Prometheus metrics
	if metrics.GPUUtilization > 0 {
		m.metrics.GPUUtilization.Set(metrics.GPUUtilization)
	}

	if metrics.GPUPower > 0 {
		m.metrics.GPUPower.Set(metrics.GPUPower)
	}

	if metrics.CPUPower > 0 {
		m.metrics.CPUPower.Set(metrics.CPUPower)
	}

	if metrics.CPUTemperature > 0 {
		m.metrics.CPUTemperature.Set(metrics.CPUTemperature)
	}

	if metrics.MemoryPressure > 0 {
		m.metrics.MemoryPressure.Set(metrics.MemoryPressure)
	}

	// Set thermal pressure as a label metric
	thermalValue := 0.0
	switch metrics.ThermalPressure {
	case "nominal":
		thermalValue = 0.0
	case "fair":
		thermalValue = 0.33
	case "serious":
		thermalValue = 0.66
	case "critical":
		thermalValue = 1.0
	}

	// For simplicity, we can use memory pressure gauge to represent thermal pressure
	// In a real implementation, you'd create a separate metric
	_ = thermalValue // Placeholder
}