package case_creation

import (
	//"aegis-api/db"
	"gorm.io/gorm"
)

type GormCaseRepository struct {
	db *gorm.DB
}

func NewGormCaseRepository(db *gorm.DB) *GormCaseRepository {
	return &GormCaseRepository{db: db}
}

func (r *GormCaseRepository) CreateCase(c *Case) error {
	return r.db.Create(c).Error
}
