package report

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Report represents the case report structure.
type Report struct {
	ID                     uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"` // Unique ID for the report
	CaseID                 uuid.UUID `gorm:"type:uuid;not null"`                              // ID of the associated case
	ExaminerID             uuid.UUID `gorm:"type:uuid;not null"`                              // ID of the examiner
	Scope                  string    `gorm:"type:text"`                                       // Scope of the report
	Objectives             string    `gorm:"type:text"`                                       // Objectives of the report
	Limitations            string    `gorm:"type:text"`                                       // Limitations of the report
	ToolsMethods           string    `gorm:"type:text"`                                       // Tools and methods used in the report
	FinalConclusion        string    `gorm:"type:text"`                                       // Final conclusion of the report
	EvidenceSummary        string    `gorm:"type:text"`                                       // Summary of the evidence in the report
	CertificationStatement string    `gorm:"type:text"`                                       // Certification statement of the report
	DateExamined           time.Time `gorm:"type:date"`                                       // Date the report was examined
	Status                 string    `gorm:"type:report_status;default:'draft'"`              // Status of the report (e.g., draft, published)
	Version                int       `gorm:"not null;default:1"`                              // Version of the report
	ReportNumber           string    `gorm:"unique"`                                          // Unique report number
	CreatedAt              time.Time `gorm:"type:timestamp;default:current_timestamp"`        // Timestamp when the report was created
	UpdatedAt              time.Time `gorm:"type:timestamp;default:current_timestamp"`        // Timestamp when the report was last updated
	Name                   string    `gorm:"type:varchar(255);not null"`                      // Name of the report (for download purposes)
	FilePath               string    `gorm:"type:varchar(255);not null"`                      // File path where the report is stored
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
	DeleteReportByID(ctx context.Context, reportID string) error
}
