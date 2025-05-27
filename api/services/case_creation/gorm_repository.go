package case_creation

import (
	"aegis-api/db"
	"gorm.io/gorm"
)

type GormCaseRepository struct {
	db *gorm.DB
}

func NewGormCaseRepository() *GormCaseRepository {
	return &GormCaseRepository{db: db.DB}
}

func (r *GormCaseRepository) CreateCase(c *Case) error {
	return r.db.Create(c).Error
}
