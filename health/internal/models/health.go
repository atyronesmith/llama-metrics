package models

// HealthStatus represents the health status of a component
type HealthStatus struct {
	Status         string         `json:"status"`          // healthy, degraded, unhealthy
	Timestamp      string         `json:"timestamp"`
	ResponseTimeMs *float64       `json:"response_time_ms,omitempty"`
	Error          *string        `json:"error,omitempty"`
	Details        map[string]any `json:"details,omitempty"`
}

// ServiceHealth represents individual service health status
type ServiceHealth struct {
	Name     string       `json:"name"`
	URL      string       `json:"url"`
	Status   HealthStatus `json:"status"`
	Critical bool         `json:"critical"`
}

// SystemMetrics represents system resource metrics
type SystemMetrics struct {
	CPU     CPUMetrics     `json:"cpu"`
	Memory  MemoryMetrics  `json:"memory"`
	Disk    DiskMetrics    `json:"disk"`
	Network NetworkMetrics `json:"network"`
	GPU     *GPUMetrics    `json:"gpu,omitempty"`
	Power   *PowerMetrics  `json:"power,omitempty"`
}

// CPUMetrics represents CPU metrics
type CPUMetrics struct {
	Percent float64   `json:"percent"`
	Count   int       `json:"count"`
	LoadAvg []float64 `json:"load_avg,omitempty"`
}

// MemoryMetrics represents memory metrics
type MemoryMetrics struct {
	Percent     float64 `json:"percent"`
	TotalGB     float64 `json:"total_gb"`
	AvailableGB float64 `json:"available_gb"`
	UsedGB      float64 `json:"used_gb"`
}

// DiskMetrics represents disk metrics
type DiskMetrics struct {
	Percent float64 `json:"percent"`
	TotalGB float64 `json:"total_gb"`
	FreeGB  float64 `json:"free_gb"`
	UsedGB  float64 `json:"used_gb"`
}

// NetworkMetrics represents network metrics
type NetworkMetrics struct {
	BytesSent   uint64 `json:"bytes_sent"`
	BytesRecv   uint64 `json:"bytes_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
}

// GPUMetrics represents GPU metrics (macOS)
type GPUMetrics struct {
	Available bool   `json:"available"`
	Data      []any  `json:"data,omitempty"`
}

// PowerMetrics represents power metrics (macOS)
type PowerMetrics struct {
	Available   bool   `json:"available"`
	BatteryInfo string `json:"battery_info,omitempty"`
}

// SystemHealth represents overall system health status
type SystemHealth struct {
	Status        string                 `json:"status"`
	Timestamp     string                 `json:"timestamp"`
	Version       string                 `json:"version"`
	UptimeSeconds float64                `json:"uptime_seconds"`
	Services      []ServiceHealth        `json:"services"`
	SystemMetrics SystemMetrics          `json:"system_metrics"`
	Summary       map[string]interface{} `json:"summary"`
}

// SimpleHealth represents a simple health check response
type SimpleHealth struct {
	Status        string                 `json:"status"`
	Timestamp     string                 `json:"timestamp"`
	Version       string                 `json:"version"`
	UptimeSeconds float64                `json:"uptime_seconds"`
	System        map[string]interface{} `json:"system"`
	Error         string                 `json:"error,omitempty"`
}

// ReadinessStatus represents readiness check response
type ReadinessStatus struct {
	Ready      bool              `json:"ready"`
	Timestamp  string            `json:"timestamp"`
	Components map[string]string `json:"components"`
}

// LivenessStatus represents liveness check response
type LivenessStatus struct {
	Alive         bool    `json:"alive"`
	Timestamp     string  `json:"timestamp"`
	UptimeSeconds float64 `json:"uptime_seconds"`
}