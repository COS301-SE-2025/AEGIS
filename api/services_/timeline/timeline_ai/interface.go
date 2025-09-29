package timelineai

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AIService defines the interface for AI assistance functionality
type AIService interface {
	// Text completion and suggestions
	GetEventSuggestions(ctx context.Context, req *SuggestionRequest) (*AIAnalysisResult, error)
	GetSeverityRecommendation(ctx context.Context, description string) (string, float64, error)
	GetTagSuggestions(ctx context.Context, description string) ([]string, error)

	// IOC extraction and analysis
	ExtractIOCs(ctx context.Context, text string) ([]IOCExtraction, error)
	AnalyzeEventContext(ctx context.Context, caseID string, eventText string) (*AIAnalysisResult, error)

	// Investigation assistance
	SuggestNextSteps(ctx context.Context, caseID string) ([]string, error)
	AnalyzeCaseProgress(ctx context.Context, caseID string) (*CaseAnalysis, error)

	// Evidence correlation
	CorrelateEvidence(ctx context.Context, caseID string, eventDescription string) ([]string, error)

	// Learning and improvement
	RecordFeedback(ctx context.Context, analysisID string, feedback *AIFeedback) error

	// Configuration
	UpdateModelConfig(ctx context.Context, config *AIModelConfig) error
	GetModelStatus(ctx context.Context) (*ModelStatus, error)
}

// AIRepository defines the interface for AI data persistence
type AIRepository interface {
	SaveAnalysis(ctx context.Context, analysis *AIAnalysisResult) error
	GetAnalysisHistory(ctx context.Context, caseID string, analysisType string) ([]*AIAnalysisResult, error)
	SaveFeedback(ctx context.Context, feedback *AIFeedback) error
	GetSuggestionPatterns(ctx context.Context, caseType string) (*PredefinedSuggestions, error)
	UpdateSuggestionPatterns(ctx context.Context, patterns *PredefinedSuggestions) error
}

// Additional types for interface
type CaseAnalysis struct {
	CaseID             string    `json:"case_id"`
	CompletionScore    float64   `json:"completion_score"`
	MissingSteps       []string  `json:"missing_steps"`
	RecommendedActions []string  `json:"recommended_actions"`
	RiskAssessment     string    `json:"risk_assessment"`
	SimilarCases       []string  `json:"similar_cases"`
	AnalyzedAt         time.Time `json:"analyzed_at"`
}

type AIFeedback struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	AnalysisID   string             `bson:"analysis_id" json:"analysis_id"`
	UserID       string             `bson:"user_id" json:"user_id"`
	FeedbackType string             `bson:"feedback_type" json:"feedback_type"` // "helpful", "not_helpful", "incorrect"
	Comments     string             `bson:"comments,omitempty" json:"comments,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}

type ModelStatus struct {
	ModelName    string    `json:"model_name"`
	Status       string    `json:"status"` // "online", "offline", "loading"
	LastChecked  time.Time `json:"last_checked"`
	ResponseTime int64     `json:"response_time_ms"`
	ErrorMessage string    `json:"error_message,omitempty"`
}
