// Package metrics provides shared Prometheus metrics definitions
// for all services in the llama-metrics project.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Request metrics
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "llama_metrics",
			Subsystem: "requests",
			Name:      "total",
			Help:      "Total number of requests processed",
		},
		[]string{"service", "method", "status", "endpoint"},
	)

	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "llama_metrics",
			Subsystem: "requests",
			Name:      "duration_seconds",
			Help:      "Request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "method", "endpoint"},
	)

	// Active requests gauge
	ActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "requests",
			Name:      "active",
			Help:      "Number of active requests",
		},
		[]string{"service", "endpoint"},
	)

	// Response size histogram
	ResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "llama_metrics",
			Subsystem: "responses",
			Name:      "size_bytes",
			Help:      "Response size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 8), // 100B to 10MB
		},
		[]string{"service", "endpoint"},
	)

	// Error counter
	ErrorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "llama_metrics",
			Subsystem: "errors",
			Name:      "total",
			Help:      "Total number of errors",
		},
		[]string{"service", "type", "endpoint"},
	)

	// Service info gauge (for version tracking)
	ServiceInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "service",
			Name:      "info",
			Help:      "Service information",
		},
		[]string{"service", "version", "build_date", "commit"},
	)
)

// Ollama-specific metrics
var (
	// Token metrics
	TokensGenerated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "llama_metrics",
			Subsystem: "ollama",
			Name:      "tokens_generated_total",
			Help:      "Total number of tokens generated",
		},
		[]string{"model", "endpoint"},
	)

	TokensPerSecond = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "llama_metrics",
			Subsystem: "ollama",
			Name:      "tokens_per_second",
			Help:      "Token generation rate",
			Buckets:   prometheus.ExponentialBuckets(1, 2, 10), // 1 to 512 tokens/sec
		},
		[]string{"model"},
	)

	// Model loading time
	ModelLoadDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "llama_metrics",
			Subsystem: "ollama",
			Name:      "model_load_duration_seconds",
			Help:      "Time taken to load a model",
			Buckets:   prometheus.ExponentialBuckets(0.1, 2, 10), // 0.1s to 51.2s
		},
		[]string{"model"},
	)

	// Queue metrics
	QueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "queue",
			Name:      "size",
			Help:      "Current queue size",
		},
		[]string{"service", "queue_name"},
	)

	QueueWaitTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "llama_metrics",
			Subsystem: "queue",
			Name:      "wait_time_seconds",
			Help:      "Time spent waiting in queue",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"service", "queue_name"},
	)
)

// System metrics (for services that collect system info)
var (
	// CPU metrics
	CPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "system",
			Name:      "cpu_usage_percent",
			Help:      "CPU usage percentage",
		},
		[]string{"service", "cpu"},
	)

	// Memory metrics
	MemoryUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "system",
			Name:      "memory_usage_bytes",
			Help:      "Memory usage in bytes",
		},
		[]string{"service", "type"}, // type: rss, vms, available, used
	)

	// GPU metrics (for systems with GPU)
	GPUUsage = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "system",
			Name:      "gpu_usage_percent",
			Help:      "GPU usage percentage",
		},
		[]string{"service", "gpu", "type"}, // type: compute, memory
	)

	GPUTemperature = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "system",
			Name:      "gpu_temperature_celsius",
			Help:      "GPU temperature in Celsius",
		},
		[]string{"service", "gpu"},
	)

	GPUPower = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "llama_metrics",
			Subsystem: "system",
			Name:      "gpu_power_watts",
			Help:      "GPU power consumption in watts",
		},
		[]string{"service", "gpu"},
	)
)

// RegisterServiceInfo registers the service information metric
func RegisterServiceInfo(serviceName, version, buildDate, commit string) {
	ServiceInfo.WithLabelValues(serviceName, version, buildDate, commit).Set(1)
}