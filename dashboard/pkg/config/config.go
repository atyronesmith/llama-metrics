package config

import (
	"os"
	"strconv"
)

// Config holds the configuration for the dashboard
type Config struct {
	Port          int
	Environment   string
	PrometheusURL string
	OllamaURL     string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	cfg := &Config{
		Port:          3001,
		Environment:   "development",
		PrometheusURL: "http://localhost:9090",
		OllamaURL:     "http://localhost:11434",
	}

	// Override with environment variables if set
	if port := os.Getenv("DASHBOARD_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if env := os.Getenv("DASHBOARD_ENV"); env != "" {
		cfg.Environment = env
	}

	if promURL := os.Getenv("PROMETHEUS_URL"); promURL != "" {
		cfg.PrometheusURL = promURL
	}

	if ollamaURL := os.Getenv("OLLAMA_URL"); ollamaURL != "" {
		cfg.OllamaURL = ollamaURL
	}

	return cfg
}