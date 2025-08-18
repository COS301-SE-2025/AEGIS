package report

import (
	"context"
	"errors"
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
	UpdateReportName(ctx context.Context, reportID uuid.UUID, name string) (*Report, error)
	ListRecentCandidates(ctx context.Context, opts RecentReportsOptions, candidateLimit int) ([]Report, error)
	GetReportsByTeamID(ctx context.Context, tenantID, teamID uuid.UUID) ([]ReportWithDetails, error)
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

// services_/report/repository.go
func (repo *ReportsRepoImpl) GetReportsByTeamID(
	ctx context.Context,
	tenantID, teamID uuid.UUID,
) ([]ReportWithDetails, error) {
	var out []ReportWithDetails
	err := repo.DB.WithContext(ctx).Raw(`
        SELECT
            r.id,
            r.case_id,
            r.team_id,
            r.name,
            ''::text AS type, -- fallback: you can switch to r.type when column exists
            r.status,
            r.version,
            to_char(r.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS last_modified,
            r.file_path,
            COALESCE(u.full_name, u.email) AS author,
            (SELECT COUNT(*) FROM case_user_roles cur WHERE cur.case_id = r.case_id) AS collaborators,
            c.title      AS case_name,  -- from cases
            c.team_name  AS team_name   -- from cases
        FROM reports r
        JOIN users u  ON r.examiner_id = u.id
        JOIN cases c  ON r.case_id     = c.id
                     AND c.tenant_id   = r.tenant_id   -- tenant safety
        WHERE r.tenant_id = ? AND r.team_id = ?
        ORDER BY r.updated_at DESC
    `, tenantID, teamID).Scan(&out).Error
	return out, err
}

// Repository layer: returns raw time.Time
func (repo *ReportsRepoImpl) GetReportsByCaseID(
	ctx context.Context,
	caseID uuid.UUID,
) ([]ReportWithDetails, error) {
	var reports []ReportWithDetails
	err := repo.DB.WithContext(ctx).Raw(`
        SELECT
            r.id,
            r.case_id,
            r.team_id,
            r.name,
            ''::text AS type, -- fallback
            r.status,
            r.version,
            to_char(r.updated_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS last_modified,
            r.file_path,
            COALESCE(u.full_name, u.email) AS author,
            (SELECT COUNT(*) FROM case_user_roles cur WHERE cur.case_id = r.case_id) AS collaborators,
            c.title     AS case_name,
            c.team_name AS team_name
        FROM reports r
        JOIN users u ON r.examiner_id = u.id
        JOIN cases c ON r.case_id     = c.id
                    AND c.tenant_id   = r.tenant_id
        WHERE r.case_id = ?
        ORDER BY r.updated_at DESC
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

func (r *ReportsRepoImpl) UpdateReportName(ctx context.Context, reportID uuid.UUID, name string) (*Report, error) {
	// Bump version + touch updated_at. NOW() is Postgres/MySQL; switch to CURRENT_TIMESTAMP if you prefer.
	res := r.DB.WithContext(ctx).Exec(`
        UPDATE reports
           SET name = ?, version = COALESCE(version, 0) + 1, updated_at = NOW()
         WHERE id = ?
    `, name, reportID)
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("not found")
	}

	// Read back the full row into your Report struct
	out := Report{}
	if err := r.DB.WithContext(ctx).First(&out, "id = ?", reportID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}
