package registration

import (
	// Assuming verifyemail package exists for email verification
	"aegis-api/services_/auditlog"
	"context"
	"fmt"
	"log"

	verifyemail "aegis-api/services_/auth/verify_email"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// This file contains the service layer of the registration module.
// The service layer handles business logic such as password hashing
// and mapping incoming DTOs to persistence-ready entities.

type RegistrationService struct {
	// repo is an interface to the user repository — used to persist users
	repo UserRepository
}

func (s *RegistrationService) VerifyUser(token string) any {
	panic("unimplemented")
}

// NewRegistrationService returns a new instance of the RegistrationService,
// injecting the appropriate UserRepository implementation (e.g. in-memory or GORM).
// func NewRegistrationService(repo UserRepository) *RegistrationService {
// 	return &RegistrationService{repo: repo}
// }

// // Register takes in a RegistrationRequest DTO, hashes the password,
// // builds a domain model, converts it to an entity, and saves it via the repository.
// func (s *RegistrationService) Register(req RegistrationRequest) (User, error) {
// 	//  Check for existing user first
// 	existingUser, err := s.repo.GetUserByEmail(req.Email)
// 	if err == nil && existingUser != nil {
// 		return User{}, fmt.Errorf("user already exists")
// 	}
// 	// Only allow continue if error is "not found"
// 	if err != nil && err != gorm.ErrRecordNotFound {
// 		return User{}, err
// 	}

// 	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
// 	if err != nil {
// 		log.Printf(" Registration failed (hash error) for %s: %v", req.Email, err)
// 		return User{}, err
// 	}

// 	model := NewUserModel(req, string(hash))
// 	id := generateID()
// 	entity := ModelToEntity(model, id)

// 	token := generateToken()
// 	entity.VerificationToken = token
// 	entity.IsVerified = false

// 	err = s.repo.CreateUser(&entity)
// 	if err != nil {
// 		log.Printf(" Registration failed (duplicate?) for %s: %v", req.Email, err)
// 		return User{}, err
// 	}

// 	log.Printf("✅ Registered new user: %s (%s %s)", entity.Email, entity.FullName, entity.Role)

// 	sendVerificationEmail(entity.Email, token)

// 	return entity, nil
// }

// func sendVerificationEmail(email string, token string) {
// 	// TODO: Replace with real SMTP integration

// }

func (s *RegistrationService) Register(req RegistrationRequest) (User, error) {
	// Check for existing user
	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return User{}, fmt.Errorf("user already exists")
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return User{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf(" Registration failed (hash error) for %s: %v", req.Email, err)
		return User{}, err
	}

	model := NewUserModel(req, string(hash))
	id := generateID()
	entity := ModelToEntity(model, id)
	entity.IsVerified = true

	err = s.repo.CreateUser(&entity)
	if err != nil {
		log.Printf(" Registration failed (duplicate?) for %s: %v", req.Email, err)
		return User{}, err
	}

	// 🔕 Email verification disabled
	token, err := verifyemail.CreateEmailVerificationToken(s.repo.GetDB(), id)
	if err != nil {
		log.Printf("❌ Failed to create email verification token for %s: %v", req.Email, err)
		return User{}, err
	}

	if err := verifyemail.SendVerificationEmail(entity.Email, token); err != nil {
		log.Printf("❌ Failed to send verification email to %s: %v", entity.Email, err)
	}

	log.Printf("✅ Registered new user: %s (%s %s)", entity.Email, entity.FullName, entity.Role)
	return entity, nil
}

func (r *GormUserRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	return r.db.
		Model(&User{}).
		Where("id = ?", userID).
		Update("password_hash", hashedPassword).
		Error
}

func (s *RegistrationService) GetAllUsers() ([]User, error) {
	return s.repo.FindAll()
}

func (r *GormUserRepository) GetByID(ctx context.Context, userID string) (*auditlog.User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}

	return &auditlog.User{
		ID:    user.ID.String(),
		Email: user.Email,
		//Role:  user.Role, // Only include this if auditlog.User struct has a Role field
	}, nil
}
