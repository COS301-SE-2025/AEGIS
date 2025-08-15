package report

import (
	"context"
	"time"

	"errors"

	"github.com/google/uuid"
)

var (
	ErrReportNotFound      = errors.New("report not found")
	ErrMongoReportNotFound = errors.New("mongo report not found")
	ErrSectionNotFound     = errors.New("section not found")
	ErrInvalidInput        = errors.New("invalid input")
)

// Report represents the case report structure.
//
//	type Report struct {
//		ID                     uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"` // Unique ID for the report
//		CaseID                 uuid.UUID `gorm:"type:uuid;not null"`                              // ID of the associated case
//		ExaminerID             uuid.UUID `gorm:"type:uuid;not null"`                              // ID of the examiner
//		Scope                  string    `gorm:"type:text"`                                       // Scope of the report
//		Objectives             string    `gorm:"type:text"`                                       // Objectives of the report
//		Limitations            string    `gorm:"type:text"`                                       // Limitations of the report
//		ToolsMethods           string    `gorm:"type:text"`                                       // Tools and methods used in the report
//		FinalConclusion        string    `gorm:"type:text"`                                       // Final conclusion of the report
//		EvidenceSummary        string    `gorm:"type:text"`                                       // Summary of the evidence in the report
//		CertificationStatement string    `gorm:"type:text"`                                       // Certification statement of the report
//		DateExamined           time.Time `gorm:"type:date"`                                       // Date the report was examined
//		Status                 string    `gorm:"type:report_status;default:'draft'"`              // Status of the report (e.g., draft, published)
//		Version                int       `gorm:"not null;default:1"`                              // Version of the report
//		ReportNumber           string    `gorm:"unique"`                                          // Unique report number
//		CreatedAt              time.Time `gorm:"type:timestamp;default:current_timestamp"`        // Timestamp when the report was created
//		UpdatedAt              time.Time `gorm:"type:timestamp;default:current_timestamp"`        // Timestamp when the report was last updated
//		Name                   string    `gorm:"type:varchar(255);not null"`                      // Name of the report (for download purposes)
//		FilePath               string    `gorm:"type:varchar(255);not null"`                      // File path where the report is stored
//	}
type Report struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CaseID       uuid.UUID `gorm:"type:uuid;not null" json:"case_id"`
	ExaminerID   uuid.UUID `gorm:"type:uuid;not null" json:"examiner_id"`
	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	MongoID      string    `gorm:"type:char(24)" json:"mongo_id"` // MongoDB ObjectID as 24-char hex
	ReportNumber string    `gorm:"type:varchar(255);unique" json:"report_number"`
	Status       string    `gorm:"type:report_status;default:'draft'" json:"status"`
	Version      int       `gorm:"not null;default:1" json:"version"`
	DateExamined time.Time `gorm:"type:date" json:"date_examined"`
	FilePath     string    `gorm:"type:varchar(255);not null" json:"file_path"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:current_timestamp" json:"updated_at"`
}

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
	DownloadReport(ctx context.Context, reportID uuid.UUID) (*Report, error)
}
type ReportWithDetails struct {
	ID            uuid.UUID `json:"id"`
	CaseID        uuid.UUID `json:"case_id"`
	Name          string    `json:"name"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	Version       int       `json:"version"`
	LastModified  string    `json:"last_modified"`
	FilePath      string    `json:"file_path"`
	Author        string    `json:"author"`        // full name of examiner
	Collaborators int       `json:"collaborators"` // count from case_user_roles
}
