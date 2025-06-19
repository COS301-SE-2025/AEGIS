package registration

import (
	"fmt"
	"log"

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

	entity.IsVerified = false

	err = s.repo.CreateUser(&entity)
	if err != nil {
		log.Printf(" Registration failed (duplicate?) for %s: %v", req.Email, err)
		return User{}, err
	}

	log.Printf("✅ Registered new user: %s (%s %s)", entity.Email, entity.FullName, entity.Role)
	// JWT-based verification link
	token, err := GenerateJWT(entity.ID, entity.Email, "verify")
	if err != nil {
		log.Printf("Failed to create verification token: %v", err)
		return User{}, err
	}
	sendVerificationEmail(entity.Email, token)

	return entity, nil
}
func sendVerificationEmail(email, token string) {
	verificationURL := fmt.Sprintf("http://localhost:8080/api/v1/verify?token=%s", token)

	log.Printf("📧 [DEV] Simulated verification email to %s", email)
	log.Printf("🔗 [DEV] Verification URL: %s", verificationURL)
}

// func sendVerificationEmail(email string, token string) {
// 	m := gomail.NewMessage()
// 	m.SetHeader("From", "no-replycapstone.incidentintel@gmail.com")
// 	m.SetHeader("To", email)
// 	m.SetHeader("Subject", "Confirm your email address")
// 	verificationURL := fmt.Sprintf("http://localhost:8080/api/v1/verify?token=%s", token)
// 	body := fmt.Sprintf(
// 		`<p>Hi there,</p>
//          <p>Thank you for registering. Click <a href="%s">here</a> to verify your email.</p>
//          <p>This link will expire in 24 hours.</p>`,
// 		verificationURL,
// 	)
// 	m.SetBody("text/html", body)

// 	d := gomail.NewDialer("smtp.mailgun.org", 587, "postmaster@yourdomain.com", "ArDTUKCwsfqA6r.")
// 	if err := d.DialAndSend(m); err != nil {
// 		log.Printf("ERROR sending verification email to %s: %v\n", email, err)
// 	}
// }

func (r *GormUserRepository) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	return r.db.
		Model(&User{}).
		Where("id = ?", userID).
		Update("password_hash", hashedPassword).
		Error
}

// VerifyUser looks up a User by the given token, marks them as verified, and clears the token.
func (s *RegistrationService) VerifyUser(token string) error {
	claims, err := VerifyJWT(token)
	if err != nil {
		return fmt.Errorf("invalid or expired verification token")
	}
	if claims.Role != "verify" {
		return fmt.Errorf("token is not a verification token")
	}

	user, err := s.repo.GetUserByEmail(claims.Email)
	if err != nil {
		return fmt.Errorf("user not found")
	}
	if user.IsVerified {
		return nil
	}
	user.IsVerified = true
	if err := s.repo.UpdateUser(user); err != nil {
		return fmt.Errorf("could not update user verification status: %v", err)
	}
	return nil
}

func (s *RegistrationService) ResendVerificationEmail(req ResendVerificationRequest) error {
	user, err := s.repo.GetUserByEmail(req.Email)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsVerified {
		return fmt.Errorf("user already verified")
	}

	token, err := GenerateJWT(user.ID, user.Email, "verify")
	if err != nil {
		return fmt.Errorf("could not generate verification token: %v", err)
	}

	sendVerificationEmail(user.Email, token)
	return nil
}
