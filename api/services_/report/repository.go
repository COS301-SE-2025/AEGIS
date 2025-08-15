package report

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportRepository interface {
	SaveReport(ctx context.Context, report *Report) error
	GetByID(ctx context.Context, reportID uuid.UUID) (*Report, error)
	GetAllReports(ctx context.Context) ([]Report, error)
	GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]ReportWithDetails, error)
	GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error)
	DeleteReportByID(ctx context.Context, reportID uuid.UUID) error
	DownloadReport(ctx context.Context, reportID uuid.UUID) (*Report, error)

	ListRecentCandidates(ctx context.Context, opts RecentReportsOptions, candidateLimit int) ([]Report, error)
}

type ReportsRepoImpl struct {
	DB *gorm.DB
}

func NewReportRepository(db *gorm.DB) ReportRepository {
	return &ReportsRepoImpl{DB: db}

}

// compile-time check
var _ ReportRepository = (*ReportsRepoImpl)(nil)

func (repo *ReportsRepoImpl) ListRecentCandidates(
	ctx context.Context,
	opts RecentReportsOptions,
	candidateLimit int,
) ([]Report, error) {
	if candidateLimit <= 0 || candidateLimit > 200 {
		candidateLimit = 60
	}

	q := repo.DB.WithContext(ctx).Model(&Report{})

	// Filters
	if opts.MineOnly && opts.ExaminerID != uuid.Nil {
		q = q.Where("examiner_id = ?", opts.ExaminerID)
	}
	if opts.CaseID != nil {
		q = q.Where("case_id = ?", *opts.CaseID)
	}
	if opts.Status != nil && strings.TrimSpace(*opts.Status) != "" {
		q = q.Where("status = ?", *opts.Status)
	}

	// Select only what we need for “recent”
	var rows []Report
	if err := q.
		Select("id, case_id, examiner_id, name, status, updated_at").
		Order("updated_at DESC").
		Limit(candidateLimit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (repo *ReportsRepoImpl) SaveReport(ctx context.Context, report *Report) error {
	return repo.DB.WithContext(ctx).Create(report).Error
}

func (repo *ReportsRepoImpl) GetByID(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	var report Report
	err := repo.DB.WithContext(ctx).First(&report, "id = ?", reportID).Error
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func (repo *ReportsRepoImpl) GetAllReports(ctx context.Context) ([]Report, error) {
	var reports []Report
	err := repo.DB.WithContext(ctx).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

// Repository layer: returns raw time.Time
func (repo *ReportsRepoImpl) GetReportsByCaseID(ctx context.Context, caseID uuid.UUID) ([]ReportWithDetails, error) {
	reports := []ReportWithDetails{}
	err := repo.DB.Raw(`
        SELECT r.id, r.case_id, r.name, r.status, r.version, r.updated_at as last_modified,
               r.file_path, u.full_name as author,
               (SELECT COUNT(*) FROM case_user_roles cur WHERE cur.case_id = r.case_id) as collaborators
        FROM reports r
        JOIN users u ON r.examiner_id = u.id
        WHERE r.case_id = ?
    `, caseID).Scan(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (repo *ReportsRepoImpl) GetReportsByEvidenceID(ctx context.Context, evidenceID uuid.UUID) ([]Report, error) {
	var reports []Report
	err := repo.DB.WithContext(ctx).Where("evidence_id = ?", evidenceID).Find(&reports).Error
	if err != nil {
		return nil, err
	}
	return reports, nil
}

func (repo *ReportsRepoImpl) DeleteReportByID(ctx context.Context, reportID uuid.UUID) error {
	if err := repo.DB.WithContext(ctx).Delete(&Report{}, "id = ?", reportID).Error; err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}
	return nil
}

func (repo *ReportsRepoImpl) DownloadReport(ctx context.Context, reportID uuid.UUID) (*Report, error) {
	var report Report
	if err := repo.DB.WithContext(ctx).First(&report, "id = ?", reportID).Error; err != nil {
		return nil, err
	}
	return &report, nil
}
