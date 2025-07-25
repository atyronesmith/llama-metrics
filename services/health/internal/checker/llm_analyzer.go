package checker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/atyronesmith/llama-metrics/health/internal/models"
)

// AnalyzeHealthWithLLM uses Ollama to analyze the health status and provide insights
func (hc *HealthChecker) AnalyzeHealthWithLLM(ctx context.Context, health models.SystemHealth) models.LLMAnalysis {
	analysis := models.LLMAnalysis{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	// First check if Ollama is available
	ollamaHealthy := false
	for _, service := range health.Services {
		if service.Name == "ollama" && service.Status.Status == "healthy" {
			ollamaHealthy = true
			break
		}
	}

	if !ollamaHealthy {
		analysis.Available = false
		analysis.Error = "Ollama is not available for analysis"
		return analysis
	}

	// Prepare the health data as a structured prompt
	prompt := hc.buildAnalysisPrompt(health)

	// Call Ollama to analyze
	response, err := hc.callOllamaForAnalysis(ctx, prompt)
	if err != nil {
		analysis.Available = false
		analysis.Error = fmt.Sprintf("Failed to get analysis from Ollama: %v", err)
		return analysis
	}

	analysis.Available = true
	analysis.Summary = response
	analysis.Details = map[string]interface{}{
		"model":         hc.config.Models.DefaultModel,
		"health_status": health.Status,
		"services":      len(health.Services),
	}

	return analysis
}

func (hc *HealthChecker) buildAnalysisPrompt(health models.SystemHealth) string {
	var sb strings.Builder

	sb.WriteString("You are a system health analyzer. Analyze the following health check data and provide a concise summary with insights and recommendations.\n\n")

	// Overall status
	sb.WriteString(fmt.Sprintf("OVERALL STATUS: %s\n", strings.ToUpper(health.Status)))
	sb.WriteString(fmt.Sprintf("Uptime: %.1f hours\n\n", health.UptimeSeconds/3600))

	// Service status
	sb.WriteString("SERVICE STATUS:\n")
	for _, service := range health.Services {
		status := "✅"
		if service.Status.Status != "healthy" {
			status = "❌"
		}
		sb.WriteString(fmt.Sprintf("%s %s: %s", status, service.Name, service.Status.Status))
		if service.Status.Error != nil {
			sb.WriteString(fmt.Sprintf(" (Error: %s)", *service.Status.Error))
		}
		if service.Status.ResponseTimeMs != nil {
			sb.WriteString(fmt.Sprintf(" [%dms]", int(*service.Status.ResponseTimeMs)))
		}
		sb.WriteString("\n")
	}

	// System metrics
	sb.WriteString(fmt.Sprintf("\nSYSTEM METRICS:\n"))
	if len(health.SystemMetrics.CPU.LoadAvg) >= 3 {
		sb.WriteString(fmt.Sprintf("- CPU: %.1f%% (Load: %.2f, %.2f, %.2f)\n",
			health.SystemMetrics.CPU.Percent,
			health.SystemMetrics.CPU.LoadAvg[0],
			health.SystemMetrics.CPU.LoadAvg[1],
			health.SystemMetrics.CPU.LoadAvg[2]))
	} else {
		sb.WriteString(fmt.Sprintf("- CPU: %.1f%%\n", health.SystemMetrics.CPU.Percent))
	}
	sb.WriteString(fmt.Sprintf("- Memory: %.1f%% (%.1f/%.1f GB used)\n",
		health.SystemMetrics.Memory.Percent,
		health.SystemMetrics.Memory.UsedGB,
		health.SystemMetrics.Memory.TotalGB))
	sb.WriteString(fmt.Sprintf("- Disk: %.1f%% (%.1f/%.1f GB used)\n",
		health.SystemMetrics.Disk.Percent,
		health.SystemMetrics.Disk.UsedGB,
		health.SystemMetrics.Disk.TotalGB))

	// Special notes
	if health.SystemMetrics.CPU.Percent > 80 {
		sb.WriteString("\n⚠️ HIGH CPU USAGE DETECTED\n")
	}
	if health.SystemMetrics.Memory.Percent > 85 {
		sb.WriteString("\n⚠️ HIGH MEMORY USAGE DETECTED\n")
	}

	sb.WriteString("\nProvide a brief analysis including:\n")
	sb.WriteString("1. Overall system health assessment\n")
	sb.WriteString("2. Any issues or concerns identified\n")
	sb.WriteString("3. Specific recommendations for any problems\n")
	sb.WriteString("4. Performance optimization suggestions if applicable\n")
	sb.WriteString("\nKeep the response concise and actionable.")

	return sb.String()
}

func (hc *HealthChecker) callOllamaForAnalysis(ctx context.Context, prompt string) (string, error) {
	// Create the request
	reqBody := map[string]interface{}{
		"model":  hc.config.Models.DefaultModel,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.7,
			"num_predict": 500, // Keep analysis concise
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the request with a reasonable timeout
	reqCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, "POST",
		fmt.Sprintf("%s/api/generate", hc.config.Server.OllamaURL),
		bytes.NewReader(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	response, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	return strings.TrimSpace(response), nil
}