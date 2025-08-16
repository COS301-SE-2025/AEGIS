package evidencecount

import "gorm.io/gorm"

type evidenceRepository struct {
	db *gorm.DB
}

func NewEvidenceRepository(db *gorm.DB) *evidenceRepository {
	return &evidenceRepository{db: db}
}

func (r *evidenceRepository) GetEvidenceCountByTenantID(tenantID string) (int64, error) {
	var count int64
	err := r.db.Table("evidence").Where("tenant_id = ?", tenantID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
