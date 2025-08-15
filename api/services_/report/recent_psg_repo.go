// services_/report/repo_pg_recent.go
package report

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormReportRepository struct{ db *gorm.DB }

func (r *GormReportRepository) ListRecentCandidates(ctx context.Context, opts RecentReportsOptions, candidateLimit int) ([]Report, error) {
	if candidateLimit <= 0 || candidateLimit > 200 {
		candidateLimit = 60 // grab a bit more than final limit to avoid missing items promoted by Mongo timestamps
	}

	q := r.db.WithContext(ctx).Table("reports").Where("deleted_at IS NULL")

	if opts.MineOnly && opts.ExaminerID != uuid.Nil {
		q = q.Where("examiner_id = ?", opts.ExaminerID)
	}
	if opts.CaseID != nil {
		q = q.Where("case_id = ?", *opts.CaseID)
	}
	if opts.Status != nil && strings.TrimSpace(*opts.Status) != "" {
		q = q.Where("status = ?", *opts.Status)
	}

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
