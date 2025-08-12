// report/report_service.go
package report

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ReportServiceImpl implements the ReportService interface for managing reports.
type ReportServiceImpl struct {
	DB *gorm.DB
}

// NewReportServiceImpl creates a new instance of ReportServiceImpl.
func NewReportServiceImpl(db *gorm.DB) ReportInterface {
	return &ReportServiceImpl{DB: db}
}

// GenerateReport generates a new report for a given case and examiner.
func (s *ReportServiceImpl) GenerateReport(ctx context.Context, caseID uuid.UUID, examinerID uuid.UUID) (*Report, error) {
	report := &Report{
		CaseID:     caseID,
		ExaminerID: examinerID,
		Scope:      "Scope of investigation",      // Example value, replace with actual logic
		Objectives: "Objectives of investigation", // Example value, replace with actual logic
		Status:     "draft",
		Version:    1,
	}

	// Save the generated report to the database.
	err := s.DB.Create(report).Error
	if err != nil {
		return nil, err
	}

	return report, nil
}

// SaveReport saves a report to the database.
func (s *ReportServiceImpl) SaveReport(ctx context.Context, report *Report) error {
	return s.DB.Create(report).Error
}

// GetReportByID retrieves a report by its ID.
func (s *ReportServiceImpl) GetReportByID(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	var report Report
	err := s.DB.First(&report, "id = ?", reportID).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// GetAllReports retrieves all reports from the repository.
func (s *ReportServiceImpl) GetAllReports(ctx context.Context) ([]Report, error) {
	var reports []Report
	err := s.DB.Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// UpdateReport updates an existing report.
func (s *ReportServiceImpl) UpdateReport(ctx context.Context, report *Report) error {
	return s.DB.Save(report).Error
}

// GetReportsByCaseID retrieves all reports for a given case ID.
func (s *ReportServiceImpl) GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]Report, error) {
	var reports []Report
	err := s.DB.Where("case_id = ?", caseID).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// GetReportsByEvidenceID retrieves all reports for a given evidence ID.
func (s *ReportServiceImpl) GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error) {
	var reports []Report
	err := s.DB.Where("evidence_id = ?", evidenceID).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}
