package registration

import (
	"time"
)

// This file contains the mapping functions to convert between different layers of the application.

func RegistrationRequestToModel(req RegistrationRequest, hash string) UserModel {
	return UserModel{
		Name:         req.Name,
		Surname:      req.Surname,
		Email:        req.Email,
		PasswordHash: hash,
	}
}


func ModelToEntity(model UserModel, id string) User {
	return User{
		ID:           id,
		Name:         model.Name,
		Surname:      model.Surname,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		CreatedAt:    time.Now(),
	}
}

func EntityToResponse(entity User) UserResponse {
	return UserResponse{
		ID:       entity.ID,
		FullName: entity.Name + " " + entity.Surname,
		Email:    entity.Email,
	}
}
