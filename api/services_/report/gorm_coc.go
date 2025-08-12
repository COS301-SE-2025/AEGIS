// services/report/coc_repo_gorm.go
package report

import (
	"context"

	coc "aegis-api/services_/chain_of_custody"

	"gorm.io/gorm"
)

type GormCoCRepo struct {
	DB *gorm.DB
}

func NewCoCRepo(db *gorm.DB) CoCRepo {
	return &GormCoCRepo{DB: db}
}

func (r *GormCoCRepo) ListByCase(ctx context.Context, caseID string) ([]coc.Entry, error) {
	var entries []coc.Entry
	if err := r.DB.WithContext(ctx).
		Where("case_id = ?", caseID).
		Order("occurred_at ASC").
		Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}
