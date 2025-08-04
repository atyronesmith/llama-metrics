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
	
	// Smart caching for AI status
	lastMetricsSnapshot map[string]interface{}
	lastCacheTime       time.Time
	cacheThresholds     map[string]float64
	
	// Context memory for enhanced summarization
	contextHistory      []ContextEntry
	conversationContext []ChatMessage
}

type requestDataPoint struct {
	timestamp    time.Time
	totalRequests float64
}

// ContextEntry represents a previous summarization with timestamp
type ContextEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Metrics   string    `json:"metrics"`
	Summary   string    `json:"summary"`
}

// ChatMessage represents a message in the conversation context
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// NewCollector creates a new metrics collector
func NewCollector(promAPI v1.API, ollamaURL string) *Collector {
	return &Collector{
		promAPI:    promAPI,
		ollamaURL:  ollamaURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		lastStatus: "System operational",
		lastMetricsSnapshot: make(map[string]interface{}),
		cacheThresholds: map[string]float64{
			"request_rate":       0.5,  // Request rate change of 0.5 req/s
			"avg_latency":        0.5,  // Latency change of 0.5 seconds
			"tokens_per_second":  5.0,  // Token rate change of 5 tokens/s
			"gpu_utilization":    10.0, // GPU usage change of 10%
			"power_consumption":  5.0,  // Power change of 5W
			"memory_usage":       100.0, // Memory change of 100MB
			"success_rate":       5.0,  // Success rate change of 5%
			"active_requests":    3.0,  // Active requests change of 3
			"queue_size":         2.0,  // Queue size change of 2
		},
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

	// GPU utilization from Mac metrics service
	gpuUtil, err := c.queryScalar(ctx, `mac_gpu_utilization_percent`)
	if err != nil {
		log.Printf("Error querying GPU utilization: %v", err)
	}
	metrics["gpu_utilization"] = toMetricValue(gpuUtil)

	// CPU Power consumption from Mac metrics service (already in watts)
	cpuPower, err := c.queryScalar(ctx, `mac_cpu_power_watts`)
	if err != nil {
		log.Printf("Error querying CPU power consumption: %v", err)
	}
	metrics["power_consumption"] = toMetricValue(cpuPower)

	// GPU Power consumption from Mac metrics service
	gpuPower, err := c.queryScalar(ctx, `mac_gpu_power_watts`)
	if err != nil {
		log.Printf("Error querying GPU power consumption: %v", err)
	}
	metrics["gpu_power"] = toMetricValue(gpuPower)

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

	// Priority queue item counters
	highPriorityItemsTotal, err := c.queryScalar(ctx, `ollama_proxy_queue_high_priority_items_total`)
	if err == nil {
		metrics["high_priority_items_total"] = int(highPriorityItemsTotal)
	}

	normalPriorityItemsTotal, err := c.queryScalar(ctx, `ollama_proxy_queue_normal_priority_items_total`)
	if err == nil {
		metrics["normal_priority_items_total"] = int(normalPriorityItemsTotal)
	}

	// Priority queue success counters
	highPrioritySuccessTotal, err := c.queryScalar(ctx, `ollama_proxy_queue_high_priority_success_total`)
	if err == nil {
		metrics["high_priority_success_total"] = int(highPrioritySuccessTotal)
	}

	normalPrioritySuccessTotal, err := c.queryScalar(ctx, `ollama_proxy_queue_normal_priority_success_total`)
	if err == nil {
		metrics["normal_priority_success_total"] = int(normalPrioritySuccessTotal)
	}

	// Get uptime data from mac-metrics service
	uptimeData := c.getUptimeData()

	// Check Ollama health
	metrics["ollama_status"] = c.checkOllamaHealth(uptimeData)

	// Check Proxy health
	metrics["proxy_status"] = c.checkProxyHealth(uptimeData)

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

// GetHighPriorityLatencyPercentiles retrieves latency percentiles for high priority requests
func (c *Collector) GetHighPriorityLatencyPercentiles() (map[string]interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	percentiles := make(map[string]interface{})
	quantiles := []int{50, 75, 95, 99}

	for _, p := range quantiles {
		quantile := float64(p) / 100.0
		query := fmt.Sprintf(`histogram_quantile(%f, rate(ollama_proxy_high_priority_request_duration_seconds_bucket[5m]))`, quantile)

		value, err := c.queryScalar(ctx, query)
		if err != nil {
			log.Printf("Error querying high priority p%d: %v", p, err)
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

	// GPU utilization from Mac metrics service
	gpuData, err := c.queryRange(ctx, `mac_gpu_utilization_percent`, startTime, endTime)
	if err != nil {
		log.Printf("Error querying GPU time series: %v", err)
	} else {
		data["gpu_utilization"] = gpuData
	}

	// Power consumption from Mac metrics service (already in watts)
	powerData, err := c.queryRange(ctx, `mac_cpu_power_watts`, startTime, endTime)
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

// hasSignificantChange checks if any metric has changed significantly since last cache
func (c *Collector) hasSignificantChange(currentMetrics map[string]interface{}) bool {
	if len(c.lastMetricsSnapshot) == 0 {
		return true // First time - always generate
	}
	
	for key, threshold := range c.cacheThresholds {
		currentVal := getFloat(currentMetrics, key)
		lastVal := getFloat(c.lastMetricsSnapshot, key)
		
		// Special handling for active_requests and queue_size (treat as integers)
		if key == "active_requests" || key == "queue_size" {
			currentInt := getInt(currentMetrics, key)
			lastInt := getInt(c.lastMetricsSnapshot, key)
			if math.Abs(float64(currentInt-lastInt)) >= threshold {
				log.Printf("Significant change detected in %s: %d -> %d (threshold: %.1f)", 
					key, lastInt, currentInt, threshold)
				return true
			}
		} else {
			// Handle float metrics
			if math.Abs(currentVal-lastVal) >= threshold {
				log.Printf("Significant change detected in %s: %.2f -> %.2f (threshold: %.1f)", 
					key, lastVal, currentVal, threshold)
				return true
			}
		}
	}
	
	return false
}

// shouldUpdateCache determines if AI status should be updated based on time and changes
func (c *Collector) shouldUpdateCache(currentMetrics map[string]interface{}) bool {
	now := time.Now()
	
	// Always update if more than 1 minute has passed
	if now.Sub(c.lastCacheTime) >= 60*time.Second {
		log.Printf("Cache expired: %v since last update", now.Sub(c.lastCacheTime))
		return true
	}
	
	// Update if significant changes detected
	if c.hasSignificantChange(currentMetrics) {
		return true
	}
	
	return false
}

// updateMetricsSnapshot stores current metrics for future comparison
func (c *Collector) updateMetricsSnapshot(metrics map[string]interface{}) {
	// Create a copy of key metrics for comparison
	c.lastMetricsSnapshot = make(map[string]interface{})
	for key := range c.cacheThresholds {
		if val, exists := metrics[key]; exists {
			c.lastMetricsSnapshot[key] = val
		}
	}
	c.lastCacheTime = time.Now()
}

// GenerateAIStatus generates a human-readable status using the LLM
func (c *Collector) GenerateAIStatus(summary map[string]interface{}, percentiles map[string]interface{}) (string, bool) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()

	// Check if we should skip generation due to high load
	activeRequests := getInt(summary, "active_requests")
	queueSize := getInt(summary, "queue_size")

	if activeRequests > 5 || queueSize > 10 {
		// System under load - use quick fallback and update snapshot
		tokensPerSec := getFloat(summary, "tokens_per_second")
		avgLatency := getFloat(summary, "avg_latency")
		status := fmt.Sprintf("High load: %d active requests, %d queued. %.1f tokens/s, %.2fs avg latency",
			activeRequests, queueSize, tokensPerSec, avgLatency)
		c.updateMetricsSnapshot(summary)
		return status, false
	}

	// Check if we're already generating
	if c.requestInProgress {
		return c.lastStatus, true
	}

	// Smart caching: only generate if significant changes or cache expired
	if !c.shouldUpdateCache(summary) {
		return c.lastStatus, true
	}

	// If consecutive failures, skip next attempts and use fallback
	if c.consecutiveTimeouts >= 2 {
		waitTime := time.Duration(c.consecutiveTimeouts*30) * time.Second // Progressive backoff
		if time.Since(c.lastGenerationTime) < waitTime {
			return c.generateFallbackStatus(summary), false
		}
	}

	// Mark as in progress
	c.requestInProgress = true
	c.lastGenerationTime = time.Now()

	// Prepare metrics summary
	metricsContext := c.prepareMetricsContext(summary)
	
	// Create metrics summary for context
	currentMetrics := c.createMetricsSummary(metricsContext)

	// Query LLM with conversation context
	response, err := c.queryLLMWithContext(currentMetrics)
	c.requestInProgress = false

	if err != nil {
		c.consecutiveTimeouts++
		log.Printf("LLM query error: %v", err)
		return c.lastStatus, false
	}

	if response != "" {
		c.lastStatus = response
		c.consecutiveTimeouts = 0
		
		// Store context for future summarizations
		c.storeContext(currentMetrics, response)
		
		// Update metrics snapshot since we successfully generated new status
		c.updateMetricsSnapshot(summary)
		
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

// getUptimeData fetches uptime data from the mac-metrics service
func (c *Collector) getUptimeData() map[string]interface{} {
	resp, err := c.httpClient.Get("http://localhost:8002/uptime")
	if err != nil {
		log.Printf("Error fetching uptime data: %v", err)
		return make(map[string]interface{})
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("Uptime service returned status %d", resp.StatusCode)
		return make(map[string]interface{})
	}

	var uptimeData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&uptimeData); err != nil {
		log.Printf("Error parsing uptime data: %v", err)
		return make(map[string]interface{})
	}

	return uptimeData
}

func (c *Collector) checkOllamaHealth(uptimeData map[string]interface{}) map[string]interface{} {
	status := map[string]interface{}{
		"status":        "unknown",
		"response_time": nil,
		"last_check":    time.Now().Unix(),
		"uptime":        nil,
	}

	// Add uptime data from mac-metrics service
	if ollamaUptime, ok := uptimeData["ollama"]; ok {
		if uptimeMap, ok := ollamaUptime.(map[string]interface{}); ok {
			if uptime, ok := uptimeMap["uptime"].(string); ok {
				status["uptime"] = uptime
			}
		}
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

func (c *Collector) checkProxyHealth(uptimeData map[string]interface{}) map[string]interface{} {
	status := map[string]interface{}{
		"status":        "unknown",
		"response_time": nil,
		"last_check":    time.Now().Unix(),
		"uptime":        nil,
	}

	// Add uptime data from mac-metrics service
	if proxyUptime, ok := uptimeData["proxy"]; ok {
		if uptimeMap, ok := proxyUptime.(map[string]interface{}); ok {
			if uptime, ok := uptimeMap["uptime"].(string); ok {
				status["uptime"] = uptime
			}
		}
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

	// Other metrics with timestamp
	now := time.Now()
	context["timestamp"] = fmt.Sprintf("%d", now.Unix()) // Unix timestamp
	context["human_time"] = now.Format("01-02-06 3:04pm") // Short format MM-DD-YY like "08-04-25 9:46pm"
	context["power_status"] = fmt.Sprintf("%.1fW power consumption", getFloat(summary, "power_consumption"))
	context["memory_status"] = fmt.Sprintf("%.0fMB memory used", getFloat(summary, "memory_usage"))
	context["success_status"] = fmt.Sprintf("%.1f%% success rate", getFloat(summary, "success_rate"))
	context["token_generation"] = fmt.Sprintf("%.1f tokens/second", getFloat(summary, "tokens_per_second"))
	context["active_requests"] = fmt.Sprintf("%d", getInt(summary, "active_requests"))
	
	// Queue priority metrics
	highPriorityTotal := getInt(summary, "high_priority_items_total")
	normalPriorityTotal := getInt(summary, "normal_priority_items_total")
	highPrioritySuccess := getInt(summary, "high_priority_success_total")
	normalPrioritySuccess := getInt(summary, "normal_priority_success_total")
	
	if highPriorityTotal > 0 || normalPriorityTotal > 0 {
		context["queue_priority_status"] = fmt.Sprintf("HP: %d/%d success, NP: %d/%d success", 
			highPrioritySuccess, highPriorityTotal, normalPrioritySuccess, normalPriorityTotal)
	} else {
		context["queue_priority_status"] = "no queue activity"
	}

	return context
}

func (c *Collector) createStatusPrompt(context map[string]string) string {
	return fmt.Sprintf(`System status: %s requests, %s, GPU %s. Status:`,
		context["active_requests"],
		context["latency_status"],
		context["gpu_status"])
}

// createMetricsSummary creates a concise metrics summary for context
func (c *Collector) createMetricsSummary(context map[string]string) string {
	return fmt.Sprintf("Time: %s - Active: %s requests, %s, GPU: %s, %s, Memory: %s, Tokens: %s, Queue: %s",
		context["human_time"],
		context["active_requests"],
		context["latency_status"], 
		context["gpu_status"],
		context["power_status"],
		context["memory_status"],
		context["token_generation"],
		context["queue_priority_status"])
}

// storeContext stores the current metrics and response in context history
func (c *Collector) storeContext(metrics string, response string) {
	entry := ContextEntry{
		Timestamp: time.Now(),
		Metrics:   metrics,
		Summary:   response,
	}
	
	// Add to context history
	c.contextHistory = append(c.contextHistory, entry)
	
	// Keep only last 2 entries
	if len(c.contextHistory) > 2 {
		c.contextHistory = c.contextHistory[len(c.contextHistory)-2:]
	}
	
	// Update conversation context
	c.updateConversationContext(metrics, response)
}

// updateConversationContext maintains the chat conversation context
func (c *Collector) updateConversationContext(metrics string, response string) {
	// Add user message (current metrics)
	userMessage := ChatMessage{
		Role:    "user",
		Content: fmt.Sprintf("New metrics: %s. Update system status summary. Use the provided unix timestamp, not placeholder text.", metrics),
	}
	
	// Add assistant response
	assistantMessage := ChatMessage{
		Role:    "assistant", 
		Content: response,
	}
	
	c.conversationContext = append(c.conversationContext, userMessage, assistantMessage)
	
	// Keep only last 6 messages (3 exchanges)
	if len(c.conversationContext) > 6 {
		c.conversationContext = c.conversationContext[len(c.conversationContext)-6:]
	}
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:11435/api/generate", bytes.NewBuffer(jsonData))
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

// queryLLMWithContext queries the LLM using chat API with conversation context
func (c *Collector) queryLLMWithContext(currentMetrics string) (string, error) {
	// Build conversation messages
	messages := make([]ChatMessage, 0)
	
	// Add existing conversation context
	messages = append(messages, c.conversationContext...)
	
	// Add current metrics as new user message
	if len(c.conversationContext) == 0 {
		// First time - add system message
		systemMessage := ChatMessage{
			Role:    "system",
			Content: "You are a system monitoring assistant. Provide concise, technical status updates based on metrics. Keep responses under 100 words and focus on key insights and trends. Start your response with 'System Status Summary (' followed by the time from the metrics in format like '08-04-25 9:46pm', then '):'. Don't show long timestamp numbers, don't say 'Timestamp:'.",
		}
		messages = append(messages, systemMessage)
	}
	
	userMessage := ChatMessage{
		Role:    "user",
		Content: fmt.Sprintf("New metrics: %s. Update system status summary. Use the provided unix timestamp, not placeholder text.", currentMetrics),
	}
	messages = append(messages, userMessage)
	
	// Create chat payload
	payload := map[string]interface{}{
		"model":    "phi3:mini",
		"messages": messages,
		"stream":   false,
		"options": map[string]interface{}{
			"temperature": 0.3,
			"num_predict": 150,
		},
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// Use chat API endpoint instead of generate
	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:11435/api/chat", bytes.NewBuffer(jsonData))
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
	
	// Extract response from chat API format
	var response string
	if message, ok := result["message"].(map[string]interface{}); ok {
		if content, ok := message["content"].(string); ok {
			response = content
		}
	}
	
	if response == "" {
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
	// Get service uptimes for more informative status
	ollamaUptime := ""
	if status, ok := summary["ollama_status"].(map[string]interface{}); ok {
		if uptime, ok := status["uptime"].(string); ok {
			ollamaUptime = fmt.Sprintf(", Ollama: %s", uptime)
		}
	}
	
	proxyUptime := ""
	if status, ok := summary["proxy_status"].(map[string]interface{}); ok {
		if uptime, ok := status["uptime"].(string); ok {
			proxyUptime = fmt.Sprintf(", Proxy: %s", uptime)
		}
	}

	return fmt.Sprintf("System operational - %d active requests, %.1f tokens/s, GPU %.0f%%%s%s",
		getInt(summary, "active_requests"),
		getFloat(summary, "tokens_per_second"),
		getFloat(summary, "gpu_utilization"),
		ollamaUptime,
		proxyUptime)
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