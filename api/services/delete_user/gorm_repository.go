package delete_user

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) DeleteUserByID(userID uuid.UUID) error {
	// Delete user where id matches
	result := r.db.Delete(&User{}, "id = ?", userID)
	return result.Error
}
