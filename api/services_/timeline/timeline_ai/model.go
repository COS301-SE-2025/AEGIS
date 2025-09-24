package timelineai

import (
	"time"
)

// AIAnalysisResult represents the result of AI analysis
type AIAnalysisResult struct {
	ID                  uint            `gorm:"primaryKey" json:"id,omitempty"`
	CaseID              string          `gorm:"index;not null" json:"case_id"`
	EventID             string          `gorm:"index" json:"event_id,omitempty"`
	AnalysisType        string          `gorm:"not null" json:"analysis_type"` // "suggestion", "severity", "tags", "ioc_extraction"
	InputText           string          `gorm:"type:text" json:"input_text"`
	Suggestions         []string        `gorm:"type:text[]" json:"suggestions,omitempty"`
	RecommendedSeverity string          `gorm:"size:20" json:"recommended_severity,omitempty"`
	ExtractedIOCs       []IOCExtraction `gorm:"type:jsonb" json:"extracted_iocs,omitempty"`
	RecommendedTags     []string        `gorm:"type:text[]" json:"recommended_tags,omitempty"`
	Confidence          float64         `gorm:"type:decimal(3,2)" json:"confidence"`
	CreatedAt           time.Time       `json:"created_at"`
	ProcessedAt         time.Time       `json:"processed_at"`
}

// IOCExtraction represents extracted indicators of compromise
type IOCExtraction struct {
	Type       string  `json:"type"` // "ip", "domain", "hash", "url", "email"
	Value      string  `json:"value"`
	Confidence float64 `json:"confidence"`
	Context    string  `json:"context"` // surrounding text
}

// SuggestionRequest represents a request for AI suggestions
type SuggestionRequest struct {
	CaseID         string             `json:"case_id"`
	EventID        string             `json:"event_id,omitempty"`
	InputText      string             `json:"input_text"`
	SuggestionType string             `json:"suggestion_type"` // "completion", "severity", "tags", "next_steps"
	Context        *SuggestionContext `json:"context,omitempty"`
}

// SuggestionContext provides additional context for better suggestions
type SuggestionContext struct {
	ExistingEvents    []TimelineEventContext `json:"existing_events,omitempty"`
	CaseType          string                 `json:"case_type,omitempty"`
	CurrentSeverity   string                 `json:"current_severity,omitempty"`
	AvailableEvidence []string               `json:"available_evidence,omitempty"`
}

// TimelineEventContext simplified timeline event for context
type TimelineEventContext struct {
	Description string    `json:"description"`
	Tags        []string  `json:"tags"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
}

// AIModelConfig represents configuration for AI models
type AIModelConfig struct {
	ModelName     string            `json:"model_name"`
	BaseURL       string            `json:"base_url"`
	Temperature   float64           `json:"temperature"`
	MaxTokens     int               `json:"max_tokens"`
	CustomPrompts map[string]string `json:"custom_prompts"`
	Enabled       bool              `json:"enabled"`
}

// PredefinedSuggestions for common DFIR activities
type PredefinedSuggestions struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	CaseType         string    `gorm:"uniqueIndex;not null" json:"case_type"`
	IncidentResponse []string  `gorm:"type:text[]" json:"incident_response"`
	Malware          []string  `gorm:"type:text[]" json:"malware"`
	NetworkForensics []string  `gorm:"type:text[]" json:"network_forensics"`
	DiskForensics    []string  `gorm:"type:text[]" json:"disk_forensics"`
	Memory           []string  `gorm:"type:text[]" json:"memory"`
	Containment      []string  `gorm:"type:text[]" json:"containment"`
	Recovery         []string  `gorm:"type:text[]" json:"recovery"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
