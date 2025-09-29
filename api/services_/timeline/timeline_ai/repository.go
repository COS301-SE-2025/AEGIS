package timelineai

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type aiRepository struct {
	db *gorm.DB
}

// NewAIRepository creates a new AI repository instance
func NewAIRepository(db *gorm.DB) AIRepository {
	return &aiRepository{
		db: db,
	}
}

func (r *aiRepository) SaveAnalysis(ctx context.Context, analysis *AIAnalysisResult) error {
	analysis.CreatedAt = time.Now()
	analysis.ProcessedAt = time.Now()
	return r.db.WithContext(ctx).Create(analysis).Error
}

func (r *aiRepository) GetAnalysisHistory(ctx context.Context, caseID string, analysisType string) ([]*AIAnalysisResult, error) {
	var results []*AIAnalysisResult

	query := r.db.WithContext(ctx).Where("case_id = ?", caseID)
	if analysisType != "" {
		query = query.Where("analysis_type = ?", analysisType)
	}

	err := query.Order("created_at DESC").Limit(50).Find(&results).Error
	return results, err
}

func (r *aiRepository) SaveFeedback(ctx context.Context, feedback *AIFeedback) error {
	feedback.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(feedback).Error
}

func (r *aiRepository) GetSuggestionPatterns(ctx context.Context, caseType string) (*PredefinedSuggestions, error) {
	var patterns PredefinedSuggestions

	// For PostgreSQL, we'll store patterns as a single row with JSONB
	// First try to get patterns for the specific case type
	err := r.db.WithContext(ctx).
		Where("case_type = ?", caseType).
		First(&patterns).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// If not found, try to get default patterns
			err = r.db.WithContext(ctx).
				Where("case_type = ?", "default").
				First(&patterns).Error

			if err == gorm.ErrRecordNotFound {
				// If no patterns exist at all, return default patterns
				return r.getDefaultPatterns(), nil
			} else if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &patterns, nil
}

func (r *aiRepository) UpdateSuggestionPatterns(ctx context.Context, patterns *PredefinedSuggestions) error {

	// Use Upsert (Create or Update) pattern
	return r.db.WithContext(ctx).
		Assign(patterns).             // Set all fields to update
		FirstOrCreate(patterns).Error // Create if doesn't exist
}

func (r *aiRepository) getDefaultPatterns() *PredefinedSuggestions {
	return &PredefinedSuggestions{
		IncidentResponse: []string{
			"Initial incident detection and triage",
			"Containment measures implemented",
			"Evidence preservation initiated",
			"Stakeholder notification completed",
			"Investigation team assembled",
		},
		Malware: []string{
			"Malware sample identified and isolated",
			"Static analysis performed on sample",
			"Dynamic analysis in sandbox environment",
			"IOCs extracted from malware",
			"Anti-virus signatures updated",
		},
		NetworkForensics: []string{
			"Network traffic analysis initiated",
			"Suspicious connections identified",
			"Packet capture analysis completed",
			"Network logs reviewed and analyzed",
			"Lateral movement patterns detected",
		},
		DiskForensics: []string{
			"Disk image acquired and verified",
			"File system analysis performed",
			"Deleted files recovered",
			"Timeline analysis completed",
			"Registry analysis performed",
		},
		Memory: []string{
			"Memory dump acquired",
			"Process analysis completed",
			"Network connections extracted from memory",
			"Injected code detected",
			"Rootkit analysis performed",
		},
		Containment: []string{
			"Affected systems isolated",
			"Network access restricted",
			"User accounts disabled",
			"Malicious processes terminated",
			"Security patches applied",
		},
		Recovery: []string{
			"System restoration initiated",
			"Data integrity verified",
			"Security monitoring enhanced",
			"User access restored",
			"Lessons learned documented",
		},
	}
}
