package handlers

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/atyronesmith/llama-metrics/proxy/internal/metrics"
	"github.com/atyronesmith/llama-metrics/proxy/internal/models"
	"github.com/atyronesmith/llama-metrics/proxy/internal/queue"
	"github.com/atyronesmith/llama-metrics/proxy/pkg/config"
	"github.com/gin-gonic/gin"
)

// ProxyHandler handles proxying requests to Ollama
type ProxyHandler struct {
	config      *config.Config
	metrics     *metrics.Collector
	httpClient  *http.Client
	queue       *queue.Manager
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler(cfg *config.Config, m *metrics.Collector) *ProxyHandler {
	h := &ProxyHandler{
		config:  cfg,
		metrics: m,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // Long timeout for LLM requests
		},
	}

	// Initialize queue manager
	h.queue = queue.NewManager(cfg.MaxQueueSize, cfg.MaxConcurrency, m)

	return h
}

// HandleGenerate handles the /api/generate endpoint
func (h *ProxyHandler) HandleGenerate(c *gin.Context) {
	start := time.Now()
	model := "unknown"

	// Extract priority from header (default to normal)
	priority := queue.PriorityNormal
	if priorityHeader := c.GetHeader("X-Priority"); priorityHeader == "high" {
		priority = queue.PriorityHigh
	}

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
		return
	}

	// Parse request to extract model
	var req models.GenerateRequest
	if err := json.Unmarshal(body, &req); err == nil {
		model = req.Model
	}

	// Submit to queue with priority
	err = h.queue.Submit(c.Request.Context(), model, priority, func() error {
		// Track active requests
		h.metrics.IncActiveRequests(model)
		defer h.metrics.DecActiveRequests(model)

		// Create request to Ollama
		targetURL := fmt.Sprintf("%s%s", h.config.OllamaURL(), c.Request.URL.Path)
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(body))
		if err != nil {
			h.metrics.RecordError(model, "create_request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return err
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Make request
		resp, err := h.httpClient.Do(proxyReq)
		if err != nil {
			h.metrics.RecordError(model, "proxy_request")
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to proxy request"})
			return err
		}
		defer resp.Body.Close()

		// Handle streaming vs non-streaming
		if req.Stream {
			h.handleStreamingResponse(c, resp, model, start)
		} else {
			h.handleNonStreamingResponse(c, resp, model, start)
		}

		return nil
	})

	if err != nil {
		h.metrics.RecordError(model, "queue_error")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	}
}

func (h *ProxyHandler) handleStreamingResponse(c *gin.Context, resp *http.Response, model string, start time.Time) {
	// Set headers for SSE
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Create a scanner to read the response line by line
	scanner := bufio.NewScanner(resp.Body)
	firstTokenTime := time.Time{}
	var totalPromptTokens, totalGeneratedTokens int
	var evalDuration int64

	for scanner.Scan() {
		line := scanner.Bytes()

		// Parse the JSON to extract metrics
		var chunk models.GenerateResponse
		if err := json.Unmarshal(line, &chunk); err == nil {
			// Record time to first token
			if firstTokenTime.IsZero() && chunk.Response != "" {
				firstTokenTime = time.Now()
				h.metrics.RecordTimeToFirstToken(model, firstTokenTime.Sub(start))
			}

			// Extract final metrics from done chunk
			if chunk.Done {
				totalPromptTokens = chunk.PromptEvalCount
				totalGeneratedTokens = chunk.EvalCount
				evalDuration = chunk.EvalDuration

				// Record model load time
				if chunk.LoadDuration > 0 {
					h.metrics.RecordModelLoadTime(model, time.Duration(chunk.LoadDuration))
				}
			}
		}

		// Write the chunk to response
		c.Data(http.StatusOK, "application/x-ndjson", line)
		c.Data(http.StatusOK, "application/x-ndjson", []byte("\n"))
		c.Writer.Flush()
	}

	// Record final metrics
	duration := time.Since(start)
	h.metrics.RecordRequest(c.Request.Method, c.Request.URL.Path, model, strconv.Itoa(resp.StatusCode), duration)

	// Record token metrics
	var tokensPerSec float64
	if evalDuration > 0 && totalGeneratedTokens > 0 {
		tokensPerSec = float64(totalGeneratedTokens) / (float64(evalDuration) / 1e9)
	}
	h.metrics.RecordTokens(model, totalPromptTokens, totalGeneratedTokens, tokensPerSec)
}

func (h *ProxyHandler) handleNonStreamingResponse(c *gin.Context, resp *http.Response, model string, start time.Time) {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_response")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to read response"})
		return
	}

	// Parse response to extract metrics
	var genResp models.GenerateResponse
	if err := json.Unmarshal(body, &genResp); err == nil {
		// Record model load time
		if genResp.LoadDuration > 0 {
			h.metrics.RecordModelLoadTime(model, time.Duration(genResp.LoadDuration))
		}

		// Record token metrics
		var tokensPerSec float64
		if genResp.EvalDuration > 0 && genResp.EvalCount > 0 {
			tokensPerSec = float64(genResp.EvalCount) / (float64(genResp.EvalDuration) / 1e9)
		}
		h.metrics.RecordTokens(model, genResp.PromptEvalCount, genResp.EvalCount, tokensPerSec)
	}

	// Record request metrics
	duration := time.Since(start)
	h.metrics.RecordRequest(c.Request.Method, c.Request.URL.Path, model, strconv.Itoa(resp.StatusCode), duration)

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Write response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// HandleChat handles the /api/chat endpoint
func (h *ProxyHandler) HandleChat(c *gin.Context) {
	start := time.Now()
	model := "unknown"

	// Extract priority from header (default to normal)
	priority := queue.PriorityNormal
	if priorityHeader := c.GetHeader("X-Priority"); priorityHeader == "high" {
		priority = queue.PriorityHigh
	}

	// Read request body
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
		return
	}

	// Parse request to extract model
	var req models.ChatRequest
	if err := json.Unmarshal(body, &req); err == nil {
		model = req.Model
	}

	// Submit to queue with priority
	err = h.queue.Submit(c.Request.Context(), model, priority, func() error {
		// Track active requests
		h.metrics.IncActiveRequests(model)
		defer h.metrics.DecActiveRequests(model)

		// Create request to Ollama
		targetURL := fmt.Sprintf("%s%s", h.config.OllamaURL(), c.Request.URL.Path)
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(body))
		if err != nil {
			h.metrics.RecordError(model, "create_request")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return err
		}

		// Copy headers
		for key, values := range c.Request.Header {
			for _, value := range values {
				proxyReq.Header.Add(key, value)
			}
		}

		// Make request
		resp, err := h.httpClient.Do(proxyReq)
		if err != nil {
			h.metrics.RecordError(model, "proxy_request")
			c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to proxy request"})
			return err
		}
		defer resp.Body.Close()

		// Handle streaming vs non-streaming
		if req.Stream {
			h.handleStreamingChatResponse(c, resp, model, start)
		} else {
			h.handleNonStreamingChatResponse(c, resp, model, start)
		}

		return nil
	})

	if err != nil {
		h.metrics.RecordError(model, "queue_error")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
	}
}

func (h *ProxyHandler) handleStreamingChatResponse(c *gin.Context, resp *http.Response, model string, start time.Time) {
	// Set headers for SSE
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Create a scanner to read the response line by line
	scanner := bufio.NewScanner(resp.Body)
	firstTokenTime := time.Time{}
	var totalPromptTokens, totalGeneratedTokens int
	var evalDuration int64

	for scanner.Scan() {
		line := scanner.Bytes()

		// Parse the JSON to extract metrics
		var chunk models.ChatResponse
		if err := json.Unmarshal(line, &chunk); err == nil {
			// Record time to first token
			if firstTokenTime.IsZero() && chunk.Message.Content != "" {
				firstTokenTime = time.Now()
				h.metrics.RecordTimeToFirstToken(model, firstTokenTime.Sub(start))
			}

			// Extract final metrics from done chunk
			if chunk.Done {
				totalPromptTokens = chunk.PromptEvalCount
				totalGeneratedTokens = chunk.EvalCount
				evalDuration = chunk.EvalDuration

				// Record model load time
				if chunk.LoadDuration > 0 {
					h.metrics.RecordModelLoadTime(model, time.Duration(chunk.LoadDuration))
				}
			}
		}

		// Write the chunk to response
		c.Data(http.StatusOK, "application/x-ndjson", line)
		c.Data(http.StatusOK, "application/x-ndjson", []byte("\n"))
		c.Writer.Flush()
	}

	// Record final metrics
	duration := time.Since(start)
	h.metrics.RecordRequest(c.Request.Method, c.Request.URL.Path, model, strconv.Itoa(resp.StatusCode), duration)

	// Record token metrics
	var tokensPerSec float64
	if evalDuration > 0 && totalGeneratedTokens > 0 {
		tokensPerSec = float64(totalGeneratedTokens) / (float64(evalDuration) / 1e9)
	}
	h.metrics.RecordTokens(model, totalPromptTokens, totalGeneratedTokens, tokensPerSec)
}

func (h *ProxyHandler) handleNonStreamingChatResponse(c *gin.Context, resp *http.Response, model string, start time.Time) {
	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_response")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to read response"})
		return
	}

	// Parse response to extract metrics
	var chatResp models.ChatResponse
	if err := json.Unmarshal(body, &chatResp); err == nil {
		// Record model load time
		if chatResp.LoadDuration > 0 {
			h.metrics.RecordModelLoadTime(model, time.Duration(chatResp.LoadDuration))
		}

		// Record token metrics
		var tokensPerSec float64
		if chatResp.EvalDuration > 0 && chatResp.EvalCount > 0 {
			tokensPerSec = float64(chatResp.EvalCount) / (float64(chatResp.EvalDuration) / 1e9)
		}
		h.metrics.RecordTokens(model, chatResp.PromptEvalCount, chatResp.EvalCount, tokensPerSec)
	}

	// Record request metrics
	duration := time.Since(start)
	h.metrics.RecordRequest(c.Request.Method, c.Request.URL.Path, model, strconv.Itoa(resp.StatusCode), duration)

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Write response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

// HandleDefault handles all other requests
func (h *ProxyHandler) HandleDefault(c *gin.Context) {
	start := time.Now()
	model := "unknown"

	// Forward the request as-is
	targetURL := fmt.Sprintf("%s%s", h.config.OllamaURL(), c.Request.URL.Path)

	// Read body if present
	var bodyBytes []byte
	if c.Request.Body != nil {
		bodyBytes, _ = io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	// Create proxy request
	proxyReq, err := http.NewRequest(c.Request.Method, targetURL, bytes.NewReader(bodyBytes))
	if err != nil {
		h.metrics.RecordError(model, "create_request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Make request
	resp, err := h.httpClient.Do(proxyReq)
	if err != nil {
		h.metrics.RecordError(model, "proxy_request")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to proxy request"})
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.metrics.RecordError(model, "read_response")
		c.JSON(http.StatusBadGateway, gin.H{"error": "Failed to read response"})
		return
	}

	// Record metrics
	duration := time.Since(start)
	h.metrics.RecordRequest(c.Request.Method, c.Request.URL.Path, model, strconv.Itoa(resp.StatusCode), duration)

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Write response
	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}