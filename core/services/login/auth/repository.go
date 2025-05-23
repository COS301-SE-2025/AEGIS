package auth

import (
	"errors"
	"aegis-api/services/registration"
	"aegis-api/db"
)

func GetUserByEmail(email string) (*registration.User, error) {
	var user registration.User

	result := db.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}
