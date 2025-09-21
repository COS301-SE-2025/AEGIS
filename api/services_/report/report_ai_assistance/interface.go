package report_ai_assistance

import (
	graphicalmapping "aegis-api/services_/GraphicalMapping"
	"aegis-api/services_/case/case_creation"
	"aegis-api/services_/evidence/metadata"
	reportshared "aegis-api/services_/report/shared"
	"aegis-api/services_/timeline"
	"context"
)

type AISuggestionRepository interface {
	GetSuggestionByID(ctx context.Context, id string) (*AISuggestion, error)
	CreateSuggestion(ctx context.Context, suggestion *AISuggestion) error
	ListSuggestionsBySection(ctx context.Context, sectionID string) ([]*AISuggestion, error)
}

type SectionRefsRepository interface {
	GetRefsBySection(ctx context.Context, sectionID string) ([]*SectionRef, error)
	CreateRef(ctx context.Context, ref *SectionRef) error
}

type AIFeedbackRepository interface {
	SubmitFeedback(ctx context.Context, feedback *AIFeedback) error
	ListFeedbackBySuggestion(ctx context.Context, suggestionID string) ([]*AIFeedback, error)
}

// AIClient defines methods the AI engine should implement
type AIClient interface {
	GenerateSuggestion(ctx context.Context, input AISuggestionInput) (string, error)
	RefineSuggestion(ctx context.Context, content, feedback string) (string, error)
	SummarizeEvidence(ctx context.Context, evidence []metadata.Evidence, iocs []graphicalmapping.IOC, timeline []timeline.TimelineEvent) (string, error)
	GenerateSectionReferences(ctx context.Context, sectionName string, report *reportshared.Report) ([]string, error)
	GenerateRecommendations(ctx context.Context, caseData *case_creation.Case, analysisSummary string) (string, error)
}

type ReportAIService interface {
	GenerateSectionSuggestion(ctx context.Context, reportIDHex string, sectionIDHex string) (*AISuggestion, error)
	SaveFeedback(ctx context.Context, feedback AIFeedback) error
	GenerateSectionReferences(ctx context.Context, sectionIDHex string, report *reportshared.Report) ([]string, error)
	EnhanceSummary(ctx context.Context, payload map[string]interface{}) (string, error)
}
