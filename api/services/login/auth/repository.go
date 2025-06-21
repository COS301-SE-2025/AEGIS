package auth

import (
	"gorm.io/gorm"
)

type GormUserRepo struct {
	db *gorm.DB
}

// type UserRepository interface {
// 	GetUserByEmail(email string) (*User, error)
// }

// func GetUserByEmail(email string) (*User, error) {
// 	var user User

// 	result := db.DB.Where("email = ?", email).First(&user)
// 	if result.Error != nil {
// 		return nil, errors.New("user not found")
// 	}

// 	return &user, nil
// }
