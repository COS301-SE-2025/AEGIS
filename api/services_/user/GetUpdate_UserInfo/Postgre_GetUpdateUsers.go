package GetUpdate_UserInfo

import (
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresUserRepository struct {
	DB *gorm.DB
}

func NewPostgresUserRepository(db *gorm.DB) *PostgresUserRepository {
	return &PostgresUserRepository{DB: db}
}


func (r *PostgresUserRepository) GetUserByID(userID uuid.UUID) (*User, error) {
	var user User
	if err := r.DB.First(&user, "id = ?", userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}



func (r *PostgresUserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := r.DB.First(&user, "email = ?", email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}


func (r *PostgresUserRepository) UpdateUser(userID uuid.UUID, updates map[string]interface{}) error {
	return r.DB.Model(&User{}).Where("id = ?", userID).Updates(updates).Error
}


func (r *PostgresUserRepository) GetUserRoles(userID uuid.UUID) ([]string, error) {
	var roles []string
	if err := r.DB.Model(&UserRole{}).Where("user_id = ?", userID).Pluck("role", &roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}
