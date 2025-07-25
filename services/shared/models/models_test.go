package models

import (
	"testing"
	"time"
)

func TestHealthStatus_Creation(t *testing.T) {
	now := time.Now()
	status := HealthStatus{
		Status:    StatusHealthy,
		Service:   "test-service",
		Version:   "1.0.0",
		Uptime:    5 * time.Minute,
		Timestamp: now,
		Details:   make(map[string]string),
	}

	if status.Status != StatusHealthy {
		t.Errorf("Expected status %s, got %s", StatusHealthy, status.Status)
	}
	if status.Service != "test-service" {
		t.Errorf("Expected service 'test-service', got %s", status.Service)
	}
	if status.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got %s", status.Version)
	}
	if status.Uptime != 5*time.Minute {
		t.Errorf("Expected uptime 5m, got %v", status.Uptime)
	}
}

func TestHealthCheck_Creation(t *testing.T) {
	check := HealthCheck{
		Name:    "database",
		Status:  StatusHealthy,
		Message: "Connection successful",
	}

	if check.Name != "database" {
		t.Errorf("Expected name 'database', got %s", check.Name)
	}
	if check.Status != StatusHealthy {
		t.Errorf("Expected status %s, got %s", StatusHealthy, check.Status)
	}
	if check.Message != "Connection successful" {
		t.Errorf("Expected message 'Connection successful', got %s", check.Message)
	}
}

func TestOllamaRequest_Creation(t *testing.T) {
	request := OllamaRequest{
		Model:  "phi3:mini",
		Prompt: "Hello, world!",
		Stream: false,
		Options: map[string]interface{}{
			"temperature": 0.7,
			"top_p":       0.9,
		},
	}

	if request.Model != "phi3:mini" {
		t.Errorf("Expected model 'phi3:mini', got %s", request.Model)
	}
	if request.Prompt != "Hello, world!" {
		t.Errorf("Expected prompt 'Hello, world!', got %s", request.Prompt)
	}
	if request.Stream != false {
		t.Errorf("Expected stream false, got %v", request.Stream)
	}
	if temp, ok := request.Options["temperature"]; !ok || temp != 0.7 {
		t.Errorf("Expected temperature 0.7, got %v", temp)
	}
}

func TestOllamaResponse_Creation(t *testing.T) {
	now := time.Now()
	response := OllamaResponse{
		Model:              "phi3:mini",
		CreatedAt:          now,
		Response:           "Hello there!",
		Done:               true,
		TotalDuration:      1500000000, // 1.5 seconds in nanoseconds
		LoadDuration:       500000000,  // 0.5 seconds in nanoseconds
		PromptEvalCount:    10,
		PromptEvalDuration: 200000000, // 0.2 seconds in nanoseconds
		EvalCount:          15,
		EvalDuration:       800000000, // 0.8 seconds in nanoseconds
	}

	if response.Model != "phi3:mini" {
		t.Errorf("Expected model 'phi3:mini', got %s", response.Model)
	}
	if response.Response != "Hello there!" {
		t.Errorf("Expected response 'Hello there!', got %s", response.Response)
	}
	if !response.Done {
		t.Errorf("Expected done true, got %v", response.Done)
	}
	if response.PromptEvalCount != 10 {
		t.Errorf("Expected prompt eval count 10, got %d", response.PromptEvalCount)
	}
	if response.EvalCount != 15 {
		t.Errorf("Expected eval count 15, got %d", response.EvalCount)
	}
}

func TestMessage_Creation(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		content  string
	}{
		{
			name:    "system message",
			role:    RoleSystem,
			content: "You are a helpful assistant.",
		},
		{
			name:    "user message",
			role:    RoleUser,
			content: "What is the weather today?",
		},
		{
			name:    "assistant message",
			role:    RoleAssistant,
			content: "I don't have access to current weather data.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			message := Message{
				Role:    tt.role,
				Content: tt.content,
			}

			if message.Role != tt.role {
				t.Errorf("Expected role %s, got %s", tt.role, message.Role)
			}
			if message.Content != tt.content {
				t.Errorf("Expected content %s, got %s", tt.content, message.Content)
			}
		})
	}
}

func TestModelInfo_Creation(t *testing.T) {
	modified := time.Now()
	model := ModelInfo{
		Name:       "phi3:mini",
		Modified:   modified,
		Size:       2048000000, // 2GB
		Digest:     "sha256:abc123",
		Parameters: "3.8B",
	}

	if model.Name != "phi3:mini" {
		t.Errorf("Expected name 'phi3:mini', got %s", model.Name)
	}
	if model.Size != 2048000000 {
		t.Errorf("Expected size 2048000000, got %d", model.Size)
	}
	if model.Digest != "sha256:abc123" {
		t.Errorf("Expected digest 'sha256:abc123', got %s", model.Digest)
	}
	if model.Parameters != "3.8B" {
		t.Errorf("Expected parameters '3.8B', got %s", model.Parameters)
	}
}

func TestErrorResponse_Creation(t *testing.T) {
	now := time.Now()
	errResp := ErrorResponse{
		Error:      "validation failed",
		Message:    "Model name is required",
		StatusCode: 400,
		Details: map[string]string{
			"field": "model",
			"value": "",
		},
		Timestamp: now,
	}

	if errResp.Error != "validation failed" {
		t.Errorf("Expected error 'validation failed', got %s", errResp.Error)
	}
	if errResp.Message != "Model name is required" {
		t.Errorf("Expected message 'Model name is required', got %s", errResp.Message)
	}
	if errResp.StatusCode != 400 {
		t.Errorf("Expected status code 400, got %d", errResp.StatusCode)
	}
	if field, ok := errResp.Details["field"]; !ok || field != "model" {
		t.Errorf("Expected details field 'model', got %v", field)
	}
}

func TestSystemMetrics_Creation(t *testing.T) {
	metrics := SystemMetrics{
		CPUUsage:    75.5,
		MemoryUsed:  4096000000, // 4GB
		MemoryTotal: 8192000000, // 8GB
		DiskUsage: map[string]float64{
			"/":     60.0,
			"/home": 45.0,
		},
	}

	if metrics.CPUUsage != 75.5 {
		t.Errorf("Expected CPU usage 75.5, got %f", metrics.CPUUsage)
	}
	if metrics.MemoryUsed != 4096000000 {
		t.Errorf("Expected memory used 4096000000, got %d", metrics.MemoryUsed)
	}
	if metrics.MemoryTotal != 8192000000 {
		t.Errorf("Expected memory total 8192000000, got %d", metrics.MemoryTotal)
	}
	if usage, ok := metrics.DiskUsage["/"]; !ok || usage != 60.0 {
		t.Errorf("Expected root disk usage 60.0, got %v", usage)
	}
}

func TestGPUMetrics_Creation(t *testing.T) {
	gpu := GPUMetrics{
		Index:       0,
		Name:        "NVIDIA RTX 4090",
		Usage:       85.0,
		MemoryUsed:  12000000000, // 12GB
		MemoryTotal: 24000000000, // 24GB
		Temperature: 72.5,
		PowerDraw:   350.0,
	}

	if gpu.Index != 0 {
		t.Errorf("Expected index 0, got %d", gpu.Index)
	}
	if gpu.Name != "NVIDIA RTX 4090" {
		t.Errorf("Expected name 'NVIDIA RTX 4090', got %s", gpu.Name)
	}
	if gpu.Usage != 85.0 {
		t.Errorf("Expected usage 85.0, got %f", gpu.Usage)
	}
	if gpu.Temperature != 72.5 {
		t.Errorf("Expected temperature 72.5, got %f", gpu.Temperature)
	}
	if gpu.PowerDraw != 350.0 {
		t.Errorf("Expected power draw 350.0, got %f", gpu.PowerDraw)
	}
}

func TestMetricsSnapshot_Creation(t *testing.T) {
	now := time.Now()
	systemMetrics := &SystemMetrics{
		CPUUsage:    50.0,
		MemoryUsed:  2048000000,
		MemoryTotal: 8192000000,
	}

	snapshot := MetricsSnapshot{
		Timestamp:      now,
		Service:        "ollama-proxy",
		RequestRate:    15.5,
		ErrorRate:      0.02,
		AvgLatency:     125.5,
		ActiveRequests: 3,
		QueueSize:      7,
		SystemMetrics:  systemMetrics,
		CustomMetrics: map[string]interface{}{
			"tokens_per_second": 45.2,
			"model_loads":       12,
		},
	}

	if snapshot.Service != "ollama-proxy" {
		t.Errorf("Expected service 'ollama-proxy', got %s", snapshot.Service)
	}
	if snapshot.RequestRate != 15.5 {
		t.Errorf("Expected request rate 15.5, got %f", snapshot.RequestRate)
	}
	if snapshot.ErrorRate != 0.02 {
		t.Errorf("Expected error rate 0.02, got %f", snapshot.ErrorRate)
	}
	if snapshot.AvgLatency != 125.5 {
		t.Errorf("Expected avg latency 125.5, got %f", snapshot.AvgLatency)
	}
	if snapshot.ActiveRequests != 3 {
		t.Errorf("Expected active requests 3, got %d", snapshot.ActiveRequests)
	}
	if snapshot.QueueSize != 7 {
		t.Errorf("Expected queue size 7, got %d", snapshot.QueueSize)
	}
	if snapshot.SystemMetrics.CPUUsage != 50.0 {
		t.Errorf("Expected system CPU usage 50.0, got %f", snapshot.SystemMetrics.CPUUsage)
	}
	if tps, ok := snapshot.CustomMetrics["tokens_per_second"]; !ok || tps != 45.2 {
		t.Errorf("Expected tokens per second 45.2, got %v", tps)
	}
}

func TestQueueItem_Creation(t *testing.T) {
	now := time.Now()
	request := OllamaRequest{
		Model:  "phi3:mini",
		Prompt: "Test prompt",
	}

	item := QueueItem{
		ID:        "req-123",
		Priority:  5,
		CreatedAt: now,
		Data:      request,
	}

	if item.ID != "req-123" {
		t.Errorf("Expected ID 'req-123', got %s", item.ID)
	}
	if item.Priority != 5 {
		t.Errorf("Expected priority 5, got %d", item.Priority)
	}
	if item.CreatedAt != now {
		t.Errorf("Expected created at %v, got %v", now, item.CreatedAt)
	}

	// Type assertion to check the data
	if req, ok := item.Data.(OllamaRequest); ok {
		if req.Model != "phi3:mini" {
			t.Errorf("Expected data model 'phi3:mini', got %s", req.Model)
		}
	} else {
		t.Errorf("Expected data to be OllamaRequest, got %T", item.Data)
	}
}

func TestConstants(t *testing.T) {
	// Test health status constants
	if StatusHealthy != "healthy" {
		t.Errorf("Expected StatusHealthy 'healthy', got %s", StatusHealthy)
	}
	if StatusDegraded != "degraded" {
		t.Errorf("Expected StatusDegraded 'degraded', got %s", StatusDegraded)
	}
	if StatusUnhealthy != "unhealthy" {
		t.Errorf("Expected StatusUnhealthy 'unhealthy', got %s", StatusUnhealthy)
	}
	if StatusUnknown != "unknown" {
		t.Errorf("Expected StatusUnknown 'unknown', got %s", StatusUnknown)
	}

	// Test role constants
	if RoleSystem != "system" {
		t.Errorf("Expected RoleSystem 'system', got %s", RoleSystem)
	}
	if RoleUser != "user" {
		t.Errorf("Expected RoleUser 'user', got %s", RoleUser)
	}
	if RoleAssistant != "assistant" {
		t.Errorf("Expected RoleAssistant 'assistant', got %s", RoleAssistant)
	}
}