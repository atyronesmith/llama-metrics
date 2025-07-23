package handlers

import (
	"fmt"
	"net/http"

	"github.com/atyronesmith/llama-metrics/proxy/pkg/config"
	"github.com/gin-gonic/gin"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	config *config.Config
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(cfg *config.Config) *HealthHandler {
	return &HealthHandler{
		config: cfg,
	}
}

// Handle returns the health status
func (h *HealthHandler) Handle(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":         "healthy",
		"proxy_url":      fmt.Sprintf("http://localhost:%d", h.config.ProxyPort),
		"metrics_url":    fmt.Sprintf("http://localhost:%d/metrics", h.config.MetricsPort),
		"ollama_backend": h.config.OllamaURL(),
	})
}