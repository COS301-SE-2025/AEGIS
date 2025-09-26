package case_deletion

import (
	"gorm.io/gorm"
)

func NewGormCaseDeletionRepository(db *gorm.DB) *GormCaseRepository {
	return &GormCaseRepository{db: db}
}

// Use this for DI in main.go
