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

	sb.WriteString("Analyze this system health data. Give 2-3 sentence summary.\n\n")

	// Overall status
	sb.WriteString(fmt.Sprintf("Status: %s\n", health.Status))

	// Service issues only
	unhealthyServices := []string{}
	for _, service := range health.Services {
		if service.Status.Status != "healthy" {
			if service.Status.Error != nil {
				unhealthyServices = append(unhealthyServices, fmt.Sprintf("%s: %s", service.Name, *service.Status.Error))
			} else {
				unhealthyServices = append(unhealthyServices, service.Name)
			}
		}
	}

	if len(unhealthyServices) > 0 {
		sb.WriteString(fmt.Sprintf("Issues: %s\n", strings.Join(unhealthyServices, ", ")))
	}

	// Key metrics only
	sb.WriteString(fmt.Sprintf("CPU: %.0f%%, Memory: %.0f%%, Disk: %.0f%%\n",
		health.SystemMetrics.CPU.Percent,
		health.SystemMetrics.Memory.Percent,
		health.SystemMetrics.Disk.Percent))

	sb.WriteString("\nSummarize the health status and any critical issues:")

	return sb.String()
}

func (hc *HealthChecker) callOllamaForAnalysis(ctx context.Context, prompt string) (string, error) {
	// Create the request
	reqBody := map[string]interface{}{
		"model":  hc.config.Models.DefaultModel,
		"prompt": prompt,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.3,
			"num_predict": 150, // Keep analysis very concise
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make the request with a very short timeout for simplified analysis
	reqCtx, cancel := context.WithTimeout(ctx, 8*time.Second)
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