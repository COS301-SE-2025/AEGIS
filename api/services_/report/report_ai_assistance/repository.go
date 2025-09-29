package report_ai_assistance

import (
	"context"
	"fmt"
	"time"

	reportshared "aegis-api/services_/report/shared"

	"gorm.io/gorm"
)

// --------------------- ReportSectionRepository ---------------------
type GormReportSectionRepo struct {
	db *gorm.DB
}

func NewGormReportSectionRepo(db *gorm.DB) *GormReportSectionRepo {
	return &GormReportSectionRepo{db: db}
}

func (r *GormReportSectionRepo) GetSectionByID(ctx context.Context, id string) (*reportshared.ReportSection, error) {
	var section reportshared.ReportSection
	if err := r.db.WithContext(ctx).First(&section, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &section, nil
}

func (r *GormReportSectionRepo) CreateSection(ctx context.Context, section *reportshared.ReportSection) error {
	section.CreatedAt = time.Now()
	section.UpdatedAt = time.Now()
	res := r.db.WithContext(ctx).Create(section)
	// Log SQL and error
	stmt := res.Statement
	fmt.Printf("[DEBUG] Gorm SQL: %s\n", stmt.SQL.String())
	fmt.Printf("[DEBUG] Gorm Vars: %v\n", stmt.Vars)
	if res.Error != nil {
		fmt.Printf("[ERROR] CreateSection failed: %v\n", res.Error)
	} else {
		fmt.Printf("[DEBUG] CreateSection result: RowsAffected=%d, ID=%s\n", res.RowsAffected, section.ID)
	}
	return res.Error
}

func (r *GormReportSectionRepo) UpdateSection(ctx context.Context, section *reportshared.ReportSection) error {
	section.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(section).Error
}

func (r *GormReportSectionRepo) ListSectionsByReport(ctx context.Context, reportID string) ([]*reportshared.ReportSection, error) {
	var sections []*reportshared.ReportSection
	if err := r.db.WithContext(ctx).Where("report_id = ?", reportID).Find(&sections).Error; err != nil {
		return nil, err
	}
	return sections, nil
}

// --------------------- AISuggestionRepository ---------------------
type GormAISuggestionRepo struct {
	db *gorm.DB
}

func NewGormAISuggestionRepo(db *gorm.DB) *GormAISuggestionRepo {
	return &GormAISuggestionRepo{db: db}
}

func (r *GormAISuggestionRepo) GetSuggestionByID(ctx context.Context, id string) (*AISuggestion, error) {
	var s AISuggestion
	if err := r.db.WithContext(ctx).First(&s, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *GormAISuggestionRepo) CreateSuggestion(ctx context.Context, suggestion *AISuggestion) error {
	suggestion.CreatedAt = time.Now()
	// Check if section exists before inserting suggestion
	var section ReportSection
	err := r.db.WithContext(ctx).Table("report_sections").First(&section, "id = ?", suggestion.SectionID).Error
	if err != nil {
		return fmt.Errorf("cannot create suggestion: section does not exist (id=%s): %w", suggestion.SectionID, err)
	}
	return r.db.WithContext(ctx).Table("report_ai_suggestions").Create(suggestion).Error
}

func (r *GormAISuggestionRepo) ListSuggestionsBySection(ctx context.Context, sectionID string) ([]*AISuggestion, error) {
	var suggestions []*AISuggestion
	if err := r.db.WithContext(ctx).Table("report_ai_suggestions").Where("section_id = ?", sectionID).Find(&suggestions).Error; err != nil {
		return nil, err
	}
	return suggestions, nil
}

// --------------------- SectionRefsRepository ---------------------
type GormSectionRefsRepo struct {
	db *gorm.DB
}

func NewGormSectionRefsRepo(db *gorm.DB) *GormSectionRefsRepo {
	return &GormSectionRefsRepo{db: db}
}

func (r *GormSectionRefsRepo) GetRefsBySection(ctx context.Context, sectionID string) ([]*SectionRef, error) {
	var refs []*SectionRef
	if err := r.db.WithContext(ctx).Where("section_id = ?", sectionID).Find(&refs).Error; err != nil {
		return nil, err
	}
	return refs, nil
}

func (r *GormSectionRefsRepo) CreateRef(ctx context.Context, ref *SectionRef) error {
	return r.db.WithContext(ctx).Create(ref).Error
}

// --------------------- AIFeedbackRepository ---------------------
type GormAIFeedbackRepo struct {
	db *gorm.DB
}

func NewGormAIFeedbackRepo(db *gorm.DB) *GormAIFeedbackRepo {
	return &GormAIFeedbackRepo{db: db}
}

func (r *GormAIFeedbackRepo) SubmitFeedback(ctx context.Context, feedback *AIFeedback) error {
	feedback.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Table("report_ai_feedback").Create(feedback).Error
}

func (r *GormAIFeedbackRepo) ListFeedbackBySuggestion(ctx context.Context, suggestionID string) ([]*AIFeedback, error) {
	var feedbacks []*AIFeedback
	if err := r.db.WithContext(ctx).Table("report_ai_feedback").Where("suggestion_id = ?", suggestionID).Find(&feedbacks).Error; err != nil {
		return nil, err
	}
	return feedbacks, nil
}
