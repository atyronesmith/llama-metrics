package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/atyronesmith/llama-metrics/mac-metrics/internal/collector"
	"github.com/atyronesmith/llama-metrics/mac-metrics/internal/server"
)

func main() {
	// Check if running on macOS
	if runtime.GOOS != "darwin" {
		log.Fatal("This service is only supported on macOS (Darwin)")
	}

	// Parse command line flags
	var (
		port = flag.Int("port", 8002, "HTTP server port")
		help = flag.Bool("help", false, "Show help information")
	)
	flag.Parse()

	if *help {
		showHelp()
		return
	}

	log.Println("Starting Mac System Metrics Service")
	log.Printf("Platform: %s/%s", runtime.GOOS, runtime.GOARCH)

	// Check sudo access for powermetrics
	checkSudoAccess()

	// Create metrics collector
	collector := collector.NewCollector()

	// Start metrics collection
	collector.Start()
	defer collector.Stop()

	// Create and start HTTP server
	srv := server.NewServer(*port, collector)

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Received shutdown signal")

		// Create a context with timeout for graceful shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Stop the collector
		collector.Stop()

		// Stop the HTTP server
		if err := srv.Stop(ctx); err != nil {
			log.Printf("Error stopping server: %v", err)
		}

		os.Exit(0)
	}()

	// Start the server (blocking)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func showHelp() {
	log.Println("Mac System Metrics Service")
	log.Println("==========================")
	log.Println()
	log.Println("This service collects macOS system metrics including:")
	log.Println("  - GPU utilization and power consumption")
	log.Println("  - CPU power consumption and temperature") 
	log.Println("  - Memory pressure")
	log.Println("  - Thermal pressure state")
	log.Println()
	log.Println("Usage:")
	log.Println("  mac-metrics [options]")
	log.Println()
	log.Println("Options:")
	flag.PrintDefaults()
	log.Println()
	log.Println("Endpoints:")
	log.Println("  GET  /            - Service information")
	log.Println("  GET  /health      - Health check")
	log.Println("  GET  /metrics     - Prometheus metrics")
	log.Println("  GET  /metrics/json - JSON metrics (legacy)")
	log.Println()
	log.Println("Requirements:")
	log.Println("  - macOS (Darwin) operating system")
	log.Println("  - sudo access for powermetrics (optional, for GPU/CPU power)")
	log.Println("  - memory_pressure command (built into macOS)")
	log.Println()
	log.Println("For full metrics including GPU/CPU power, run with sudo:")
	log.Println("  sudo mac-metrics")
}

func checkSudoAccess() {
	// Check if we can run sudo -n (non-interactive)
	// This doesn't actually run powermetrics but checks sudo access
	log.Println("Checking system access...")
	
	if os.Getuid() == 0 {
		log.Println("✅ Running as root - full metrics available")
		return
	}

	log.Println("⚠️  Running as regular user")
	log.Println("   - Memory pressure: Available")
	log.Println("   - GPU/CPU power: Requires sudo access")
	log.Println("   - CPU temperature: Limited (requires osx-cpu-temp or sudo)")
	log.Println()
	log.Println("For full metrics, run: sudo mac-metrics")
}