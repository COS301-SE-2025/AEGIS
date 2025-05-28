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
	CreateUser(user *User) error
	GetUserByEmail(email string) (*User, error)
	UpdateUser(user *User) error
	GetUserByToken(token string) (*User, error) 
}

//
// ─────────────────────────────────────────────────────────────────────
//   In-Memory Repository (Dev / Testing)
// ─────────────────────────────────────────────────────────────────────
//

// type InMemoryUserRepository struct {
// 	users map[string]*User
// }

// func NewInMemoryUserRepository() *InMemoryUserRepository {
// 	return &InMemoryUserRepository{
// 		users: make(map[string]*User),
// 	}
// }

// func (r *InMemoryUserRepository) CreateUser(user *User) error {
// 	if _, exists := r.users[user.Email]; exists {
// 		return errors.New("user already exists")
// 	}
// 	r.users[user.Email] = user
// 	return nil
// }

// func (r *InMemoryUserRepository) GetUserByEmail(email string) (*User, error) {
// 	user, exists := r.users[email]
// 	if !exists {
// 		return nil, errors.New("user not found")
// 	}
// 	return user, nil
// }

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

// 2. Implement in GormUserRepository
func (r *GormUserRepository) GetUserByToken(token string) (*User, error) {
	var user User
	result := r.db.Where("verification_token = ?", token).First(&user)
	return &user, result.Error
}
func (r *GormUserRepository) UpdateUser(user *User) error {
	result := r.db.Save(user)
	return result.Error
}
