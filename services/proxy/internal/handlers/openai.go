package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/atyronesmith/llama-metrics/proxy/internal/metrics"
	"github.com/atyronesmith/llama-metrics/proxy/internal/models"
	"github.com/atyronesmith/llama-metrics/proxy/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OpenAIHandler handles OpenAI-compatible API requests
type OpenAIHandler struct {
	config     *config.Config
	metrics    *metrics.Collector
	httpClient *http.Client
}

// NewOpenAIHandler creates a new OpenAI handler
func NewOpenAIHandler(cfg *config.Config, m *metrics.Collector) *OpenAIHandler {
	return &OpenAIHandler{
		config:  cfg,
		metrics: m,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// HandleChatCompletions handles the /v1/chat/completions endpoint
func (h *OpenAIHandler) HandleChatCompletions(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()
	model := "unknown"

	// Add request ID to response headers
	c.Header("X-Request-ID", requestID)

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_body")
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}

	// Record request size
	h.metrics.RecordRequestSize(model, "/v1/chat/completions", len(body))

	// Parse OpenAI request
	var openAIReq models.ChatCompletionRequest
	if err := json.Unmarshal(body, &openAIReq); err != nil {
		h.metrics.RecordError(model, "parse_request")
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "Invalid request format")
		return
	}

	model = h.mapOpenAIModelToOllama(openAIReq.Model)

	// Track active requests
	h.metrics.IncActiveRequests(model)
	defer h.metrics.DecActiveRequests(model)

	// Convert to Ollama format
	ollamaReq := h.convertChatToOllama(openAIReq)

	// Call Ollama
	if openAIReq.Stream {
		h.handleStreamingChatCompletion(c, ollamaReq, openAIReq, model, requestID, start)
	} else {
		h.handleNonStreamingChatCompletion(c, ollamaReq, openAIReq, model, requestID, start)
	}
}

// HandleCompletions handles the /v1/completions endpoint
func (h *OpenAIHandler) HandleCompletions(c *gin.Context) {
	start := time.Now()
	requestID := uuid.New().String()
	model := "unknown"

	// Add request ID to response headers
	c.Header("X-Request-ID", requestID)

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_body")
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "Failed to read request body")
		return
	}

	// Record request size
	h.metrics.RecordRequestSize(model, "/v1/completions", len(body))

	// Parse OpenAI request
	var openAIReq models.CompletionRequest
	if err := json.Unmarshal(body, &openAIReq); err != nil {
		h.metrics.RecordError(model, "parse_request")
		h.sendOpenAIError(c, http.StatusBadRequest, "invalid_request_error", "Invalid request format")
		return
	}

	model = h.mapOpenAIModelToOllama(openAIReq.Model)

	// Track active requests
	h.metrics.IncActiveRequests(model)
	defer h.metrics.DecActiveRequests(model)

	// Convert to Ollama format
	ollamaReq := h.convertCompletionToOllama(openAIReq)

	// Call Ollama
	if openAIReq.Stream {
		h.handleStreamingCompletion(c, ollamaReq, openAIReq, model, requestID, start)
	} else {
		h.handleNonStreamingCompletion(c, ollamaReq, openAIReq, model, requestID, start)
	}
}

// convertChatToOllama converts OpenAI chat request to Ollama format
func (h *OpenAIHandler) convertChatToOllama(openAIReq models.ChatCompletionRequest) models.ChatRequest {
	messages := make([]models.Message, len(openAIReq.Messages))
	for i, msg := range openAIReq.Messages {
		messages[i] = models.Message{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	options := make(map[string]interface{})
	if openAIReq.Temperature > 0 {
		options["temperature"] = openAIReq.Temperature
	}
	if openAIReq.TopP > 0 {
		options["top_p"] = openAIReq.TopP
	}
	if openAIReq.MaxTokens > 0 {
		options["num_predict"] = openAIReq.MaxTokens
	}
	if openAIReq.Stop != nil {
		options["stop"] = openAIReq.Stop
	}
	if openAIReq.Seed > 0 {
		options["seed"] = openAIReq.Seed
	}

	return models.ChatRequest{
		Model:    h.mapOpenAIModelToOllama(openAIReq.Model),
		Messages: messages,
		Stream:   openAIReq.Stream,
		Options:  options,
	}
}

// convertCompletionToOllama converts OpenAI completion request to Ollama format
func (h *OpenAIHandler) convertCompletionToOllama(openAIReq models.CompletionRequest) models.GenerateRequest {
	prompt := ""
	switch p := openAIReq.Prompt.(type) {
	case string:
		prompt = p
	case []string:
		if len(p) > 0 {
			prompt = p[0]
		}
	}

	options := make(map[string]interface{})
	if openAIReq.Temperature > 0 {
		options["temperature"] = openAIReq.Temperature
	}
	if openAIReq.TopP > 0 {
		options["top_p"] = openAIReq.TopP
	}
	if openAIReq.MaxTokens > 0 {
		options["num_predict"] = openAIReq.MaxTokens
	}
	if openAIReq.Stop != nil {
		options["stop"] = openAIReq.Stop
	}

	return models.GenerateRequest{
		Model:   h.mapOpenAIModelToOllama(openAIReq.Model),
		Prompt:  prompt,
		Stream:  openAIReq.Stream,
		Options: options,
	}
}

// handleStreamingChatCompletion handles streaming chat completion
func (h *OpenAIHandler) handleStreamingChatCompletion(c *gin.Context, ollamaReq models.ChatRequest, openAIReq models.ChatCompletionRequest, model, requestID string, start time.Time) {
	// Make request to Ollama
	reqBody, _ := json.Marshal(ollamaReq)
	targetURL := fmt.Sprintf("%s/api/chat", h.config.OllamaURL())

	proxyReq, err := http.NewRequest("POST", targetURL, bytes.NewReader(reqBody))
	if err != nil {
		h.metrics.RecordError(model, "create_request")
		h.sendOpenAIError(c, http.StatusInternalServerError, "internal_error", "Failed to create request")
		return
	}

	proxyReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		h.metrics.RecordError(model, "proxy_request")
		h.sendOpenAIError(c, http.StatusBadGateway, "internal_error", "Failed to proxy request")
		return
	}
	defer resp.Body.Close()

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	// Process streaming response
	scanner := bufio.NewScanner(resp.Body)
	firstTokenTime := time.Time{}
	promptTokens := 0
	generatedTokens := 0
	var evalDuration int64
	var accumulatedContent strings.Builder

	for scanner.Scan() {
		line := scanner.Bytes()

		var ollamaResp models.ChatResponse
		if err := json.Unmarshal(line, &ollamaResp); err != nil {
			continue
		}

		// Record time to first token
		if firstTokenTime.IsZero() && ollamaResp.Message.Content != "" {
			firstTokenTime = time.Now()
			h.metrics.RecordTimeToFirstToken(model, firstTokenTime.Sub(start))
		}

		// Accumulate content
		accumulatedContent.WriteString(ollamaResp.Message.Content)

		// Convert to OpenAI format
		openAIResp := models.StreamingChatCompletionResponse{
			ID:      requestID,
			Object:  "chat.completion.chunk",
			Created: time.Now().Unix(),
			Model:   openAIReq.Model,
			Choices: []models.ChatChoice{
				{
					Index: 0,
					Delta: &models.ChatMessage{
						Content: ollamaResp.Message.Content,
					},
				},
			},
		}

		// Add finish reason if done
		if ollamaResp.Done {
			openAIResp.Choices[0].FinishReason = "stop"
			promptTokens = ollamaResp.PromptEvalCount
			generatedTokens = ollamaResp.EvalCount
			evalDuration = ollamaResp.EvalDuration
		}

		// Send the chunk
		data, _ := json.Marshal(openAIResp)
		c.SSEvent("", fmt.Sprintf("data: %s\n\n", string(data)))
		c.Writer.Flush()
	}

	// Send final [DONE] message
	c.SSEvent("", "data: [DONE]\n\n")
	c.Writer.Flush()

	// Record metrics
	duration := time.Since(start)
	h.metrics.RecordRequest("POST", "/v1/chat/completions", model, "200", duration)

	// Calculate and record token metrics
	totalTokens := promptTokens + generatedTokens
	var tokensPerSec float64
	if evalDuration > 0 && generatedTokens > 0 {
		tokensPerSec = float64(generatedTokens) / (float64(evalDuration) / 1e9)
	}
	h.metrics.RecordTokens(model, promptTokens, generatedTokens, tokensPerSec)

	// Record enhanced metrics
	h.metrics.RecordRequestMetadata(models.RequestMetadata{
		RequestID:        requestID,
		Model:            model,
		User:             openAIReq.User,
		StartTime:        start,
		EndTime:          time.Now(),
		PromptTokens:     promptTokens,
		CompletionTokens: generatedTokens,
		TotalTokens:      totalTokens,
		Stream:           true,
		StatusCode:       200,
		Endpoint:         "/v1/chat/completions",
		Method:           "POST",
		ResponseTime:     duration,
		TimeToFirstToken: firstTokenTime.Sub(start),
		TokensPerSecond:  tokensPerSec,
	})

	// Record response size (approximate for streaming)
	responseSize := len(accumulatedContent.String()) + 200 // Add overhead for JSON structure
	h.metrics.RecordResponseSize(model, "/v1/chat/completions", responseSize)
}

// handleNonStreamingChatCompletion handles non-streaming chat completion
func (h *OpenAIHandler) handleNonStreamingChatCompletion(c *gin.Context, ollamaReq models.ChatRequest, openAIReq models.ChatCompletionRequest, model, requestID string, start time.Time) {
	// Make request to Ollama
	reqBody, _ := json.Marshal(ollamaReq)
	targetURL := fmt.Sprintf("%s/api/chat", h.config.OllamaURL())

	proxyReq, err := http.NewRequest("POST", targetURL, bytes.NewReader(reqBody))
	if err != nil {
		h.metrics.RecordError(model, "create_request")
		h.sendOpenAIError(c, http.StatusInternalServerError, "internal_error", "Failed to create request")
		return
	}

	proxyReq.Header.Set("Content-Type", "application/json")

	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		h.metrics.RecordError(model, "proxy_request")
		h.sendOpenAIError(c, http.StatusBadGateway, "internal_error", "Failed to proxy request")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_response")
		h.sendOpenAIError(c, http.StatusBadGateway, "internal_error", "Failed to read response")
		return
	}

	var ollamaResp models.ChatResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		h.metrics.RecordError(model, "parse_response")
		h.sendOpenAIError(c, http.StatusBadGateway, "internal_error", "Failed to parse response")
		return
	}

	// Convert to OpenAI format
	openAIResp := models.ChatCompletionResponse{
		ID:      requestID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   openAIReq.Model,
		Choices: []models.ChatChoice{
			{
				Index:        0,
				Message:      models.ChatMessage{
					Role:    ollamaResp.Message.Role,
					Content: ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
		Usage: &models.Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
	}

	// Record metrics
	duration := time.Since(start)
	h.metrics.RecordRequest("POST", "/v1/chat/completions", model, "200", duration)

	// Calculate and record token metrics
	var tokensPerSec float64
	if ollamaResp.EvalDuration > 0 && ollamaResp.EvalCount > 0 {
		tokensPerSec = float64(ollamaResp.EvalCount) / (float64(ollamaResp.EvalDuration) / 1e9)
	}
	h.metrics.RecordTokens(model, ollamaResp.PromptEvalCount, ollamaResp.EvalCount, tokensPerSec)

	// Record enhanced metrics
	h.metrics.RecordRequestMetadata(models.RequestMetadata{
		RequestID:        requestID,
		Model:            model,
		User:             openAIReq.User,
		StartTime:        start,
		EndTime:          time.Now(),
		PromptTokens:     ollamaResp.PromptEvalCount,
		CompletionTokens: ollamaResp.EvalCount,
		TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		Stream:           false,
		StatusCode:       200,
		Endpoint:         "/v1/chat/completions",
		Method:           "POST",
		ResponseTime:     duration,
		TokensPerSecond:  tokensPerSec,
	})

	// Send response and record size
	respBody, _ := json.Marshal(openAIResp)
	h.metrics.RecordResponseSize(model, "/v1/chat/completions", len(respBody))

	c.JSON(http.StatusOK, openAIResp)
}

// handleStreamingCompletion handles streaming completion (legacy API)
func (h *OpenAIHandler) handleStreamingCompletion(c *gin.Context, ollamaReq models.GenerateRequest, openAIReq models.CompletionRequest, model, requestID string, start time.Time) {
	// Similar to handleStreamingChatCompletion but for the legacy completions API
	// Implementation omitted for brevity - follows same pattern
}

// handleNonStreamingCompletion handles non-streaming completion (legacy API)
func (h *OpenAIHandler) handleNonStreamingCompletion(c *gin.Context, ollamaReq models.GenerateRequest, openAIReq models.CompletionRequest, model, requestID string, start time.Time) {
	// Similar to handleNonStreamingChatCompletion but for the legacy completions API
	// Implementation omitted for brevity - follows same pattern
}

// mapOpenAIModelToOllama maps OpenAI model names to Ollama model names
func (h *OpenAIHandler) mapOpenAIModelToOllama(openAIModel string) string {
	// Map common OpenAI models to Ollama equivalents
	modelMap := map[string]string{
		"gpt-4":                    "llama2:70b",
		"gpt-4-turbo":             "llama2:70b",
		"gpt-3.5-turbo":           "llama2:13b",
		"gpt-3.5-turbo-16k":       "llama2:13b",
		"text-davinci-003":        "llama2:7b",
		"text-davinci-002":        "llama2:7b",
		"code-davinci-002":        "codellama:7b",
		"text-embedding-ada-002":  "nomic-embed-text",
	}

	if ollamaModel, ok := modelMap[openAIModel]; ok {
		return ollamaModel
	}

	// If no mapping found, return as-is (might be a direct Ollama model name)
	return openAIModel
}

// sendOpenAIError sends an OpenAI-formatted error response
func (h *OpenAIHandler) sendOpenAIError(c *gin.Context, statusCode int, errorType, message string) {
	errorResp := models.OpenAIError{
		Error: models.ErrorDetail{
			Message: message,
			Type:    errorType,
		},
	}
	c.JSON(statusCode, errorResp)
}