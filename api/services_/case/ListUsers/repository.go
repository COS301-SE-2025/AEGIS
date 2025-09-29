package ListUsers

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
	//"aegis-api/db"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) GetAllUsers(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.Table("users").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *UserRepository) GetUsersByTenant(ctx context.Context, tenantID uuid.UUID, page int, pageSize int) ([]User, int64, error) {
	var users []User
	var total int64
	db := r.db.WithContext(ctx).Table("users").Where("tenant_id = ?", tenantID)
	db.Count(&total)
	err := db.Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}
	return users, total, nil
}
func (r *UserRepository) GetUserByID(ctx context.Context, userID uuid.UUID) (*User, error) {
	var user User
	err := r.db.WithContext(ctx).Table("users").Where("id = ?", userID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
