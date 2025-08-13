package report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReportService defines the business logic for managing reports.
type ReportService interface {
	GenerateReport(ctx context.Context, caseID uuid.UUID, examinerID uuid.UUID) (*Report, error)
	SaveReport(ctx context.Context, report *Report) error
	GetReportByID(ctx context.Context, reportID uuid.UUID) (*Report, error)
	UpdateReport(ctx context.Context, report *Report) error
	GetAllReports(ctx context.Context) ([]Report, error)
	GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]ReportWithDetails, error)
	GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error)
	DeleteReportByID(ctx context.Context, reportID uuid.UUID) error
	DownloadReport(ctx context.Context, reportID uuid.UUID) (*Report, error)
}

// ReportServiceImpl is the concrete implementation of ReportService.
type ReportServiceImpl struct {
	repo ReportRepository
	// artifactsRepo   ReportArtifactsRepository
	storage     Storage
	auditLogger AuditLogger
	authorizer  Authorizer
	coCRepo     GormCoCRepo
}

func NewReportService(
	repo ReportRepository,
	//  artifactsRepo ReportArtifactsRepository,
	storage Storage,
	auditLogger AuditLogger,
	authorizer Authorizer,
	coCRepo GormCoCRepo,
) ReportService {
	return &ReportServiceImpl{
		repo: repo,
		//artifactsRepo: artifactsRepo,
		storage:     storage,
		auditLogger: auditLogger,
		authorizer:  authorizer,
		coCRepo:     coCRepo,
	}
}

// GenerateReport creates a new report for a given case and examiner.
// Here you could include more logic such as fetching case data, formatting content, etc.
func (s *ReportServiceImpl) GenerateReport(ctx context.Context, caseID uuid.UUID, examinerID uuid.UUID) (*Report, error) {
	report := &Report{
		ID:         uuid.New(),
		CaseID:     caseID,
		ExaminerID: examinerID,
		// Add default or generated fields here
	}

	if err := s.repo.SaveReport(ctx, report); err != nil {
		return nil, fmt.Errorf("failed to generate report: %w", err)
	}
	return report, nil
}

// SaveReport persists a report to the repository.
func (s *ReportServiceImpl) SaveReport(ctx context.Context, report *Report) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	return s.repo.SaveReport(ctx, report)
}

// GetReportByID retrieves a report by its ID.
func (s *ReportServiceImpl) GetReportByID(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	return s.repo.GetByID(ctx, reportID)
}

// UpdateReport updates an existing report in the repository.
func (s *ReportServiceImpl) UpdateReport(ctx context.Context, report *Report) error {
	// You could add business logic like checking if the report exists first
	return s.repo.SaveReport(ctx, report) // assuming SaveReport handles both insert/update
}

// GetAllReports retrieves all reports.
func (s *ReportServiceImpl) GetAllReports(ctx context.Context) ([]Report, error) {
	return s.repo.GetAllReports(ctx)
}

// GetReportsByCaseID retrieves all reports for a specific case.
// Service layer: convert timestamps to Africa/Johannesburg
func (s *ReportServiceImpl) GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]ReportWithDetails, error) {
	reports, err := s.repo.GetReportsByCaseID(ctx, caseID)
	if err != nil {
		return nil, err
	}

	// Load timezone once
	loc, _ := time.LoadLocation("Africa/Johannesburg")

	for i := range reports {
		t, err := time.Parse(time.RFC3339, reports[i].LastModified) // or use the actual format your DB returns
		if err != nil {
			continue // or handle error
		}
		reports[i].LastModified = t.In(loc).Format("2006-01-02 15:04:05")
	}

	return reports, nil
}

// GetReportsByEvidenceID retrieves all reports for a specific evidence item.
func (s *ReportServiceImpl) GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error) {
	return s.repo.GetReportsByEvidenceID(ctx, evidenceID)
}

// DeleteReportByID deletes a report by ID.
func (s *ReportServiceImpl) DeleteReportByID(ctx context.Context, reportID uuid.UUID) error {
	return s.repo.DeleteReportByID(ctx, reportID)
}

// DownloadReport fetches the report for downloading.
func (s *ReportServiceImpl) DownloadReport(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	report, err := s.repo.DownloadReport(ctx, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to download report: %w", err)
	}
	return report, nil
}
