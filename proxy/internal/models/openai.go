package models

import "time"

// OpenAI Chat Completion Request/Response

// ChatCompletionRequest represents an OpenAI chat completion request
type ChatCompletionRequest struct {
	Model            string                 `json:"model"`
	Messages         []ChatMessage          `json:"messages"`
	Temperature      float64                `json:"temperature,omitempty"`
	TopP             float64                `json:"top_p,omitempty"`
	N                int                    `json:"n,omitempty"`
	Stream           bool                   `json:"stream,omitempty"`
	Stop             interface{}            `json:"stop,omitempty"`
	MaxTokens        int                    `json:"max_tokens,omitempty"`
	PresencePenalty  float64                `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64                `json:"frequency_penalty,omitempty"`
	LogitBias        map[string]float64     `json:"logit_bias,omitempty"`
	User             string                 `json:"user,omitempty"`
	ResponseFormat   *ResponseFormat        `json:"response_format,omitempty"`
	Seed             int                    `json:"seed,omitempty"`
	Tools            []Tool                 `json:"tools,omitempty"`
	ToolChoice       interface{}            `json:"tool_choice,omitempty"`
	Functions        []Function             `json:"functions,omitempty"` // Deprecated
	FunctionCall     interface{}            `json:"function_call,omitempty"` // Deprecated
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	Role         string       `json:"role"`
	Content      string       `json:"content"`
	Name         string       `json:"name,omitempty"`
	ToolCalls    []ToolCall   `json:"tool_calls,omitempty"`
	ToolCallID   string       `json:"tool_call_id,omitempty"`
	FunctionCall *FunctionCall `json:"function_call,omitempty"` // Deprecated
}

// ChatCompletionResponse represents an OpenAI chat completion response
type ChatCompletionResponse struct {
	ID                string               `json:"id"`
	Object            string               `json:"object"`
	Created           int64                `json:"created"`
	Model             string               `json:"model"`
	Choices           []ChatChoice         `json:"choices"`
	Usage             *Usage               `json:"usage,omitempty"`
	SystemFingerprint string               `json:"system_fingerprint,omitempty"`
}

// ChatChoice represents a choice in a chat completion response
type ChatChoice struct {
	Index        int          `json:"index"`
	Message      ChatMessage  `json:"message"`
	Delta        *ChatMessage `json:"delta,omitempty"` // For streaming
	FinishReason string       `json:"finish_reason,omitempty"`
	LogProbs     *LogProbs    `json:"logprobs,omitempty"`
}

// StreamingChatCompletionResponse represents a streaming response chunk
type StreamingChatCompletionResponse struct {
	ID                string       `json:"id"`
	Object            string       `json:"object"`
	Created           int64        `json:"created"`
	Model             string       `json:"model"`
	Choices           []ChatChoice `json:"choices"`
	SystemFingerprint string       `json:"system_fingerprint,omitempty"`
}

// Completion API (legacy)

// CompletionRequest represents an OpenAI completion request
type CompletionRequest struct {
	Model            string             `json:"model"`
	Prompt           interface{}        `json:"prompt"`
	Suffix           string             `json:"suffix,omitempty"`
	MaxTokens        int                `json:"max_tokens,omitempty"`
	Temperature      float64            `json:"temperature,omitempty"`
	TopP             float64            `json:"top_p,omitempty"`
	N                int                `json:"n,omitempty"`
	Stream           bool               `json:"stream,omitempty"`
	LogProbs         int                `json:"logprobs,omitempty"`
	Echo             bool               `json:"echo,omitempty"`
	Stop             interface{}        `json:"stop,omitempty"`
	PresencePenalty  float64            `json:"presence_penalty,omitempty"`
	FrequencyPenalty float64            `json:"frequency_penalty,omitempty"`
	BestOf           int                `json:"best_of,omitempty"`
	LogitBias        map[string]float64 `json:"logit_bias,omitempty"`
	User             string             `json:"user,omitempty"`
}

// CompletionResponse represents an OpenAI completion response
type CompletionResponse struct {
	ID      string              `json:"id"`
	Object  string              `json:"object"`
	Created int64               `json:"created"`
	Model   string              `json:"model"`
	Choices []CompletionChoice  `json:"choices"`
	Usage   *Usage              `json:"usage,omitempty"`
}

// CompletionChoice represents a choice in a completion response
type CompletionChoice struct {
	Text         string    `json:"text"`
	Index        int       `json:"index"`
	LogProbs     *LogProbs `json:"logprobs,omitempty"`
	FinishReason string    `json:"finish_reason,omitempty"`
}

// Common structures

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Tool represents a function/tool definition
type Tool struct {
	Type     string   `json:"type"`
	Function Function `json:"function"`
}

// Function represents a function definition
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolCall represents a tool call in a message
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ResponseFormat specifies the format of the response
type ResponseFormat struct {
	Type string `json:"type"` // "text" or "json_object"
}

// LogProbs represents log probability information
type LogProbs struct {
	Content []LogProbContent `json:"content,omitempty"`
}

// LogProbContent represents log probability for a token
type LogProbContent struct {
	Token       string            `json:"token"`
	LogProb     float64           `json:"logprob"`
	Bytes       []byte            `json:"bytes,omitempty"`
	TopLogProbs []TopLogProbEntry `json:"top_logprobs,omitempty"`
}

// TopLogProbEntry represents a top log probability entry
type TopLogProbEntry struct {
	Token   string  `json:"token"`
	LogProb float64 `json:"logprob"`
	Bytes   []byte  `json:"bytes,omitempty"`
}

// Error response

// OpenAIError represents an error response
type OpenAIError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Message string  `json:"message"`
	Type    string  `json:"type"`
	Param   string  `json:"param,omitempty"`
	Code    *string `json:"code,omitempty"`
}

// Enhanced metrics tracking

// RequestMetadata stores metadata about a request for tracking
type RequestMetadata struct {
	RequestID        string
	Model            string
	User             string
	StartTime        time.Time
	EndTime          time.Time
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Stream           bool
	StatusCode       int
	Error            string
	Endpoint         string
	Method           string
	ResponseTime     time.Duration
	TimeToFirstToken time.Duration
	TokensPerSecond  float64
}