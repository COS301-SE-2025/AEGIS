package registration

import (
	verifyemail "aegis-api/services_/auth/verify_email" // Assuming verifyemail package exists for email verification
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
	entity.IsVerified = false

	err = s.repo.CreateUser(&entity)
	if err != nil {
		log.Printf(" Registration failed (duplicate?) for %s: %v", req.Email, err)
		return User{}, err
	}

	token, err := verifyemail.CreateEmailVerificationToken(s.repo.GetDB(), id)
	if err != nil {
		log.Printf("❌ Failed to create email verification token for %s: %v", req.Email, err)
		return User{}, err
	}

	if err := verifyemail.SendVerificationEmail(entity.Email, token); err != nil {
		log.Printf("❌ Failed to send verification email to %s: %v", entity.Email, err)
		// Consider whether to proceed or rollback
	}

	log.Printf("✅ Registered new user: %s (%s %s)", entity.Email, entity.FullName, entity.Role)
	return entity, nil
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

	token, err := verifyemail.CreateEmailVerificationToken(s.repo.GetDB(), user.ID)
	if err != nil {
		return fmt.Errorf("could not generate verification token: %v", err)
	}

	if err := verifyemail.SendVerificationEmail(user.Email, token); err != nil {
		return fmt.Errorf("could not send verification email: %v", err)
	}

	return nil
}
