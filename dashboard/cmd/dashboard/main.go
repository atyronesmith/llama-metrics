package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atyronesmith/llamastack-prometheus/dashboard/internal/handlers"
	"github.com/atyronesmith/llamastack-prometheus/dashboard/internal/metrics"
	"github.com/atyronesmith/llamastack-prometheus/dashboard/internal/websocket"
	"github.com/atyronesmith/llamastack-prometheus/dashboard/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Set Gin mode based on environment
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Prometheus client
	client, err := api.NewClient(api.Config{
		Address: cfg.PrometheusURL,
	})
	if err != nil {
		log.Fatalf("Error creating Prometheus client: %v", err)
	}
	promAPI := v1.NewAPI(client)

	// Create metrics collector
	metricsCollector := metrics.NewCollector(promAPI, cfg.OllamaURL)

	// Create WebSocket hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// Start background metrics broadcaster
	go startMetricsBroadcaster(metricsCollector, wsHub)

	// Create router
	router := gin.Default()

	// Load HTML templates
	router.LoadHTMLGlob("web/templates/*")

	// Static files
	router.Static("/static", "./web/static")

	// Create handlers
	dashboardHandler := handlers.NewDashboardHandler(metricsCollector, wsHub)
	apiHandler := handlers.NewAPIHandler(metricsCollector)
	wsHandler := handlers.NewWebSocketHandler(wsHub)

	// Routes
	router.GET("/", dashboardHandler.Index)
	router.GET("/ws", wsHandler.HandleWebSocket)

	// API endpoints
	api := router.Group("/api")
	{
		api.GET("/metrics", apiHandler.GetMetrics)
		api.GET("/metrics/summary", apiHandler.GetMetricsSummary)
		api.GET("/metrics/timeseries", apiHandler.GetTimeSeriesData)
		api.GET("/status", apiHandler.GetAIStatus)
		api.GET("/health", apiHandler.Health)
	}

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Printf("🚀 Ollama LLM Dashboard started on http://localhost:%d", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

// startMetricsBroadcaster broadcasts metrics updates to all connected clients
func startMetricsBroadcaster(collector *metrics.Collector, hub *websocket.Hub) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get latest metrics
			summary, err := collector.GetSummaryMetrics()
			if err != nil {
				log.Printf("Error getting summary metrics: %v", err)
				continue
			}

			percentiles, err := collector.GetLatencyPercentiles()
			if err != nil {
				log.Printf("Error getting latency percentiles: %v", err)
				continue
			}

			// Get AI status
			aiStatus, isAIGenerated := collector.GenerateAIStatus(summary, percentiles)

			// Prepare broadcast data
			data := map[string]interface{}{
				"summary":            summary,
				"latency_percentiles": percentiles,
				"timestamp":          time.Now().Format(time.RFC3339),
				"ai_status":          aiStatus,
				"is_ai_generated":    isAIGenerated,
			}

			// Broadcast to all connected clients
			hub.Broadcast(data)
		}
	}
}