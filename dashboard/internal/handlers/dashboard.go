package handlers

import (
	"net/http"

	"github.com/atyronesmith/llamastack-prometheus/dashboard/internal/metrics"
	"github.com/atyronesmith/llamastack-prometheus/dashboard/internal/websocket"
	"github.com/gin-gonic/gin"
)

// DashboardHandler handles the main dashboard page
type DashboardHandler struct {
	collector *metrics.Collector
	hub       *websocket.Hub
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(collector *metrics.Collector, hub *websocket.Hub) *DashboardHandler {
	return &DashboardHandler{
		collector: collector,
		hub:       hub,
	}
}

// Index renders the main dashboard page
func (h *DashboardHandler) Index(c *gin.Context) {
	c.HTML(http.StatusOK, "dashboard.html", gin.H{
		"title": "Ollama LLM Dashboard",
	})
}