package registration
//Web Layer
// This layer is responsible for handling HTTP requests and responses.
// It decodes incoming requests, calls the service layer, and encodes the responses.
// It should not contain any business logic or data access code.
// It should only handle HTTP-specific concerns like request/response encoding/decoding.

import (
	"errors"
	"regexp"
	"strings"
)

type RegistrationRequest struct {
	FullName string `json:"full_name"`
	Email   string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
	/*
Password is required to be hashed.
From client side, password is sent in plain text.
On the server side, it is hashed using bcrypt before storage.
*/

}

type UserResponse struct {
	ID      string `json:"id"`
	FullName string `json:"full_name"`
	Email   string `json:"email"`
}



func (r RegistrationRequest) Validate() error {
	if strings.TrimSpace(r.FullName) == "" {
		return errors.New("Full name is required")
	}
	
	
	matched, _ := regexp.MatchString(`^[\w\.-]+@[\w\.-]+\.\w+$`, r.Email)
	if !matched {
		return errors.New("invalid email address format")
	}

	if len(r.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if !isStrongPassword(r.Password) {
		return errors.New("password must contain uppercase, lowercase, and a digit")
	}
		if matched, _ := regexp.MatchString(`^[\w\.-]+@[\w\.-]+\.\w+$`, r.Email); !matched {
		return errors.New("invalid email address format")
	}
		// Validate ENUM role
	validRoles := map[string]bool{
		"Incident Responder": true, "Forensic Analyst": true, "Malware Analyst": true,
		"Threat Intelligent Analyst": true, "DFIR Manager": true, "Legal/Compliance Liaison": true,
		"Detection Engineer": true, "Generic user": true,
	}
	if _, ok := validRoles[r.Role]; !ok {
		return errors.New("invalid user role")
	}
	return nil
}

func isStrongPassword(password string) bool {
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasDigit := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasUpper && hasLower && hasDigit
}