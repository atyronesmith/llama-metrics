package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/atyronesmith/llama-metrics/proxy/internal/handlers"
	"github.com/atyronesmith/llama-metrics/proxy/internal/metrics"
	"github.com/atyronesmith/llama-metrics/proxy/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	cfg := config.DefaultConfig()
	cfg.LoadFromEnv()
	cfg.LoadFromFlags()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Set Gin mode based on log level
	if cfg.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize metrics
	metricsCollector := metrics.NewCollector()

	// Start system metrics collector
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use standard system collector for all platforms
	systemCollector := metrics.NewSystemCollector(metricsCollector, 10*time.Second)
	systemCollector.Start(ctx)

	// On macOS, also start Mac-specific collector
	if runtime.GOOS == "darwin" {
		macCollector := metrics.NewMacSystemCollector(metricsCollector, 10*time.Second)
		macCollector.Start(ctx)
		log.Println("📱 Mac system metrics collector started")
	}

	// Create handlers
	proxyHandler := handlers.NewProxyHandler(cfg, metricsCollector)
	openAIHandler := handlers.NewOpenAIHandler(cfg, metricsCollector)
	healthHandler := handlers.NewHealthHandler(cfg)

		// Setup proxy router
	proxyRouter := gin.Default()

	// Ollama native API routes
	proxyRouter.POST("/api/generate", proxyHandler.HandleGenerate)
	proxyRouter.POST("/api/chat", proxyHandler.HandleChat)

	// OpenAI-compatible API routes
	proxyRouter.POST("/v1/chat/completions", openAIHandler.HandleChatCompletions)
	proxyRouter.POST("/v1/completions", openAIHandler.HandleCompletions)
	proxyRouter.GET("/v1/models", func(c *gin.Context) {
		// Proxy to Ollama's models endpoint and transform response
		proxyHandler.HandleDefault(c)
	})

	// Default handler for all unmatched routes - this will handle all other paths
	proxyRouter.NoRoute(proxyHandler.HandleDefault)

	// Setup metrics router
	metricsRouter := gin.New()
	metricsRouter.GET("/metrics", gin.WrapH(promhttp.Handler()))
	metricsRouter.GET("/health", healthHandler.Handle)

	// Create servers
	proxySrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.ProxyPort),
		Handler: proxyRouter,
	}

	metricsSrv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
		Handler: metricsRouter,
	}

	// Start servers
	go func() {
		log.Printf("🚀 Ollama Monitoring Proxy Started")
		log.Printf("🔄 Proxy listening on http://localhost:%d", cfg.ProxyPort)
		log.Printf("📊 Metrics available at http://localhost:%d/metrics", cfg.MetricsPort)
		log.Printf("🎯 Forwarding requests to %s", cfg.OllamaURL())
		log.Printf("🖥️  Running on %s/%s", runtime.GOOS, runtime.GOARCH)
		log.Printf("Use proxy URL in your applications for monitoring")

		if err := proxySrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start proxy server: %v", err)
		}
	}()

	go func() {
		if err := metricsSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start metrics server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down servers...")

	// Shutdown servers with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := proxySrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Proxy server forced to shutdown: %v", err)
	}

	if err := metricsSrv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Metrics server forced to shutdown: %v", err)
	}

	log.Println("✅ Servers stopped")
}