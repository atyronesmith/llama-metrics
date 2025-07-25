package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	sharedconfig "github.com/llama-metrics/shared/config"
	"gopkg.in/yaml.v3"
)

// Config represents the complete configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Models     ModelConfig      `yaml:"models"`
	Monitoring MonitoringConfig `yaml:"monitoring"`

	// Embedded shared configs for consistency
	BaseConfig       sharedconfig.BaseConfig       `yaml:"-"`
	PrometheusConfig sharedconfig.PrometheusConfig `yaml:"-"`
	OllamaConfig     sharedconfig.OllamaConfig     `yaml:"-"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	OllamaURL      string `yaml:"ollama_url"`
	ProxyPort      int    `yaml:"proxy_port"`
	ProxyHost      string `yaml:"proxy_host"`
	MetricsPort    int    `yaml:"metrics_port"`
	MetricsHost    string `yaml:"metrics_host"`
	DashboardPort  int    `yaml:"dashboard_port"`
	DashboardHost  string `yaml:"dashboard_host"`
	PrometheusPort int    `yaml:"prometheus_port"`
	PrometheusHost string `yaml:"prometheus_host"`
}

// ModelConfig represents model configuration
type ModelConfig struct {
	DefaultModel     string   `yaml:"default_model"`
	AvailableModels []string `yaml:"available_models"`
}

// MonitoringConfig represents monitoring configuration
type MonitoringConfig struct {
	MetricsInterval       int `yaml:"metrics_interval"`
	RequestTimeout        int `yaml:"request_timeout"`
	MaxConcurrentRequests int `yaml:"max_concurrent_requests"`
	MaxQueueSize          int `yaml:"max_queue_size"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// If no path provided, look for config.yml in current directory
	if configPath == "" {
		configPath = "config.yml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Try to find it relative to the executable
		execPath, _ := os.Executable()
		execDir := filepath.Dir(execPath)
		configPath = filepath.Join(execDir, "config.yml")

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			// Try parent directory (when running from health/ subdirectory)
			configPath = filepath.Join("..", "config.yml")
		}
	}

	// Read the file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Set defaults if not specified
	if config.Server.OllamaURL == "" {
		config.Server.OllamaURL = "http://localhost:11434"
	}
	if config.Server.ProxyHost == "" {
		config.Server.ProxyHost = "localhost"
	}
	if config.Server.MetricsHost == "" {
		config.Server.MetricsHost = "localhost"
	}
	if config.Server.DashboardHost == "" {
		config.Server.DashboardHost = "localhost"
	}
	if config.Models.DefaultModel == "" {
		config.Models.DefaultModel = "phi3:mini"
	}

	// Initialize shared configs
	config.BaseConfig = sharedconfig.DefaultBaseConfig("llama-health")
	config.BaseConfig.Port = 8080 // Health check service port
	config.BaseConfig.MetricsPort = config.Server.MetricsPort

	config.PrometheusConfig = sharedconfig.PrometheusConfig{
		MetricsPath:      "/metrics",
		MetricsNamespace: "llama_metrics",
	}

	config.OllamaConfig = sharedconfig.OllamaConfig{
		URL:     config.Server.OllamaURL,
		Timeout: time.Duration(config.Monitoring.RequestTimeout) * time.Second,
	}

	return &config, nil
}