package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	sharedconfig "github.com/llama-metrics/shared/config"
)

// Config holds the proxy configuration
type Config struct {
	sharedconfig.BaseConfig
	sharedconfig.OllamaConfig

	// Proxy-specific fields
	OllamaHost     string
	OllamaPort     int
	ProxyPort      int
	MaxQueueSize   int
	MaxConcurrency int
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	base := sharedconfig.DefaultBaseConfig("ollama-proxy")

	return &Config{
		BaseConfig: base,
		OllamaConfig: sharedconfig.OllamaConfig{
			URL:     "http://localhost:11434",
			Timeout: 30 * time.Second,
		},
		OllamaHost:     "localhost",
		OllamaPort:     11434,
		ProxyPort:      11435,
		MaxQueueSize:   100,
		MaxConcurrency: 4,  // Reduced to prevent Ollama overload
	}
}

// LoadFromFlags loads configuration from command-line flags
func (c *Config) LoadFromFlags() {
	flag.StringVar(&c.OllamaHost, "ollama-host", c.OllamaHost, "Ollama server host")
	flag.IntVar(&c.OllamaPort, "ollama-port", c.OllamaPort, "Ollama server port")
	flag.IntVar(&c.ProxyPort, "proxy-port", c.ProxyPort, "Proxy server port")
	flag.IntVar(&c.MetricsPort, "metrics-port", c.MetricsPort, "Metrics server port")
	flag.StringVar(&c.LogLevel, "log-level", c.LogLevel, "Log level (debug, info, warn, error)")
	flag.IntVar(&c.MaxQueueSize, "max-queue-size", c.MaxQueueSize, "Maximum request queue size")
	flag.IntVar(&c.MaxConcurrency, "max-concurrency", c.MaxConcurrency, "Maximum concurrent requests to Ollama")

	flag.Parse()

	// Update shared config fields
	c.Port = c.ProxyPort
	c.OllamaConfig.URL = c.OllamaURL()
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	if host := os.Getenv("OLLAMA_HOST"); host != "" {
		c.OllamaHost = host
	}

	if port := os.Getenv("OLLAMA_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.OllamaPort)
	}

	if port := os.Getenv("PROXY_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.ProxyPort)
	}

	if port := os.Getenv("METRICS_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &c.MetricsPort)
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		c.LogLevel = level
	}

	if size := os.Getenv("MAX_QUEUE_SIZE"); size != "" {
		fmt.Sscanf(size, "%d", &c.MaxQueueSize)
	}

	if concurrency := os.Getenv("MAX_CONCURRENCY"); concurrency != "" {
		fmt.Sscanf(concurrency, "%d", &c.MaxConcurrency)
	}

	// Update shared config fields
	c.Port = c.ProxyPort
	c.OllamaConfig.URL = c.OllamaURL()
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate base config first
	if err := c.BaseConfig.Validate(); err != nil {
		return err
	}

	if c.OllamaPort <= 0 || c.OllamaPort > 65535 {
		return fmt.Errorf("invalid Ollama port: %d", c.OllamaPort)
	}

	if c.ProxyPort <= 0 || c.ProxyPort > 65535 {
		return fmt.Errorf("invalid proxy port: %d", c.ProxyPort)
	}

	if c.ProxyPort == c.MetricsPort {
		return fmt.Errorf("proxy port and metrics port cannot be the same")
	}

	return nil
}

// OllamaURL returns the full URL for the Ollama server
func (c *Config) OllamaURL() string {
	return fmt.Sprintf("http://%s:%d", c.OllamaHost, c.OllamaPort)
}