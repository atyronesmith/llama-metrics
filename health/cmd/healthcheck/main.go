package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"encoding/json"

	"github.com/atyronesmith/llama-metrics/health/internal/checker"
	"github.com/atyronesmith/llama-metrics/health/internal/models"
	"github.com/atyronesmith/llama-metrics/health/pkg/config"
	"github.com/gin-gonic/gin"
)

var (
	configPath = flag.String("config", "", "Path to config.yml file")
	port       = flag.Int("port", 8080, "Port to listen on")
	mode       = flag.String("mode", "server", "Mode: server or cli")
	checkType  = flag.String("check", "comprehensive", "Check type for CLI mode: comprehensive, simple, readiness, liveness, analyzed")
)

func main() {
	flag.Parse()

	// Set version from environment
	if version := os.Getenv("VERSION"); version == "" {
		os.Setenv("VERSION", "1.0.0")
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create health checker
	healthChecker := checker.NewHealthChecker(cfg)

	if *mode == "cli" {
		// CLI mode - run check and exit
		runCLICheck(healthChecker, *checkType)
		return
	}

	// Server mode - start HTTP server
	runServer(healthChecker, *port)
}

func runCLICheck(hc *checker.HealthChecker, checkType string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	switch checkType {
	case "comprehensive":
		health := hc.GetComprehensiveHealth(ctx)
		printJSON(health)
	case "simple":
		health := hc.GetSimpleHealth()
		printJSON(health)
	case "readiness":
		status := hc.GetReadinessStatus()
		printJSON(status)
		if !status.Ready {
			os.Exit(1)
		}
	case "liveness":
		status := hc.GetLivenessStatus()
		printJSON(status)
		if !status.Alive {
			os.Exit(1)
		}
	case "analyzed":
		fmt.Println("\033[0;34m🔍 Running comprehensive health check with LLM analysis...\033[0m")
		analyzed := hc.GetAnalyzedHealth(ctx)

		// Print a formatted summary instead of raw JSON
		printAnalyzedHealth(analyzed)
	default:
		log.Fatalf("Unknown check type: %s", checkType)
	}
}

func printJSON(v interface{}) {
	// Pretty print JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(v)
}

func printAnalyzedHealth(analyzed models.AnalyzedHealth) {
	// Color codes
	const (
		colorReset  = "\033[0m"
		colorRed    = "\033[0;31m"
		colorGreen  = "\033[0;32m"
		colorYellow = "\033[1;33m"
		colorBlue   = "\033[0;34m"
		colorPurple = "\033[0;35m"
		colorCyan   = "\033[0;36m"
		colorBold   = "\033[1m"
	)

	// Print overall status
	fmt.Println()
	statusColor := colorGreen
	statusIcon := "✅"
	if analyzed.Status == "unhealthy" {
		statusColor = colorRed
		statusIcon = "❌"
	} else if analyzed.Status == "degraded" {
		statusColor = colorYellow
		statusIcon = "⚠️"
	}

	fmt.Printf("%s%s Overall Status: %s%s%s\n", colorBold, statusIcon, statusColor, strings.ToUpper(analyzed.Status), colorReset)
	fmt.Printf("%sUptime: %.1f hours%s\n\n", colorCyan, analyzed.UptimeSeconds/3600, colorReset)

	// Print services summary
	fmt.Printf("%s📊 Services Summary:%s\n", colorBlue, colorReset)
	for _, service := range analyzed.Services {
		icon := "✅"
		color := colorGreen
		if service.Status.Status != "healthy" {
			icon = "❌"
			color = colorRed
		}

		responseTime := ""
		if service.Status.ResponseTimeMs != nil {
			responseTime = fmt.Sprintf(" (%dms)", int(*service.Status.ResponseTimeMs))
		}

		fmt.Printf("  %s %s%-12s%s %s%s%s%s\n",
			icon,
			colorCyan,
			service.Name+":",
			colorReset,
			color,
			service.Status.Status,
			colorReset,
			responseTime)

		if service.Status.Error != nil {
			fmt.Printf("     %s└─ Error: %s%s\n", colorRed, *service.Status.Error, colorReset)
		}
	}

	// Print system metrics summary
	fmt.Printf("\n%s💻 System Resources:%s\n", colorBlue, colorReset)

	cpuColor := colorGreen
	if analyzed.SystemMetrics.CPU.Percent > 80 {
		cpuColor = colorRed
	} else if analyzed.SystemMetrics.CPU.Percent > 60 {
		cpuColor = colorYellow
	}
	fmt.Printf("  CPU:    %s%.1f%%%s", cpuColor, analyzed.SystemMetrics.CPU.Percent, colorReset)
	if len(analyzed.SystemMetrics.CPU.LoadAvg) >= 3 {
		fmt.Printf(" (Load: %.2f, %.2f, %.2f)",
			analyzed.SystemMetrics.CPU.LoadAvg[0],
			analyzed.SystemMetrics.CPU.LoadAvg[1],
			analyzed.SystemMetrics.CPU.LoadAvg[2])
	}
	fmt.Println()

	memColor := colorGreen
	if analyzed.SystemMetrics.Memory.Percent > 85 {
		memColor = colorRed
	} else if analyzed.SystemMetrics.Memory.Percent > 70 {
		memColor = colorYellow
	}
	fmt.Printf("  Memory: %s%.1f%%%s (%.1f/%.1f GB)\n",
		memColor,
		analyzed.SystemMetrics.Memory.Percent,
		colorReset,
		analyzed.SystemMetrics.Memory.UsedGB,
		analyzed.SystemMetrics.Memory.TotalGB)

	diskColor := colorGreen
	if analyzed.SystemMetrics.Disk.Percent > 80 {
		diskColor = colorRed
	} else if analyzed.SystemMetrics.Disk.Percent > 60 {
		diskColor = colorYellow
	}
	fmt.Printf("  Disk:   %s%.1f%%%s (%.1f/%.1f GB)\n",
		diskColor,
		analyzed.SystemMetrics.Disk.Percent,
		colorReset,
		analyzed.SystemMetrics.Disk.UsedGB,
		analyzed.SystemMetrics.Disk.TotalGB)

	// Print LLM analysis if available
	if analyzed.Analysis != nil && analyzed.Analysis.Available {
		fmt.Printf("\n%s🤖 AI Health Analysis:%s\n", colorPurple, colorReset)
		fmt.Println(strings.Repeat("─", 60))

		// Format the analysis text with proper line wrapping
		lines := strings.Split(analyzed.Analysis.Summary, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				fmt.Println()
				continue
			}

			// Highlight section headers
			if strings.Contains(line, "Overall System Health") ||
			   strings.Contains(line, "Issues") ||
			   strings.Contains(line, "Recommendations") ||
			   strings.Contains(line, "Performance Optimization") {
				fmt.Printf("\n%s%s%s%s\n", colorBold, colorCyan, line, colorReset)
			} else {
				// Word wrap long lines
				wrapAndPrint(line, 80)
			}
		}
		fmt.Println()
	} else if analyzed.Analysis != nil && !analyzed.Analysis.Available {
		fmt.Printf("\n%s⚠️  AI Analysis Unavailable: %s%s\n", colorYellow, analyzed.Analysis.Error, colorReset)
	}
}

func wrapAndPrint(text string, width int) {
	words := strings.Fields(text)
	if len(words) == 0 {
		return
	}

	line := words[0]
	for _, word := range words[1:] {
		if len(line)+1+len(word) > width {
			fmt.Println(line)
			line = word
		} else {
			line += " " + word
		}
	}
	if line != "" {
		fmt.Println(line)
	}
}

func runServer(hc *checker.HealthChecker, port int) {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// Health check endpoints
	router.GET("/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		health := hc.GetComprehensiveHealth(ctx)
		c.JSON(http.StatusOK, health)
	})

	router.GET("/health/simple", func(c *gin.Context) {
		health := hc.GetSimpleHealth()
		c.JSON(http.StatusOK, health)
	})

	router.GET("/readiness", func(c *gin.Context) {
		status := hc.GetReadinessStatus()
		statusCode := http.StatusOK
		if !status.Ready {
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, status)
	})

	router.GET("/liveness", func(c *gin.Context) {
		status := hc.GetLivenessStatus()
		statusCode := http.StatusOK
		if !status.Alive {
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, status)
	})

	// Analyzed health endpoint with LLM insights
	router.GET("/health/analyzed", func(c *gin.Context) {
		ctx := c.Request.Context()
		analyzed := hc.GetAnalyzedHealth(ctx)
		c.JSON(http.StatusOK, analyzed)
	})

	// Legacy endpoints for compatibility
	router.GET("/api/health", func(c *gin.Context) {
		ctx := c.Request.Context()
		health := hc.GetComprehensiveHealth(ctx)
		c.JSON(http.StatusOK, health)
	})

	// Start server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server forced to shutdown: %v", err)
		}
	}()

	log.Printf("Health check server listening on port %d", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}