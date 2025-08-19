package graphicalmapping

import (
	"gorm.io/gorm"
)

type iocRepository struct {
	db *gorm.DB
}

func NewIOCRepository(db *gorm.DB) IOCRepository {
	return &iocRepository{db: db}
}

func (r *iocRepository) Create(ioc *IOC) error {
	return r.db.Create(ioc).Error
}

func (r *iocRepository) GetByID(id string) (*IOC, error) {
	var ioc IOC
	if err := r.db.First(&ioc, id).Error; err != nil {
		return nil, err
	}
	return &ioc, nil
}

func (r *iocRepository) ListByTenant(tenantID string) ([]*IOC, error) {
	var iocs []*IOC
	if err := r.db.Where("tenant_id = ?", tenantID).Find(&iocs).Error; err != nil {
		return nil, err
	}
	return iocs, nil
}

func (r *iocRepository) ListByCase(caseID string) ([]*IOC, error) {
	var iocs []*IOC
	if err := r.db.Where("case_id = ?", caseID).Find(&iocs).Error; err != nil {
		return nil, err
	}
	return iocs, nil
}

func (r *iocRepository) FindSimilar(tenantID string, iocType, value string) ([]*IOC, error) {
	var iocs []*IOC
	if err := r.db.Where("tenant_id = ? AND type = ? AND value = ?", tenantID, iocType, value).Find(&iocs).Error; err != nil {
		return nil, err
	}
	return iocs, nil
}
