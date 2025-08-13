// services/report/ports.go
package report

import (
	coc "aegis-api/services_/chain_of_custody"
	"context"
	"time"
)

type CoCRepo interface {
	ListByCase(ctx context.Context, caseID string) ([]coc.Entry, error)
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

//perform joint operations on different tables to get fields for the Authors name and the Count of collaborators that are working on a case
