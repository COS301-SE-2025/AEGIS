package registration

import (
	"errors"
	"strings" 
	"gorm.io/gorm"
)

// ----------------------
// Interface Definition
// ----------------------

type UserRepository interface {
	CreateUser(user *UserEntity) error
	GetUserByEmail(email string) (*UserEntity, error)
	GetUserByToken(token string) (*UserEntity, error)
	UpdateUser(user *UserEntity) error
}

//
// ─────────────────────────────────────────────────────────────────────
//   In-Memory Repository (Dev / Testing)
// ─────────────────────────────────────────────────────────────────────
//

type InMemoryUserRepository struct {
	users map[string]*UserEntity
}

func NewInMemoryUserRepository() *InMemoryUserRepository {
	return &InMemoryUserRepository{
		users: make(map[string]*UserEntity),
	}
}

func (r *InMemoryUserRepository) CreateUser(user *UserEntity) error {
	if _, exists := r.users[user.Email]; exists {
		return errors.New("user already exists")
	}
	r.users[user.Email] = user
	return nil
}

func (r *InMemoryUserRepository) GetUserByEmail(email string) (*UserEntity, error) {
	user, exists := r.users[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

//
// ─────────────────────────────────────────────────────────────────────
//   GORM Repository (Production)
// ─────────────────────────────────────────────────────────────────────
//

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

func (r *GormUserRepository) CreateUser(user *UserEntity) error {
	err := r.db.Create(user).Error
	if err != nil {
		// Detect uniqueness constraint violation
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "UNIQUE constraint") {
			return errors.New("user already exists")
		}
	}
	return err
}


func (r *GormUserRepository) GetUserByEmail(email string) (*UserEntity, error) {
	var user UserEntity
	result := r.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
