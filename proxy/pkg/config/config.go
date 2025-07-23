package config

import (
	"flag"
	"fmt"
	"os"
)

// Config holds the proxy configuration
type Config struct {
	OllamaHost     string
	OllamaPort     int
	ProxyPort      int
	MetricsPort    int
	LogLevel       string
	MaxQueueSize   int
	MaxConcurrency int
}

// DefaultConfig returns a Config with default values
func DefaultConfig() *Config {
	return &Config{
		OllamaHost:     "localhost",
		OllamaPort:     11434,
		ProxyPort:      11435,
		MetricsPort:    8001,
		LogLevel:       "info",
		MaxQueueSize:   100,
		MaxConcurrency: 10,
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
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.OllamaPort <= 0 || c.OllamaPort > 65535 {
		return fmt.Errorf("invalid Ollama port: %d", c.OllamaPort)
	}

	if c.ProxyPort <= 0 || c.ProxyPort > 65535 {
		return fmt.Errorf("invalid proxy port: %d", c.ProxyPort)
	}

	if c.MetricsPort <= 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("invalid metrics port: %d", c.MetricsPort)
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