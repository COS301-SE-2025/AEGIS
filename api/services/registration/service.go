package registration

import (
	"golang.org/x/crypto/bcrypt"
	"fmt"
	"log"
	"gorm.io/gorm"

)

// This file contains the service layer of the registration module.
// The service layer handles business logic such as password hashing
// and mapping incoming DTOs to persistence-ready entities.

type RegistrationService struct {
	// repo is an interface to the user repository — used to persist users
	repo UserRepository
}

// NewRegistrationService returns a new instance of the RegistrationService,
// injecting the appropriate UserRepository implementation (e.g. in-memory or GORM).
func NewRegistrationService(repo UserRepository) *RegistrationService {
	return &RegistrationService{repo: repo}
}

// Register takes in a RegistrationRequest DTO, hashes the password,
// builds a domain model, converts it to an entity, and saves it via the repository.
func (s *RegistrationService) Register(req RegistrationRequest) (User, error) {
	//  Check for existing user first
	existingUser, err := s.repo.GetUserByEmail(req.Email)
	if err == nil && existingUser != nil {
		return User{}, fmt.Errorf("user already exists")
	}
	// Only allow continue if error is "not found"
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

	token := generateToken()
	entity.VerificationToken = token
	entity.IsVerified = false

	err = s.repo.CreateUser(&entity)
	if err != nil {
		log.Printf(" Registration failed (duplicate?) for %s: %v", req.Email, err)
		return User{}, err
	}

	log.Printf("✅ Registered new user: %s (%s %s)", entity.Email, entity.FullName, entity.Role)

	sendVerificationEmail(entity.Email, token)

	return entity, nil
}

func sendVerificationEmail(email string, token string) {
	// TODO: Replace with real SMTP integration
	fmt.Printf("📧 Send verification link to %s:\n", email)
	fmt.Printf("👉 http://localhost:8080/verify?token=%s\n", token)
}
