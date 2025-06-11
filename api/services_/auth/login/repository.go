package login

import (
	"aegis-api/db"
	"aegis-api/services_/auth/registration"
	"errors"
)

func GetUserByEmail(email string) (*registration.User, error) {
	var user registration.User

	result := db.DB.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}
