package metrics

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// Collector handles metrics collection from Prometheus and AI status generation
type Collector struct {
	promAPI    v1.API
	ollamaURL  string
	httpClient *http.Client

	// Request history for local rate calculation
	requestHistory []requestDataPoint
	historyMutex   sync.RWMutex

	// AI status generation state
	lastStatus          string
	lastGenerationTime  time.Time
	requestInProgress   bool
	consecutiveTimeouts int
	statusMutex         sync.RWMutex
}

type requestDataPoint struct {
	timestamp    time.Time
	totalRequests float64
}

// NewCollector creates a new metrics collector
func NewCollector(promAPI v1.API, ollamaURL string) *Collector {
	return &Collector{
		promAPI:    promAPI,
		ollamaURL:  ollamaURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		lastStatus: "System operational",
	}
}

// toMetricValue converts a float64 to interface{}, converting NaN/Inf to nil
func toMetricValue(val float64) interface{} {
	if math.IsNaN(val) || math.IsInf(val, 0) {
		return nil
	}
	return val
}

// GetSummaryMetrics retrieves summary metrics from Prometheus
func (c *Collector) GetSummaryMetrics() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	metrics := make(map[string]interface{})

	// Request rate
	requestRate, err := c.calculateRequestRate(ctx)
	if err != nil {
		log.Printf("Error calculating request rate: %v", err)
	}
	metrics["request_rate"] = toMetricValue(requestRate)

	// Average latency
	avgLatency, err := c.queryScalar(ctx, `sum(rate(ollama_proxy_request_duration_seconds_sum{endpoint="/api/generate"}[5m])) / sum(rate(ollama_proxy_request_duration_seconds_count{endpoint="/api/generate"}[5m]))`)
	if err != nil {
		log.Printf("Error querying average latency: %v", err)
	}
	metrics["avg_latency"] = toMetricValue(avgLatency)

	// Success rate
	successRate, err := c.calculateSuccessRate(ctx)
	if err != nil {
		log.Printf("Error calculating success rate: %v", err)
	}
	metrics["success_rate"] = toMetricValue(successRate)

	// Token generation rate
	tokenRate, err := c.queryScalar(ctx, `rate(ollama_proxy_generated_tokens_total[5m])`)
	if err != nil {
		log.Printf("Error querying token rate: %v", err)
	}
	metrics["tokens_per_second"] = toMetricValue(tokenRate)

	// GPU utilization
	gpuUtil, err := c.queryScalar(ctx, `ollama_proxy_gpu_active_residency_percent`)
	if err != nil {
		log.Printf("Error querying GPU utilization: %v", err)
	}
	metrics["gpu_utilization"] = toMetricValue(gpuUtil)

	// Power consumption (convert from milliwatts to watts)
	powerMilliwatts, err := c.queryScalar(ctx, `ollama_proxy_cpu_power_milliwatts`)
	if err != nil {
		log.Printf("Error querying power consumption: %v", err)
	}
	metrics["power_consumption"] = toMetricValue(powerMilliwatts / 1000.0)

	// Memory usage - track just the main Ollama serve process, not all runners
	memoryBytes, err := c.queryScalar(ctx, `ollama_proxy_ollama_serve_memory_bytes`)
	if err != nil {
		log.Printf("Error querying memory: %v", err)
		memoryBytes = 0.0
	}
	metrics["memory_usage"] = toMetricValue(memoryBytes / (1024 * 1024)) // Convert to MB

	// Active requests
	activeReqs, err := c.queryScalar(ctx, `sum(ollama_proxy_active_requests)`)
	if err != nil {
		log.Printf("Error querying active requests: %v", err)
	}
	metrics["active_requests"] = int(activeReqs)

	// Queue metrics
	queueSize, err := c.queryScalar(ctx, `ollama_proxy_queue_size`)
	if err == nil {
		metrics["queue_size"] = int(queueSize)
	}

	queueRate, err := c.queryScalar(ctx, `ollama_proxy_queue_processing_rate`)
	if err == nil {
		metrics["queue_processing_rate"] = queueRate
	}

	maxQueueSize, err := c.queryScalar(ctx, `ollama_proxy_queue_peak_size`)
	if err == nil {
		metrics["max_queue_size"] = int(maxQueueSize)
	}

	// Check Ollama health
	metrics["ollama_status"] = c.checkOllamaHealth()

	// Check Proxy health
	metrics["proxy_status"] = c.checkProxyHealth()

	// Direct requests count
	totalRequests, err := c.queryScalar(ctx, `ollama_proxy_requests_total`)
	if err != nil {
		log.Printf("Error querying total requests: %v", err)
	}
	metrics["direct_requests"] = int(totalRequests)
	metrics["routing_ratio"] = 0 // No routing in this setup

	return metrics, nil
}

// GetLatencyPercentiles retrieves latency percentiles from Prometheus
func (c *Collector) GetLatencyPercentiles() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	percentiles := make(map[string]interface{})
	quantiles := []int{50, 75, 95, 99}

	for _, p := range quantiles {
		quantile := float64(p) / 100.0
		query := fmt.Sprintf(`histogram_quantile(%f, rate(ollama_proxy_request_duration_seconds_bucket[5m]))`, quantile)

		value, err := c.queryScalar(ctx, query)
		if err != nil {
			log.Printf("Error querying p%d: %v", p, err)
			percentiles[fmt.Sprintf("p%d", p)] = nil
		} else {
			percentiles[fmt.Sprintf("p%d", p)] = toMetricValue(value)
		}
	}

	return percentiles, nil
}

// GetTimeSeriesData retrieves time series data for charts
func (c *Collector) GetTimeSeriesData(hours int) (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(hours) * time.Hour)

	data := make(map[string]interface{})

	// Token generation rate
	tokensData, err := c.queryRange(ctx, `rate(ollama_proxy_generated_tokens_total[1m])`, startTime, endTime)
	if err != nil {
		log.Printf("Error querying tokens time series: %v", err)
	} else {
		data["tokens_per_second"] = tokensData
	}

	// Memory usage
	memoryData, err := c.queryRange(ctx, `ollama_proxy_memory_usage_bytes / 1024 / 1024`, startTime, endTime)
	if err != nil {
		log.Printf("Error querying memory time series: %v", err)
	} else {
		data["memory_usage"] = memoryData
	}

	// GPU utilization
	gpuData, err := c.queryRange(ctx, `ollama_proxy_gpu_active_residency_percent`, startTime, endTime)
	if err != nil {
		log.Printf("Error querying GPU time series: %v", err)
	} else {
		data["gpu_utilization"] = gpuData
	}

	// Power consumption (convert from milliwatts to watts)
	powerData, err := c.queryRange(ctx, `ollama_proxy_cpu_power_milliwatts / 1000`, startTime, endTime)
	if err != nil {
		log.Printf("Error querying power time series: %v", err)
	} else {
		data["power_consumption"] = powerData
	}

	// Queue metrics
	queueSizeData, err := c.queryRange(ctx, `ollama_proxy_queue_size`, startTime, endTime)
	if err == nil {
		data["queue_size"] = queueSizeData
	}

	queueRateData, err := c.queryRange(ctx, `ollama_proxy_queue_processing_rate`, startTime, endTime)
	if err == nil {
		data["queue_processing_rate"] = queueRateData
	}

	return data, nil
}

// GenerateAIStatus generates a human-readable status using the LLM
func (c *Collector) GenerateAIStatus(summary map[string]interface{}, percentiles map[string]interface{}) (string, bool) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	// Check if we should skip generation
	activeRequests := getInt(summary, "active_requests")
	queueSize := getInt(summary, "queue_size")

	if activeRequests > 5 || queueSize > 10 {
		// System under load
		tokensPerSec := getFloat(summary, "tokens_per_second")
		avgLatency := getFloat(summary, "avg_latency")
		status := fmt.Sprintf("High load: %d active requests, %d queued. %.1f tokens/s, %.2fs avg latency",
			activeRequests, queueSize, tokensPerSec, avgLatency)
		return status, false
	}

	// Check if we're already generating
	if c.requestInProgress {
		return c.lastStatus, true
	}

	// Only generate every 15 seconds
	if time.Since(c.lastGenerationTime) < 15*time.Second {
		return c.lastStatus, true
	}

	// If too many timeouts, wait longer
	if c.consecutiveTimeouts >= 3 && time.Since(c.lastGenerationTime) < 60*time.Second {
		return fmt.Sprintf("⚠️ LLM temporarily unavailable - %s", c.lastStatus), false
	}

	// Mark as in progress
	c.requestInProgress = true
	c.lastGenerationTime = time.Now()

	// Prepare context
	context := c.prepareMetricsContext(summary)

	// Create prompt
	prompt := c.createStatusPrompt(context)

	// Query LLM
	response, err := c.queryLLM(prompt)
	c.requestInProgress = false

	if err != nil {
		c.consecutiveTimeouts++
		log.Printf("LLM query error: %v", err)
		return c.lastStatus, false
	}

	if response != "" {
		c.lastStatus = response
		c.consecutiveTimeouts = 0
		return response, true
	}

	// Fallback status
	c.lastStatus = c.generateFallbackStatus(summary)
	return c.lastStatus, false
}

// Helper functions

func (c *Collector) calculateRequestRate(ctx context.Context) (float64, error) {
	// Get current total requests
	totalRequests, err := c.queryScalar(ctx, `ollama_proxy_requests_total`)
	if err != nil {
		return 0.0, err
	}

	// Update request history
	c.updateRequestHistory(totalRequests)

	// Calculate local rate
	localRate := c.calculateLocalRequestRate()
	if localRate > 0 {
		return localRate, nil
	}

	// Try Prometheus rate
	rate, err := c.queryScalar(ctx, `rate(ollama_proxy_requests_total[2m])`)
	if err != nil {
		return 0.0, err
	}

	return rate, nil
}

func (c *Collector) updateRequestHistory(totalRequests float64) {
	c.historyMutex.Lock()
	defer c.historyMutex.Unlock()

	c.requestHistory = append(c.requestHistory, requestDataPoint{
		timestamp:    time.Now(),
		totalRequests: totalRequests,
	})

	// Keep only last 20 data points
	if len(c.requestHistory) > 20 {
		c.requestHistory = c.requestHistory[len(c.requestHistory)-20:]
	}
}

func (c *Collector) calculateLocalRequestRate() float64 {
	c.historyMutex.RLock()
	defer c.historyMutex.RUnlock()

	if len(c.requestHistory) < 2 {
		return 0.0
	}

	oldest := c.requestHistory[0]
	newest := c.requestHistory[len(c.requestHistory)-1]

	timeDiff := newest.timestamp.Sub(oldest.timestamp).Seconds()
	if timeDiff <= 0 {
		return 0.0
	}

	requestDiff := newest.totalRequests - oldest.totalRequests
	return requestDiff / timeDiff
}

func (c *Collector) calculateSuccessRate(ctx context.Context) (float64, error) {
	successRate, err := c.queryScalar(ctx, `rate(ollama_proxy_requests_total{status="200"}[5m])`)
	if err != nil {
		return 0.0, err
	}

	totalRate, err := c.queryScalar(ctx, `rate(ollama_proxy_requests_total[5m])`)
	if err != nil {
		return 0.0, err
	}

	if totalRate > 0 {
		return (successRate / totalRate) * 100, nil
	}

	return 0.0, nil
}

func (c *Collector) queryScalar(ctx context.Context, query string) (float64, error) {
	result, _, err := c.promAPI.Query(ctx, query, time.Now())
	if err != nil {
		return 0.0, err
	}

	switch v := result.(type) {
	case model.Vector:
		if len(v) > 0 {
			val := float64(v[0].Value)
			// Return the raw value, including NaN
			return val, nil
		}
	}

	return 0.0, nil
}

func (c *Collector) queryRange(ctx context.Context, query string, start, end time.Time) ([]map[string]interface{}, error) {
	r := v1.Range{
		Start: start,
		End:   end,
		Step:  30 * time.Second,
	}

	result, _, err := c.promAPI.QueryRange(ctx, query, r)
	if err != nil {
		return nil, err
	}

	var data []map[string]interface{}

	switch v := result.(type) {
	case model.Matrix:
		if len(v) > 0 {
			for _, pair := range v[0].Values {
				data = append(data, map[string]interface{}{
					"x": pair.Timestamp.Unix() * 1000, // Convert to milliseconds
					"y": float64(pair.Value),
				})
			}
		}
	}

	return data, nil
}

func (c *Collector) checkOllamaHealth() map[string]interface{} {
	status := map[string]interface{}{
		"status":        "unknown",
		"response_time": nil,
		"last_check":    time.Now().Unix(),
	}

	start := time.Now()
	resp, err := c.httpClient.Get(c.ollamaURL + "/api/tags")
	if err != nil {
		status["status"] = "offline"
		return status
	}
	defer resp.Body.Close()

	responseTime := time.Since(start).Milliseconds()

	if resp.StatusCode == 200 {
		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			if models, ok := data["models"].([]interface{}); ok && len(models) > 0 {
				status["status"] = "healthy"
			} else {
				status["status"] = "unhealthy"
			}
			status["response_time"] = responseTime
		}
	} else {
		status["status"] = "unhealthy"
		status["response_time"] = responseTime
	}

	return status
}

func (c *Collector) checkProxyHealth() map[string]interface{} {
	status := map[string]interface{}{
		"status":        "unknown",
		"response_time": nil,
		"last_check":    time.Now().Unix(),
	}

	// Proxy health endpoint is on metrics port 8001
	start := time.Now()
	resp, err := c.httpClient.Get("http://localhost:8001/health")
	if err != nil {
		status["status"] = "offline"
		return status
	}
	defer resp.Body.Close()

	responseTime := time.Since(start).Milliseconds()

	if resp.StatusCode == 200 {
		// For now, just check if we get a 200 response
		status["status"] = "healthy"
		status["response_time"] = responseTime

		// Try to parse the response if it's JSON
		var data map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
			// If there's a status field in the response, use it
			if s, ok := data["status"].(string); ok {
				status["status"] = s
			}
		}
	} else {
		status["status"] = "unhealthy"
		status["response_time"] = responseTime
	}

	return status
}

func (c *Collector) prepareMetricsContext(summary map[string]interface{}) map[string]string {
	context := make(map[string]string)

	// Analyze request activity
	rate := getFloat(summary, "request_rate")
	if rate > 2.0 {
		context["request_activity"] = "very high activity"
	} else if rate > 1.0 {
		context["request_activity"] = "high activity"
	} else if rate > 0.2 {
		context["request_activity"] = "moderate activity"
	} else if rate > 0 {
		context["request_activity"] = "low activity"
	} else {
		context["request_activity"] = "idle"
	}

	// Analyze latency
	latency := getFloat(summary, "avg_latency")
	if latency > 5.0 {
		context["latency_status"] = "very high latency"
	} else if latency > 2.0 {
		context["latency_status"] = "elevated latency"
	} else if latency > 0.5 {
		context["latency_status"] = "normal latency"
	} else {
		context["latency_status"] = "excellent latency"
	}

	// GPU status
	gpu := getFloat(summary, "gpu_utilization")
	if gpu > 80 {
		context["gpu_status"] = "high GPU usage"
	} else if gpu > 50 {
		context["gpu_status"] = "moderate GPU usage"
	} else if gpu > 10 {
		context["gpu_status"] = "light GPU usage"
	} else {
		context["gpu_status"] = "minimal GPU usage"
	}

	// Other metrics
	context["power_status"] = fmt.Sprintf("%.1fW power consumption", getFloat(summary, "power_consumption"))
	context["memory_status"] = fmt.Sprintf("%.0fMB memory used", getFloat(summary, "memory_usage"))
	context["success_status"] = fmt.Sprintf("%.1f%% success rate", getFloat(summary, "success_rate"))
	context["token_generation"] = fmt.Sprintf("%.1f tokens/second", getFloat(summary, "tokens_per_second"))
	context["active_requests"] = fmt.Sprintf("%d", getInt(summary, "active_requests"))

	return context
}

func (c *Collector) createStatusPrompt(context map[string]string) string {
	return fmt.Sprintf(`Generate a brief status summary for an AI server monitoring dashboard. Use the metrics below to create one paragraph (2-3 sentences).

Current metrics:
- Request Activity: %s
- Latency: %s
- GPU: %s
- Power: %s
- Memory: %s
- Reliability: %s
- Token Generation: %s

Write a status summary:`,
		context["request_activity"],
		context["latency_status"],
		context["gpu_status"],
		context["power_status"],
		context["memory_status"],
		context["success_status"],
		context["token_generation"])
}

func (c *Collector) queryLLM(prompt string) (string, error) {
	payload := map[string]interface{}{
		"model":  "phi3:mini",
		"prompt": prompt,
		"stream": false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.ollamaURL+"/api/generate", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Priority", "high")  // AI summaries get high priority

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("LLM returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	// Validate response
	response = strings.TrimSpace(response)
	if response == "" {
		return "", fmt.Errorf("empty response")
	}

	// Check for error indicators
	errorIndicators := []string{"sorry", "I need", "dictionary", "python", "document", "instruction"}
	lowerResponse := strings.ToLower(response)
	for _, indicator := range errorIndicators {
		if strings.Contains(lowerResponse, indicator) {
			return "", fmt.Errorf("invalid LLM response")
		}
	}

	// Limit length
	if len(response) > 500 {
		response = response[:497] + "..."
	}

	return response, nil
}

func (c *Collector) generateFallbackStatus(summary map[string]interface{}) string {
	return fmt.Sprintf("System operational: %d active requests, %.1f tokens/s, %.2fs latency, GPU %.0f%%",
		getInt(summary, "active_requests"),
		getFloat(summary, "tokens_per_second"),
		getFloat(summary, "avg_latency"),
		getFloat(summary, "gpu_utilization"))
}

// Utility functions
func getFloat(m map[string]interface{}, key string) float64 {
	if v, ok := m[key].(float64); ok {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return 0.0
		}
		return v
	}
	return 0.0
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key].(int); ok {
		return v
	}
	if v, ok := m[key].(float64); ok {
		return int(v)
	}
	return 0
}