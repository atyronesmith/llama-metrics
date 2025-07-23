package models

// AnalyzedHealth represents comprehensive health with LLM analysis
type AnalyzedHealth struct {
	SystemHealth
	Analysis *LLMAnalysis `json:"llm_analysis,omitempty"`
}

// LLMAnalysis represents the LLM's analysis of health status
type LLMAnalysis struct {
	Available bool                   `json:"available"`
	Summary   string                 `json:"summary,omitempty"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Error     string                 `json:"error,omitempty"`
	Timestamp string                 `json:"timestamp"`
}