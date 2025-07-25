// Package config provides shared configuration structures and utilities
// for all services in the llama-metrics project.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// BaseConfig contains common configuration fields used by all services
type BaseConfig struct {
	// Service identification
	ServiceName string `yaml:"service_name" json:"service_name"`
	Version     string `yaml:"version" json:"version"`

	// Server configuration
	Port        int    `yaml:"port" json:"port"`
	MetricsPort int    `yaml:"metrics_port" json:"metrics_port"`
	LogLevel    string `yaml:"log_level" json:"log_level"`

	// Timeouts
	ReadTimeout  time.Duration `yaml:"read_timeout" json:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout" json:"write_timeout"`

	// Health check
	HealthPath string `yaml:"health_path" json:"health_path"`
}

// PrometheusConfig contains Prometheus-specific configuration
type PrometheusConfig struct {
	MetricsPath      string `yaml:"metrics_path" json:"metrics_path"`
	MetricsNamespace string `yaml:"metrics_namespace" json:"metrics_namespace"`
}

// OllamaConfig contains Ollama-specific configuration
type OllamaConfig struct {
	URL     string        `yaml:"url" json:"url"`
	Timeout time.Duration `yaml:"timeout" json:"timeout"`
}

// LoadEnvString loads a string value from environment variable with a default
func LoadEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// LoadEnvInt loads an integer value from environment variable with a default
func LoadEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// LoadEnvBool loads a boolean value from environment variable with a default
func LoadEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// LoadEnvDuration loads a duration value from environment variable with a default
func LoadEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// DefaultBaseConfig returns a BaseConfig with sensible defaults
func DefaultBaseConfig(serviceName string) BaseConfig {
	return BaseConfig{
		ServiceName:  serviceName,
		Version:      LoadEnvString("VERSION", "1.0.0"),
		Port:         LoadEnvInt("PORT", 8080),
		MetricsPort:  LoadEnvInt("METRICS_PORT", 8001),
		LogLevel:     LoadEnvString("LOG_LEVEL", "info"),
		ReadTimeout:  LoadEnvDuration("READ_TIMEOUT", 30*time.Second),
		WriteTimeout: LoadEnvDuration("WRITE_TIMEOUT", 30*time.Second),
		HealthPath:   LoadEnvString("HEALTH_PATH", "/health"),
	}
}

// Validate performs basic validation on the configuration
func (c *BaseConfig) Validate() error {
	if c.ServiceName == "" {
		return fmt.Errorf("service name is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", c.Port)
	}
	if c.MetricsPort <= 0 || c.MetricsPort > 65535 {
		return fmt.Errorf("invalid metrics port number: %d", c.MetricsPort)
	}
	if c.Port == c.MetricsPort {
		return fmt.Errorf("port and metrics port cannot be the same: %d", c.Port)
	}
	return nil
}