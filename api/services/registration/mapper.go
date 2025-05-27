package registration

import (
	"time"
)

// This file contains the mapping functions to convert between different layers of the application.

func RegistrationRequestToModel(req RegistrationRequest, hash string) UserModel {
	return UserModel{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
	}
}


func ModelToEntity(model UserModel, id string) User {
	return User{
		ID:           id,
		FullName:     model.FullName,
		Email:        model.Email,
		PasswordHash: model.PasswordHash,
		Role:         model.Role, 
		CreatedAt:    time.Now(),
	}
}

func EntityToResponse(entity User) UserResponse {
	return UserResponse{
		ID:       entity.ID,
		FullName: entity.FullName,
		Email:    entity.Email,
	}
}
