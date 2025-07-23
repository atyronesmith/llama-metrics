package metrics

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"
)

// SystemCollector collects system metrics periodically
type SystemCollector struct {
	metrics  *Collector
	interval time.Duration
}

// NewSystemCollector creates a new system metrics collector
func NewSystemCollector(metrics *Collector, interval time.Duration) *SystemCollector {
	return &SystemCollector{
		metrics:  metrics,
		interval: interval,
	}
}

// Start begins collecting system metrics in the background
func (s *SystemCollector) Start(ctx context.Context) {
	go s.collect(ctx)
}

func (s *SystemCollector) collect(ctx context.Context) {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	// Collect immediately on start
	s.collectOnce()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.collectOnce()
		}
	}
}

func (s *SystemCollector) collectOnce() {
	// Collect CPU usage
	cpuPercent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Printf("Error collecting CPU metrics: %v", err)
	} else if len(cpuPercent) > 0 {
		s.metrics.CPUUsage.Set(cpuPercent[0])
	}

	// Collect Ollama process memory usage
	s.collectOllamaMemory()
}

// collectOllamaMemory finds and monitors the Ollama process memory usage
func (s *SystemCollector) collectOllamaMemory() {
	// Get all processes
	processes, err := process.Processes()
	if err != nil {
		log.Printf("Error getting processes: %v", err)
		return
	}

	// Sum memory usage across all Ollama processes
	var totalMemory uint64 = 0
	var serveMemory uint64 = 0
	foundOllama := false
	foundServe := false

	for _, p := range processes {
		name, err := p.Name()
		if err != nil {
			continue
		}

		// Also check command line args for better detection
		cmdline, _ := p.Cmdline()

		// Check if this is an Ollama process (main serve or runner)
		if strings.Contains(strings.ToLower(name), "ollama") ||
		   strings.Contains(strings.ToLower(cmdline), "ollama") {
			// Get memory info
			memInfo, err := p.MemoryInfo()
			if err != nil {
				continue
			}

			// Add to total memory (RSS - Resident Set Size)
			totalMemory += memInfo.RSS
			foundOllama = true

			// Check if this is the main serve process
			if strings.Contains(cmdline, "serve") && !strings.Contains(cmdline, "runner") {
				serveMemory = memInfo.RSS
				foundServe = true
			}
		}
	}

	// Set the total memory usage metric
	if foundOllama {
		s.metrics.MemoryUsage.Set(float64(totalMemory))
		log.Printf("Ollama total memory usage: %.2f MB", float64(totalMemory)/(1024*1024))
	} else {
		// If Ollama process not found, set to 0
		s.metrics.MemoryUsage.Set(0)
		log.Printf("Ollama process not found")
	}

	// Set the serve process memory metric
	if foundServe {
		s.metrics.OllamaServeMemory.Set(float64(serveMemory))
		log.Printf("Ollama serve process memory: %.2f MB", float64(serveMemory)/(1024*1024))
	} else {
		s.metrics.OllamaServeMemory.Set(0)
	}
}