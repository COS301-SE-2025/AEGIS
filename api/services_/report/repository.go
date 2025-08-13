// services/report/repository.go
package report

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// CaseReportsRepoImpl implements the CaseReportsRepo interface for interacting with the DB.
type CaseReportsRepoImpl struct {
	DB *gorm.DB
}

// NewCaseReportsRepo creates a new instance of CaseReportsRepoImpl.
func NewCaseReportsRepo(db *gorm.DB) CaseReportsRepo {
	return &CaseReportsRepoImpl{DB: db}
}

// SaveReport saves a report to the database.
func (repo *CaseReportsRepoImpl) SaveReport(ctx context.Context, report *Report) error {
	return repo.DB.Create(report).Error
}

// GetByID retrieves a report by its ID.
func (repo *CaseReportsRepoImpl) GetByID(ctx context.Context, reportID string) (*CaseReportRow, error) {
	var report CaseReportRow
	err := repo.DB.First(&report, "id = ?", reportID).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

// GetAllReports retrieves all reports from the repository.
func (repo *CaseReportsRepoImpl) GetAllReports(ctx context.Context) ([]CaseReportRow, error) {
	var reports []CaseReportRow
	err := repo.DB.Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// GetReportsByCaseID retrieves all reports for a given case ID.
func (repo *CaseReportsRepoImpl) GetReportsByCaseID(ctx context.Context, caseID string) ([]CaseReportRow, error) {
	var reports []CaseReportRow
	err := repo.DB.Where("case_id = ?", caseID).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// GetReportsByEvidenceID retrieves all reports for a given evidence ID.
func (repo *CaseReportsRepoImpl) GetReportsByEvidenceID(ctx context.Context, evidenceID string) ([]CaseReportRow, error) {
	var reports []CaseReportRow
	err := repo.DB.Where("evidence_id = ?", evidenceID).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// DeleteReportByID deletes a report by its ID
func (repo *CaseReportsRepoImpl) DeleteReportByID(ctx context.Context, reportID string) error {
	// Perform the delete operation using GORM
	if err := repo.DB.WithContext(ctx).Where("id = ?", reportID).Delete(&CaseReportRow{}).Error; err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}
	return nil
}
