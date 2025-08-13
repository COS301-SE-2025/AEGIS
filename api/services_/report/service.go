// services/report/report_service.go
package report

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ReportService struct contains the repositories and services needed for report management.
type ReportService struct {
	CaseReportsRepo     CaseReportsRepo
	ReportArtifactsRepo ReportArtifactsRepo
	Storage             Storage
	AuditLogger         AuditLogger
	Authorizer          Authorizer
	CoCRepo             CoCRepo
}

// NewReportService creates a new instance of ReportService.
func NewReportService(
	caseReportsRepo CaseReportsRepo,
	reportArtifactsRepo ReportArtifactsRepo,
	storage Storage,
	auditLogger AuditLogger,
	authorizer Authorizer,
	coCRepo CoCRepo, // Added for CoCRepo
) *ReportService {
	return &ReportService{
		CaseReportsRepo:     caseReportsRepo,
		ReportArtifactsRepo: reportArtifactsRepo,
		Storage:             storage,
		AuditLogger:         auditLogger,
		Authorizer:          authorizer,
		CoCRepo:             coCRepo,
	}
}

// GenerateReport creates a new report for a given case ID and examiner ID.
func (s *ReportService) GenerateReport(ctx context.Context, caseID uuid.UUID, examinerID uuid.UUID) (*Report, error) {
	// Fetch chain of custody information for the given case
	cocEntries, err := s.CoCRepo.ListByCase(ctx, caseID.String()) // This uses coCRepo to get CoC data
	if err != nil {
		return nil, err
	}

	// Logic to generate a new report including CoC entries
	report := &Report{
		CaseID:     caseID,
		ExaminerID: examinerID,
		Scope:      "Scope of investigation",
		Objectives: "Objectives of investigation",
		Status:     "draft",
		Version:    1,
	}

	// Example: Concatenate CoC entries into the "EvidenceSummary" field
	var cocSummary string
	for _, entry := range cocEntries {
		cocSummary += *entry.Reason + "\n" // You can format it however you need
	}

	report.EvidenceSummary = cocSummary // Include the CoC information in the report

	// Save the generated report to the database.
	err = s.CaseReportsRepo.SaveReport(ctx, report)
	if err != nil {
		return nil, err
	}

	// Log the report generation activity using the audit logger.
	err = s.AuditLogger.LogGenerateReport(ctx, caseID.String(), report.ID.String(), "artifactID", examinerID.String(), "ip", "userAgent")
	if err != nil {
		return nil, err
	}

	return report, nil
}

// SaveReport saves a report to the database.
func (s *ReportService) SaveReport(ctx context.Context, report *Report) error {
	return s.CaseReportsRepo.SaveReport(ctx, report)
}

// GetReportByID retrieves a report by its ID.
func (s *ReportService) GetReportByID(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	reportRow, err := s.CaseReportsRepo.GetByID(ctx, reportID.String())
	if err != nil {
		return nil, err
	}

	// Map the row from the database to the Report struct.
	report := &Report{
		ID:                     uuid.MustParse(reportRow.ID),
		CaseID:                 uuid.MustParse(reportRow.CaseID),
		ExaminerID:             uuid.MustParse(reportRow.ExaminerID),
		Scope:                  *reportRow.Scope,
		Objectives:             *reportRow.Objectives,
		Limitations:            *reportRow.Limitations,
		ToolsMethods:           *reportRow.ToolsMethods,
		FinalConclusion:        *reportRow.FinalConclusion,
		EvidenceSummary:        *reportRow.EvidenceSummary,
		CertificationStatement: *reportRow.CertificationStatement,
		DateExamined:           *reportRow.DateExamined,
		Status:                 reportRow.Status,
		Version:                reportRow.Version,
		ReportNumber:           *reportRow.ReportNumber,
		CreatedAt:              reportRow.CreatedAt,
		UpdatedAt:              reportRow.UpdatedAt,
	}

	return report, nil
}

// UpdateReport updates an existing report.
func (s *ReportService) UpdateReport(ctx context.Context, report *Report) error {
	return s.CaseReportsRepo.SaveReport(ctx, report)
}

// GetAllReports retrieves all reports from the repository.
func (s *ReportService) GetAllReports(ctx context.Context) ([]Report, error) {
	reportRows, err := s.CaseReportsRepo.GetAllReports(ctx)
	if err != nil {
		return nil, err
	}

	var reports []Report
	// Map each row to a Report struct.
	for _, row := range reportRows {
		reports = append(reports, Report{
			ID:                     uuid.MustParse(row.ID),
			CaseID:                 uuid.MustParse(row.CaseID),
			ExaminerID:             uuid.MustParse(row.ExaminerID),
			Scope:                  *row.Scope,
			Objectives:             *row.Objectives,
			Limitations:            *row.Limitations,
			ToolsMethods:           *row.ToolsMethods,
			FinalConclusion:        *row.FinalConclusion,
			EvidenceSummary:        *row.EvidenceSummary,
			CertificationStatement: *row.CertificationStatement,
			DateExamined:           *row.DateExamined,
			Status:                 row.Status,
			Version:                row.Version,
			ReportNumber:           *row.ReportNumber,
			CreatedAt:              row.CreatedAt,
			UpdatedAt:              row.UpdatedAt,
		})
	}

	return reports, nil
}

// GetReportsByCaseID retrieves all reports for a given case ID.
func (s *ReportService) GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]Report, error) {
	// Fetch all reports associated with the given caseID from the repository.
	reportRows, err := s.CaseReportsRepo.GetReportsByCaseID(ctx, caseID.String()) // Replace with actual repository function
	if err != nil {
		return nil, err
	}

	// Map each row to a Report struct.
	var reports []Report
	for _, row := range reportRows {
		reports = append(reports, Report{
			ID:                     uuid.MustParse(row.ID),
			CaseID:                 uuid.MustParse(row.CaseID),
			ExaminerID:             uuid.MustParse(row.ExaminerID),
			Scope:                  *row.Scope,
			Objectives:             *row.Objectives,
			Limitations:            *row.Limitations,
			ToolsMethods:           *row.ToolsMethods,
			FinalConclusion:        *row.FinalConclusion,
			EvidenceSummary:        *row.EvidenceSummary,
			CertificationStatement: *row.CertificationStatement,
			DateExamined:           *row.DateExamined,
			Status:                 row.Status,
			Version:                row.Version,
			ReportNumber:           *row.ReportNumber,
			CreatedAt:              row.CreatedAt,
			UpdatedAt:              row.UpdatedAt,
		})
	}

	return reports, nil
}

// GetReportsByEvidenceID retrieves all reports for a given evidence ID.
func (s *ReportService) GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error) {
	// Fetch all reports associated with the given evidenceID from the repository.
	reportRows, err := s.CaseReportsRepo.GetReportsByEvidenceID(ctx, evidenceID.String()) // Replace with actual repository function
	if err != nil {
		return nil, err
	}

	// Map each row to a Report struct.
	var reports []Report
	for _, row := range reportRows {
		reports = append(reports, Report{
			ID:                     uuid.MustParse(row.ID),
			CaseID:                 uuid.MustParse(row.CaseID),
			ExaminerID:             uuid.MustParse(row.ExaminerID),
			Scope:                  *row.Scope,
			Objectives:             *row.Objectives,
			Limitations:            *row.Limitations,
			ToolsMethods:           *row.ToolsMethods,
			FinalConclusion:        *row.FinalConclusion,
			EvidenceSummary:        *row.EvidenceSummary,
			CertificationStatement: *row.CertificationStatement,
			DateExamined:           *row.DateExamined,
			Status:                 row.Status,
			Version:                row.Version,
			ReportNumber:           *row.ReportNumber,
			CreatedAt:              row.CreatedAt,
			UpdatedAt:              row.UpdatedAt,
		})
	}

	return reports, nil
}

// DeleteReportByID deletes a report by its ID and logs the action.
func (s *ReportService) DeleteReportByID(ctx context.Context, reportID string) error {
	// Log the delete action
	// You may not need user information here, or you can pass a default context or other parameters if needed
	err := s.AuditLogger.LogDeleteReport(ctx, reportID, "", "", "", "", "", time.Now()) // Assuming default or nil values for logging
	if err != nil {
		return fmt.Errorf("failed to log delete event: %w", err)
	}

	// Call the repository method to delete the report from the database
	err = s.CaseReportsRepo.DeleteReportByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	// Return nil if the report was successfully deleted
	return nil
}
