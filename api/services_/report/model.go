package report

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Report represents the case report structure.
type Report struct {
	ID                     uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	CaseID                 uuid.UUID `gorm:"type:uuid;not null"`
	ExaminerID             uuid.UUID `gorm:"type:uuid;not null"`
	Scope                  string    `gorm:"type:text"`
	Objectives             string    `gorm:"type:text"`
	Limitations            string    `gorm:"type:text"`
	ToolsMethods           string    `gorm:"type:text"`
	FinalConclusion        string    `gorm:"type:text"`
	EvidenceSummary        string    `gorm:"type:text"`
	CertificationStatement string    `gorm:"type:text"`
	DateExamined           time.Time `gorm:"type:date"`
	Status                 string    `gorm:"type:report_status;default:'draft'"`
	Version                int       `gorm:"not null;default:1"`
	ReportNumber           string    `gorm:"unique"`
	CreatedAt              time.Time `gorm:"type:timestamp;default:current_timestamp"`
	UpdatedAt              time.Time `gorm:"type:timestamp;default:current_timestamp"`
}

// ReportInterface defines the methods for managing reports.
// report/ports.go

// ReportInterface defines the methods for managing reports.
type ReportInterface interface {
	GenerateReport(ctx context.Context, caseID uuid.UUID, examinerID uuid.UUID) (*Report, error)
	SaveReport(ctx context.Context, report *Report) error
	GetReportByID(ctx context.Context, reportID uuid.UUID) (*Report, error)
	UpdateReport(ctx context.Context, report *Report) error
	GetAllReports(ctx context.Context) ([]Report, error)
	GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]Report, error)         // Updated to use uuid.UUID
	GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error) // Updated to use uuid.UUID
}
