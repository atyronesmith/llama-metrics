package metrics

import (
	"time"

	"github.com/atyronesmith/llama-metrics/proxy/internal/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Collector holds all Prometheus metrics for the proxy
type Collector struct {
	// Request metrics
	RequestCount    *prometheus.CounterVec
	RequestDuration *prometheus.HistogramVec
	HighPriorityRequestDuration *prometheus.HistogramVec
	NormalPriorityRequestDuration *prometheus.HistogramVec
	ActiveRequests  *prometheus.GaugeVec

	// Token metrics
	PromptTokens    *prometheus.CounterVec
	GeneratedTokens *prometheus.CounterVec
	TokensPerSecond *prometheus.HistogramVec

	// Performance metrics
	TimeToFirstToken  *prometheus.HistogramVec
	ModelLoadDuration *prometheus.HistogramVec

	// Error tracking
	ErrorCount *prometheus.CounterVec

	// System metrics
	CPUUsage    prometheus.Gauge
	MemoryUsage prometheus.Gauge
	OllamaServeMemory prometheus.Gauge

	// Queue metrics
	QueueSize            prometheus.Gauge
	QueueProcessingRate  prometheus.Gauge
	QueueWaitTime        *prometheus.HistogramVec
	QueuePeakSize        prometheus.Gauge
	QueueHighPriorityCount    prometheus.Gauge
	QueueNormalPriorityCount  prometheus.Gauge
	QueueHighPriorityWaitTime prometheus.Histogram
	QueueNormalPriorityWaitTime prometheus.Histogram

	// Context length
	ContextLength *prometheus.HistogramVec

	// Mac-specific metrics
	GPUUtilization prometheus.Gauge
	GPUPower       prometheus.Gauge
	CPUPower       prometheus.Gauge
	CPUTemperature prometheus.Gauge
	MemoryPressure prometheus.Gauge
	DiskReadRate   prometheus.Gauge
	DiskWriteRate  prometheus.Gauge
	DiskIOPS       prometheus.Gauge

	// Enhanced AI metrics
	RequestID        *prometheus.CounterVec
	UserRequests     *prometheus.CounterVec
	TokenCost        *prometheus.CounterVec
	RequestSizeByte  *prometheus.HistogramVec
	ResponseSizeByte *prometheus.HistogramVec
}

// NewCollector creates and registers all Prometheus metrics
func NewCollector() *Collector {
	return &Collector{
		RequestCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_proxy_requests_total",
				Help: "Total number of requests",
			},
			[]string{"method", "endpoint", "model", "status"},
		),

		RequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_request_duration_seconds",
				Help:    "Request duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0},
			},
			[]string{"method", "endpoint", "model"},
		),

		HighPriorityRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_high_priority_request_duration_seconds",
				Help:    "High priority request duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0},
			},
			[]string{"method", "endpoint", "model"},
		),

		NormalPriorityRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_normal_priority_request_duration_seconds",
				Help:    "Normal priority request duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 2.5, 5.0, 10.0, 30.0, 60.0, 120.0},
			},
			[]string{"method", "endpoint", "model"},
		),

		ActiveRequests: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_active_requests",
				Help: "Number of active requests",
			},
			[]string{"model"},
		),

		PromptTokens: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_proxy_prompt_tokens_total",
				Help: "Total prompt tokens processed",
			},
			[]string{"model"},
		),

		GeneratedTokens: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_proxy_generated_tokens_total",
				Help: "Total tokens generated",
			},
			[]string{"model"},
		),

		TokensPerSecond: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_tokens_per_second",
				Help:    "Tokens generated per second",
				Buckets: []float64{10, 50, 100, 200, 500, 1000, 2000},
			},
			[]string{"model"},
		),

		TimeToFirstToken: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_time_to_first_token_seconds",
				Help:    "Time to first token in seconds",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
			},
			[]string{"model"},
		),

		ModelLoadDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_model_load_duration_seconds",
				Help:    "Model load duration in seconds",
				Buckets: []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0},
			},
			[]string{"model"},
		),

		ErrorCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_proxy_errors_total",
				Help: "Total number of errors",
			},
			[]string{"model", "error_type"},
		),

		CPUUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_cpu_usage_percent",
				Help: "CPU usage percentage",
			},
		),

		MemoryUsage: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_memory_usage_bytes",
				Help: "Total Ollama processes memory usage in bytes (RSS) - includes all serve and runner processes",
			},
		),

		OllamaServeMemory: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_ollama_serve_memory_bytes",
				Help: "Memory usage of the main Ollama serve process in bytes (RSS)",
			},
		),

		QueueSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_queue_size",
				Help: "Current request queue size",
			},
		),

		QueueProcessingRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_queue_processing_rate",
				Help: "Queue processing rate (requests per second)",
			},
		),

		QueueWaitTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_queue_wait_time_seconds",
				Help:    "Time spent waiting in queue before processing",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
			[]string{"model"},
		),

		QueuePeakSize: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_queue_peak_size",
				Help: "Peak queue size since startup",
			},
		),

		QueueHighPriorityCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_queue_high_priority_count",
				Help: "Current number of high priority requests in the queue",
			},
		),

		QueueNormalPriorityCount: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_queue_normal_priority_count",
				Help: "Current number of normal priority requests in the queue",
			},
		),

		QueueHighPriorityWaitTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_queue_high_priority_wait_time_seconds",
				Help:    "Time spent waiting in high priority queue before processing",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
		),

		QueueNormalPriorityWaitTime: promauto.NewHistogram(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_queue_normal_priority_wait_time_seconds",
				Help:    "Time spent waiting in normal priority queue before processing",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
			},
		),

		ContextLength: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_context_length",
				Help:    "Context length in tokens",
				Buckets: []float64{128, 256, 512, 1024, 2048, 4096, 8192, 16384, 32768},
			},
			[]string{"model"},
		),

		// Mac-specific metrics
		GPUUtilization: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_gpu_active_residency_percent",
				Help: "GPU active residency percentage",
			},
		),

		GPUPower: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_gpu_power_milliwatts",
				Help: "GPU power consumption in milliwatts",
			},
		),

		CPUPower: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_cpu_power_milliwatts",
				Help: "CPU package power consumption in milliwatts",
			},
		),

		CPUTemperature: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_cpu_temperature_celsius",
				Help: "CPU temperature in Celsius",
			},
		),

		MemoryPressure: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_memory_pressure_percent",
				Help: "Memory pressure percentage",
			},
		),

		DiskReadRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_disk_read_bytes_per_second",
				Help: "Disk read rate in bytes per second",
			},
		),

		DiskWriteRate: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_disk_write_bytes_per_second",
				Help: "Disk write rate in bytes per second",
			},
		),

		DiskIOPS: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "ollama_proxy_disk_iops",
				Help: "Disk I/O operations per second",
			},
		),

		// Enhanced AI metrics
		RequestID: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_proxy_request_by_id_total",
				Help: "Total requests by request ID",
			},
			[]string{"request_id", "model", "user"},
		),

		UserRequests: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_proxy_user_requests_total",
				Help: "Total requests by user",
			},
			[]string{"user", "model", "endpoint"},
		),

		TokenCost: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ollama_proxy_token_cost_total",
				Help: "Estimated token cost in cents",
			},
			[]string{"model", "user"},
		),

		RequestSizeByte: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_request_size_bytes",
				Help:    "Request size in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000},
			},
			[]string{"model", "endpoint"},
		),

		ResponseSizeByte: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "ollama_proxy_response_size_bytes",
				Help:    "Response size in bytes",
				Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000},
			},
			[]string{"model", "endpoint"},
		),
	}
}

// RecordRequest records metrics for a request
func (c *Collector) RecordRequest(method, endpoint, model, status string, duration time.Duration) {
	c.RequestCount.WithLabelValues(method, endpoint, model, status).Inc()
	c.RequestDuration.WithLabelValues(method, endpoint, model).Observe(duration.Seconds())
}

// RecordRequestWithPriority records metrics for a request including priority-specific latencies
func (c *Collector) RecordRequestWithPriority(method, endpoint, model, status string, duration time.Duration, priority int) {
	// Record standard metrics
	c.RequestCount.WithLabelValues(method, endpoint, model, status).Inc()
	c.RequestDuration.WithLabelValues(method, endpoint, model).Observe(duration.Seconds())

	// Record priority-specific latencies
	if priority == 1 { // High priority
		c.HighPriorityRequestDuration.WithLabelValues(method, endpoint, model).Observe(duration.Seconds())
	} else { // Normal priority
		c.NormalPriorityRequestDuration.WithLabelValues(method, endpoint, model).Observe(duration.Seconds())
	}
}

// RecordTokens records token metrics from a response
func (c *Collector) RecordTokens(model string, promptTokens, generatedTokens int, tokensPerSec float64) {
	if promptTokens > 0 {
		c.PromptTokens.WithLabelValues(model).Add(float64(promptTokens))
		c.ContextLength.WithLabelValues(model).Observe(float64(promptTokens))
	}

	if generatedTokens > 0 {
		c.GeneratedTokens.WithLabelValues(model).Add(float64(generatedTokens))
	}

	if tokensPerSec > 0 {
		c.TokensPerSecond.WithLabelValues(model).Observe(tokensPerSec)
	}
}

// RecordModelLoadTime records model loading duration
func (c *Collector) RecordModelLoadTime(model string, duration time.Duration) {
	c.ModelLoadDuration.WithLabelValues(model).Observe(duration.Seconds())
}

// RecordTimeToFirstToken records the time to first token
func (c *Collector) RecordTimeToFirstToken(model string, duration time.Duration) {
	c.TimeToFirstToken.WithLabelValues(model).Observe(duration.Seconds())
}

// RecordError increments the error counter
func (c *Collector) RecordError(model, errorType string) {
	c.ErrorCount.WithLabelValues(model, errorType).Inc()
}

// SetActiveRequests sets the number of active requests for a model
func (c *Collector) SetActiveRequests(model string, count float64) {
	c.ActiveRequests.WithLabelValues(model).Set(count)
}

// IncActiveRequests increments the active requests counter
func (c *Collector) IncActiveRequests(model string) {
	c.ActiveRequests.WithLabelValues(model).Inc()
}

// DecActiveRequests decrements the active requests counter
func (c *Collector) DecActiveRequests(model string) {
	c.ActiveRequests.WithLabelValues(model).Dec()
}

// RecordRequestMetadata records enhanced metadata for AI requests
func (c *Collector) RecordRequestMetadata(metadata models.RequestMetadata) {
	// Record request by ID
	c.RequestID.WithLabelValues(metadata.RequestID, metadata.Model, metadata.User).Inc()

	// Record user requests
	if metadata.User != "" {
		c.UserRequests.WithLabelValues(metadata.User, metadata.Model, metadata.Endpoint).Inc()
	}

	// Estimate and record token cost (example pricing)
	costPerToken := c.getTokenCost(metadata.Model)
	totalCost := float64(metadata.TotalTokens) * costPerToken
	if totalCost > 0 && metadata.User != "" {
		c.TokenCost.WithLabelValues(metadata.Model, metadata.User).Add(totalCost)
	}
}

// RecordRequestSize records the size of a request
func (c *Collector) RecordRequestSize(model, endpoint string, sizeBytes int) {
	c.RequestSizeByte.WithLabelValues(model, endpoint).Observe(float64(sizeBytes))
}

// RecordResponseSize records the size of a response
func (c *Collector) RecordResponseSize(model, endpoint string, sizeBytes int) {
	c.ResponseSizeByte.WithLabelValues(model, endpoint).Observe(float64(sizeBytes))
}

// RecordQueueWaitTime records the time a request spent in the queue
func (c *Collector) RecordQueueWaitTime(model string, duration time.Duration) {
	c.QueueWaitTime.WithLabelValues(model).Observe(duration.Seconds())
}

// RecordQueueProcessingRate records the queue processing rate
func (c *Collector) RecordQueueProcessingRate(rate float64) {
	c.QueueProcessingRate.Set(rate)
}

// getTokenCost returns the cost per token in cents for a given model
func (c *Collector) getTokenCost(model string) float64 {
	// Example pricing in cents per 1000 tokens
	// These are example prices - adjust based on your actual costs
	costPer1000Tokens := map[string]float64{
		"llama2:7b":          0.01,   // $0.0001 per 1K tokens
		"llama2:13b":         0.02,   // $0.0002 per 1K tokens
		"llama2:70b":         0.10,   // $0.001 per 1K tokens
		"codellama:7b":       0.01,   // $0.0001 per 1K tokens
		"mistral:7b":         0.01,   // $0.0001 per 1K tokens
		"mixtral:8x7b":       0.05,   // $0.0005 per 1K tokens
		"nomic-embed-text":   0.005,  // $0.00005 per 1K tokens
	}

	if cost, ok := costPer1000Tokens[model]; ok {
		return cost / 1000.0 // Convert to cost per token
	}

	// Default cost for unknown models
	return 0.01 / 1000.0
}