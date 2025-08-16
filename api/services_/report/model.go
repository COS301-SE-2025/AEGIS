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

type Report struct {
	ID         uuid.UUID `gorm:"primaryKey;type:uuid;default:uuid_generate_v4()" json:"id"`
	CaseID     uuid.UUID `gorm:"type:uuid;not null" json:"case_id"`
	ExaminerID uuid.UUID `gorm:"type:uuid;not null" json:"examiner_id"`
	// NEW ↓↓↓
	TenantID uuid.UUID `gorm:"type:uuid;not null;index:idx_reports_tenant_team_updated,priority:1" json:"tenant_id"`
	TeamID   uuid.UUID `gorm:"type:uuid;not null;index:idx_reports_tenant_team_updated,priority:2" json:"team_id"`
	// NEW ↑↑↑

	Name         string    `gorm:"type:varchar(255);not null" json:"name"`
	MongoID      string    `gorm:"type:char(24)" json:"mongo_id"`
	ReportNumber string    `gorm:"type:varchar(255);unique" json:"report_number"`
	Status       string    `gorm:"type:report_status;default:'draft'" json:"status"`
	Version      int       `gorm:"not null;default:1" json:"version"`
	DateExamined time.Time `gorm:"type:date" json:"date_examined"`
	FilePath     string    `gorm:"type:varchar(255);not null" json:"file_path"`
	CreatedAt    time.Time `gorm:"type:timestamp;default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `gorm:"type:timestamp;default:current_timestamp;index:idx_reports_tenant_team_updated,priority:3" json:"updated_at"`
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
