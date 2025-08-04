package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/atyronesmith/llama-metrics/mac-metrics/internal/collector"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server represents the HTTP server for Mac metrics
type Server struct {
	collector *collector.Collector
	server    *http.Server
	port      int
}

// NewServer creates a new HTTP server
func NewServer(port int, collector *collector.Collector) *Server {
	return &Server{
		collector: collector,
		port:      port,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", s.healthHandler).Methods("GET")

	// JSON metrics endpoint (for legacy compatibility)
	router.HandleFunc("/metrics/json", s.jsonMetricsHandler).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", promhttp.Handler()).Methods("GET")

	// Root endpoint with service info
	router.HandleFunc("/", s.rootHandler).Methods("GET")

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting Mac metrics server on port %d", s.port)
	log.Printf("Endpoints:")
	log.Printf("  - http://localhost:%d/health", s.port)
	log.Printf("  - http://localhost:%d/metrics", s.port)
	log.Printf("  - http://localhost:%d/metrics/json", s.port)

	return s.server.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	log.Println("Shutting down Mac metrics server...")
	return s.server.Shutdown(ctx)
}

// healthHandler handles health check requests
func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	health := s.collector.HealthCheck()
	w.Header().Set("Content-Type", "application/json")

	if health["status"] == "healthy" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(health)
}

// jsonMetricsHandler handles JSON metrics requests (legacy compatibility)
func (s *Server) jsonMetricsHandler(w http.ResponseWriter, r *http.Request) {
	metrics := s.collector.GetMetrics()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(metrics)
}

// rootHandler provides service information
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	info := map[string]interface{}{
		"service":     "Mac System Metrics",
		"version":     "1.0.0",
		"description": "Go-based Mac system metrics collector with powermetrics support",
		"endpoints": map[string]string{
			"/health":       "Health check and status information",
			"/metrics":      "Prometheus metrics endpoint",
			"/metrics/json": "JSON metrics endpoint (legacy compatibility)",
		},
		"capabilities": map[string]interface{}{
			"gpu_utilization":  "GPU usage percentage",
			"gpu_power":        "GPU power consumption",
			"cpu_power":        "CPU power consumption",
			"cpu_temperature":  "CPU temperature",
			"memory_pressure":  "Memory pressure percentage",
			"thermal_pressure": "Thermal state monitoring",
		},
		"requirements": map[string]interface{}{
			"platform":      "macOS (Darwin)",
			"powermetrics":  "sudo access for full GPU/CPU power metrics",
			"memory_pressure": "Built-in macOS command",
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(info)
}

// Port returns the server port
func (s *Server) Port() int {
	return s.port
}