package registration

import (
	"time"

	"github.com/google/uuid"
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

func ModelToEntity(model UserModel, id uuid.UUID) User {
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
		ID:       entity.ID.String(),
		FullName: entity.FullName,
		Email:    entity.Email,
	}
}

// NewUserModel constructs a new UserModel from a RegistrationRequest DTO and hashed password.
// This allows clean separation between web-layer input and domain-level logic.
func NewUserModel(req RegistrationRequest, hash string) UserModel {
	return UserModel{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
	}
}
