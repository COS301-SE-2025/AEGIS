// services/report/ports.go
package report

import (
	coc "aegis-api/services_/chain_of_custody"
	"context"
	"time"
)

// CaseReportsRepo interface defines methods for interacting with the case reports.
type CaseReportsRepo interface {
	SaveReport(ctx context.Context, report *Report) error
	GetByID(ctx context.Context, reportID string) (*CaseReportRow, error)
	GetAllReports(ctx context.Context) ([]CaseReportRow, error)
	GetReportsByCaseID(ctx context.Context, caseID string) ([]CaseReportRow, error)         // New method
	GetReportsByEvidenceID(ctx context.Context, evidenceID string) ([]CaseReportRow, error) // New method
	DeleteReportByID(ctx context.Context, reportID string) error
}

// Add this interface:
type CoCRepo interface {
	ListByCase(ctx context.Context, caseID string) ([]coc.Entry, error)
}

// CaseReportRow represents a row of report data from the database.
type CaseReportRow struct {
	ID                     string
	CaseID                 string
	ExaminerID             string
	Scope                  *string
	Objectives             *string
	Limitations            *string
	ToolsMethods           *string
	FinalConclusion        *string
	EvidenceSummary        *string
	CertificationStatement *string
	DateExamined           *time.Time
	Status                 string
	Version                int
	ReportNumber           *string
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Name                   string // New field for report name
	FilePath               string // New field for the file path
}

// ReportArtifactsRepo handles saving report artifacts (PDF, JSON, CSV).
type ReportArtifactsRepo interface {
	Insert(ctx context.Context, row InsertArtifactRow) (string, error)
}

// InsertArtifactRow represents the data for an artifact associated with a report.
type InsertArtifactRow struct {
	CaseID      string
	ReportID    string
	Format      string // "pdf", "json", "csv"
	StorageRef  string
	SizeBytes   int64
	SHA256      string
	GeneratedBy string
}

// Storage interface for storing report artifacts.
type Storage interface {
	Put(ctx context.Context, path string, data []byte) (storageRef string, size int64, err error)
}

// AuditLogger interface for logging actions related to reports.
type AuditLogger interface {
	LogGenerateReport(ctx context.Context, caseID, reportID, artifactID, actorID, ip, ua string) error
	LogDownloadReport(ctx context.Context, reportID string, userID string, userRole string, userAgent string, ipAddress string, email string, timestamp time.Time) error
	LogDeleteReport(ctx context.Context, reportID string, userID string, userRole string, userAgent string, ipAddress string, email string, timestamp time.Time) error
}

// Authorizer interface for checking if a user is authorized to generate reports.
type Authorizer interface {
	CanGenerateReport(ctx context.Context, userID, caseID string) bool
}
