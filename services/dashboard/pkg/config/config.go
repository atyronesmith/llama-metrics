package config

import (
	"os"
	"strconv"
	"time"

	sharedconfig "github.com/llama-metrics/shared/config"
)

// Config holds the configuration for the dashboard
type Config struct {
	sharedconfig.BaseConfig
	sharedconfig.PrometheusConfig
	sharedconfig.OllamaConfig

	// Dashboard-specific fields
	Environment   string
	PrometheusURL string
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	base := sharedconfig.DefaultBaseConfig("llama-dashboard")
	base.Port = 3001 // Dashboard default port

	cfg := &Config{
		BaseConfig: base,
		PrometheusConfig: sharedconfig.PrometheusConfig{
			MetricsPath:      "/metrics",
			MetricsNamespace: "llama_metrics",
		},
		OllamaConfig: sharedconfig.OllamaConfig{
			URL:     "http://localhost:11434",
			Timeout: 30 * time.Second,
		},
		Environment:   "development",
		PrometheusURL: "http://localhost:9090",
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
		cfg.OllamaConfig.URL = ollamaURL
	}

	// Load shared config from environment
	cfg.MetricsPort = sharedconfig.LoadEnvInt("METRICS_PORT", cfg.MetricsPort)
	cfg.LogLevel = sharedconfig.LoadEnvString("LOG_LEVEL", cfg.LogLevel)

	return cfg
}