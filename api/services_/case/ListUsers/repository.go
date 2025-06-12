package ListUsers

import (
	"context"

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
