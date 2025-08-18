package update_status

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ReportStatusRepository interface {
	UpdateReportStatus(ctx context.Context, reportID uuid.UUID, status ReportStatus) (*Report, error)
}

type ReportsRepoStatusImpl struct {
	DB *gorm.DB
}

func NewReportStatusRepository(db *gorm.DB) ReportStatusRepository {
	return &ReportsRepoStatusImpl{DB: db}

}

// compile-time check
var _ ReportStatusRepository = (*ReportsRepoStatusImpl)(nil)


func (r *ReportsRepoStatusImpl) UpdateReportStatus(ctx context.Context, reportID uuid.UUID, status ReportStatus) (*Report, error) {
	res := r.DB.WithContext(ctx).Exec(`
        UPDATE reports
           SET status = ?, version = COALESCE(version, 0) + 1, updated_at = NOW()
         WHERE id = ?
    `, status, reportID)

	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		return nil, errors.New("report not found")
	}

	out := Report{}
	if err := r.DB.WithContext(ctx).First(&out, "id = ?", reportID).Error; err != nil {
		return nil, err
	}
	return &out, nil
}
