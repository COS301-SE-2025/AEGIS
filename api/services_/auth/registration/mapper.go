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
	user := User{
		ID:                  id,
		FullName:            model.FullName,
		Email:               model.Email,
		PasswordHash:        model.PasswordHash,
		Role:                model.Role,
		CreatedAt:           time.Now(),
		IsVerified:          false,
		TokenVersion:        1, // default for all users
		ExternalTokenStatus: "active",
	}

	if model.Role == "External Collaborator" {
		exp := time.Now().Add(10 * 24 * time.Hour)
		user.ExternalTokenExpiry = &exp
		user.ExternalTokenStatus = "active"
	}

	return user
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
