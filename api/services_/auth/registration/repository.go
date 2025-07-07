package registration

import (
	"errors"
	"log"
	"strings"

	"gorm.io/gorm"
)

// ----------------------
// Interface Definition
// ----------------------

type GormUserRepository struct {
	db *gorm.DB
}

func NewRegistrationService(repo UserRepository) *RegistrationService {
	return &RegistrationService{repo: repo}
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	if db == nil {
		log.Fatal("DB is nil in NewGormUserRepository")
	}
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) CreateUser(user *User) error {
	err := r.db.Create(user).Error
	if err != nil {
		// Detect uniqueness constraint violation
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return errors.New("user already exists")
		}
	}
	return err
}

func (r *GormUserRepository) GetUserByEmail(email string) (*User, error) {
	var user User
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByFullName fetches a user by full name.
func (r *GormUserRepository) GetUserByFullName(fullName string) (*User, error) {
	var user User
	result := r.db.Where("full_name = ?", fullName).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
func (r *GormUserRepository) UpdateUser(user *User) error {
	return r.db.Save(user).Error
}

func (r *GormUserRepository) UpdateUserTokenInfo(user *User) error {
	return r.db.Model(user).Updates(map[string]interface{}{
		"token_version":         user.TokenVersion,
		"external_token_status": user.ExternalTokenStatus,
		"external_token_expiry": user.ExternalTokenExpiry,
	}).Error
}

func (r *GormUserRepository) GetUserByID(userID string) (*User, error) {
	var u User
	if err := r.db.First(&u, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	return &User{
		ID:                  u.ID,
		Email:               u.Email,
		Role:                u.Role,
		TokenVersion:        u.TokenVersion,
		ExternalTokenStatus: u.ExternalTokenStatus,
		ExternalTokenExpiry: u.ExternalTokenExpiry,
		IsVerified:          u.IsVerified,
	}, nil
}

func (r *GormUserRepository) FindAll() ([]User, error) {
	var users []User
	err := r.db.Find(&users).Error
	return users, err
}
