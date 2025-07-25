// Package models provides shared data models and structures
// for all services in the llama-metrics project.
package models

import (
	"time"
)

// HealthStatus represents the health status of a service
type HealthStatus struct {
	Status      string            `json:"status"`
	Service     string            `json:"service"`
	Version     string            `json:"version"`
	Uptime      time.Duration     `json:"uptime"`
	Timestamp   time.Time         `json:"timestamp"`
	Checks      []HealthCheck     `json:"checks,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

// OllamaRequest represents a request to Ollama
type OllamaRequest struct {
	Model    string                 `json:"model"`
	Prompt   string                 `json:"prompt,omitempty"`
	Messages []Message              `json:"messages,omitempty"`
	Stream   bool                   `json:"stream"`
	Options  map[string]interface{} `json:"options,omitempty"`
	System   string                 `json:"system,omitempty"`
	Template string                 `json:"template,omitempty"`
	Context  []int                  `json:"context,omitempty"`
	Format   string                 `json:"format,omitempty"`
}

// OllamaResponse represents a response from Ollama
type OllamaResponse struct {
	Model              string    `json:"model"`
	CreatedAt          time.Time `json:"created_at"`
	Response           string    `json:"response,omitempty"`
	Message            *Message  `json:"message,omitempty"`
	Done               bool      `json:"done"`
	Context            []int     `json:"context,omitempty"`
	TotalDuration      int64     `json:"total_duration,omitempty"`
	LoadDuration       int64     `json:"load_duration,omitempty"`
	PromptEvalCount    int       `json:"prompt_eval_count,omitempty"`
	PromptEvalDuration int64     `json:"prompt_eval_duration,omitempty"`
	EvalCount          int       `json:"eval_count,omitempty"`
	EvalDuration       int64     `json:"eval_duration,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModelInfo represents information about a loaded model
type ModelInfo struct {
	Name       string    `json:"name"`
	Modified   time.Time `json:"modified"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Parameters string    `json:"parameters,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error      string            `json:"error"`
	Message    string            `json:"message,omitempty"`
	StatusCode int               `json:"status_code,omitempty"`
	Details    map[string]string `json:"details,omitempty"`
	Timestamp  time.Time         `json:"timestamp"`
}

// MetricsSnapshot represents a point-in-time metrics snapshot
type MetricsSnapshot struct {
	Timestamp       time.Time              `json:"timestamp"`
	Service         string                 `json:"service"`
	RequestRate     float64                `json:"request_rate"`
	ErrorRate       float64                `json:"error_rate"`
	AvgLatency      float64                `json:"avg_latency_ms"`
	ActiveRequests  int                    `json:"active_requests"`
	QueueSize       int                    `json:"queue_size,omitempty"`
	SystemMetrics   *SystemMetrics         `json:"system_metrics,omitempty"`
	CustomMetrics   map[string]interface{} `json:"custom_metrics,omitempty"`
}

// SystemMetrics represents system-level metrics
type SystemMetrics struct {
	CPUUsage       float64            `json:"cpu_usage_percent"`
	MemoryUsed     int64              `json:"memory_used_bytes"`
	MemoryTotal    int64              `json:"memory_total_bytes"`
	GPUMetrics     []GPUMetrics       `json:"gpu_metrics,omitempty"`
	DiskUsage      map[string]float64 `json:"disk_usage_percent,omitempty"`
}

// GPUMetrics represents GPU-specific metrics
type GPUMetrics struct {
	Index       int     `json:"index"`
	Name        string  `json:"name"`
	Usage       float64 `json:"usage_percent"`
	MemoryUsed  int64   `json:"memory_used_bytes"`
	MemoryTotal int64   `json:"memory_total_bytes"`
	Temperature float64 `json:"temperature_celsius"`
	PowerDraw   float64 `json:"power_draw_watts"`
}

// QueueItem represents an item in a processing queue
type QueueItem struct {
	ID        string      `json:"id"`
	Priority  int         `json:"priority"`
	CreatedAt time.Time   `json:"created_at"`
	Data      interface{} `json:"data"`
}

// Constants for health and status checks
const (
	StatusHealthy   = "healthy"
	StatusDegraded  = "degraded"
	StatusUnhealthy = "unhealthy"
	StatusUnknown   = "unknown"
)

// Constants for message roles
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
)