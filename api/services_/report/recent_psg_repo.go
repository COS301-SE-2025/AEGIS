// services_/report/repo_pg_recent.go
package report

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormReportRepository struct{ db *gorm.DB }

// compile-time check
//var _ ReportRepository = (*GormReportRepository)(nil)

func (r *GormReportRepository) ListRecentCandidates(
	ctx context.Context,
	opts RecentReportsOptions,
	candidateLimit int,
) ([]Report, error) {
	if candidateLimit <= 0 || candidateLimit > 200 {
		candidateLimit = 60 // grab more than final limit; Mongo timestamps might promote some
	}

	q := r.db.WithContext(ctx).Model(&Report{})

	// 🔒 Multi-tenancy scope
	if opts.TenantID != uuid.Nil {
		q = q.Where("tenant_id = ?", opts.TenantID)
	}
	// If you want “all teams in tenant”, leave TeamID nil; otherwise filter by team
	if opts.TeamID != nil && *opts.TeamID != uuid.Nil {
		q = q.Where("team_id = ?", *opts.TeamID)
	}

	// Additional filters
	if opts.MineOnly && opts.ExaminerID != uuid.Nil {
		q = q.Where("examiner_id = ?", opts.ExaminerID)
	}
	if opts.CaseID != nil {
		q = q.Where("case_id = ?", *opts.CaseID)
	}
	if opts.Status != nil && strings.TrimSpace(*opts.Status) != "" {
		q = q.Where("status = ?", *opts.Status)
	}

	// Select only what we need for recent
	var rows []Report
	if err := q.
		Select("id, case_id, examiner_id, tenant_id, team_id, name, status, updated_at").
		Order("updated_at DESC").
		Limit(candidateLimit).
		Find(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}
