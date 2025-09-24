package timelineai

import "time"

// SuggestionResponseDTO represents the response for suggestion requests
type SuggestionResponseDTO struct {
	Success     bool     `json:"success"`
	Suggestions []string `json:"suggestions"`
	Confidence  float64  `json:"confidence"`
	Message     string   `json:"message,omitempty"`
}

// SeverityRecommendationDTO represents severity recommendation response
type SeverityRecommendationDTO struct {
	Success             bool    `json:"success"`
	RecommendedSeverity string  `json:"recommended_severity"`
	Confidence          float64 `json:"confidence"`
	Message             string  `json:"message,omitempty"`
}

// TagSuggestionDTO represents tag suggestion response
type TagSuggestionDTO struct {
	Success bool     `json:"success"`
	Tags    []string `json:"tags"`
	Message string   `json:"message,omitempty"`
}

// IOCExtractionDTO represents IOC extraction response
type IOCExtractionDTO struct {
	Success bool            `json:"success"`
	IOCs    []IOCExtraction `json:"iocs"`
	Count   int             `json:"count"`
	Message string          `json:"message,omitempty"`
}

// ContextAnalysisDTO represents comprehensive event analysis response
type ContextAnalysisDTO struct {
	Success             bool            `json:"success"`
	RecommendedSeverity string          `json:"recommended_severity,omitempty"`
	SuggestedTags       []string        `json:"suggested_tags,omitempty"`
	ExtractedIOCs       []IOCExtraction `json:"extracted_iocs,omitempty"`
	Suggestions         []string        `json:"suggestions,omitempty"`
	Confidence          float64         `json:"confidence"`
	Message             string          `json:"message,omitempty"`
}

// CaseAnalysisDTO represents case progress analysis response
type CaseAnalysisDTO struct {
	Success            bool      `json:"success"`
	CaseID             string    `json:"case_id"`
	CompletionScore    float64   `json:"completion_score"`
	MissingSteps       []string  `json:"missing_steps"`
	RecommendedActions []string  `json:"recommended_actions"`
	RiskAssessment     string    `json:"risk_assessment"`
	AnalyzedAt         time.Time `json:"analyzed_at"`
	Message            string    `json:"message,omitempty"`
}

// ModelStatusDTO represents AI model status response
type ModelStatusDTO struct {
	Success      bool      `json:"success"`
	ModelName    string    `json:"model_name"`
	Status       string    `json:"status"`
	LastChecked  time.Time `json:"last_checked"`
	ResponseTime int64     `json:"response_time_ms"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// FeedbackRequestDTO represents feedback submission request
type FeedbackRequestDTO struct {
	AnalysisID   string `json:"analysis_id" validate:"required"`
	FeedbackType string `json:"feedback_type" validate:"required,oneof=helpful not_helpful incorrect"`
	Comments     string `json:"comments,omitempty"`
}

// ConfigUpdateDTO represents AI model configuration update
type ConfigUpdateDTO struct {
	ModelName     string            `json:"model_name,omitempty"`
	BaseURL       string            `json:"base_url,omitempty"`
	Temperature   float64           `json:"temperature,omitempty"`
	MaxTokens     int               `json:"max_tokens,omitempty"`
	CustomPrompts map[string]string `json:"custom_prompts,omitempty"`
	Enabled       *bool             `json:"enabled,omitempty"`
}

// NextStepsRequestDTO represents request for next step suggestions
type NextStepsRequestDTO struct {
	CaseID string `json:"case_id" validate:"required"`
}

// EvidenceCorrelationRequestDTO represents evidence correlation request
type EvidenceCorrelationRequestDTO struct {
	CaseID           string `json:"case_id" validate:"required"`
	EventDescription string `json:"event_description" validate:"required"`
}

// EvidenceCorrelationDTO represents evidence correlation response
type EvidenceCorrelationDTO struct {
	Success            bool     `json:"success"`
	CorrelatedEvidence []string `json:"correlated_evidence"`
	Message            string   `json:"message,omitempty"`
}
