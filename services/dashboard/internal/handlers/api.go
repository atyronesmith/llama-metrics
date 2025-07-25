package handlers

import (
	"log"
	"net/http"
	"time"

	"github.com/atyronesmith/llamastack-prometheus/dashboard/internal/metrics"
	"github.com/gin-gonic/gin"
)

// APIHandler handles API endpoints
type APIHandler struct {
	collector *metrics.Collector
}

// NewAPIHandler creates a new API handler
func NewAPIHandler(collector *metrics.Collector) *APIHandler {
	return &APIHandler{
		collector: collector,
	}
}

// GetMetrics returns all metrics
func (h *APIHandler) GetMetrics(c *gin.Context) {
	summary, err := h.collector.GetSummaryMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	percentiles, err := h.collector.GetLatencyPercentiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary":    summary,
		"percentiles": percentiles,
		"timestamp":  time.Now().Format(time.RFC3339),
	})
}

// GetMetricsSummary returns summary metrics
func (h *APIHandler) GetMetricsSummary(c *gin.Context) {
	summary, err := h.collector.GetSummaryMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	percentiles, err := h.collector.GetLatencyPercentiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	highPriorityPercentiles, err := h.collector.GetHighPriorityLatencyPercentiles()
	if err != nil {
		log.Printf("Error getting high priority percentiles: %v", err)
		highPriorityPercentiles = nil
	}

	c.JSON(http.StatusOK, gin.H{
		"summary":             summary,
		"latency_percentiles": percentiles,
		"high_priority_percentiles": highPriorityPercentiles,
		"timestamp":          time.Now().Format(time.RFC3339),
	})
}

// GetTimeSeriesData returns time series data for graphs
func (h *APIHandler) GetTimeSeriesData(c *gin.Context) {
	hours := 1 // Default to 1 hour

	data, err := h.collector.GetTimeSeriesData(hours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":      data,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// GetAIStatus returns the AI-generated status
func (h *APIHandler) GetAIStatus(c *gin.Context) {
	summary, err := h.collector.GetSummaryMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	percentiles, err := h.collector.GetLatencyPercentiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	status, isAIGenerated := h.collector.GenerateAIStatus(summary, percentiles)

	c.JSON(http.StatusOK, gin.H{
		"status":          status,
		"is_ai_generated": isAIGenerated,
		"timestamp":       time.Now().Format(time.RFC3339),
	})
}

// Health returns the health status of the dashboard
func (h *APIHandler) Health(c *gin.Context) {
	// Simple health check for now
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "dashboard",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}