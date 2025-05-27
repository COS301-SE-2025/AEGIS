package registration

import (
	
	"github.com/google/uuid"
	"crypto/rand"
	"encoding/hex"
	//"errors"
	"golang.org/x/crypto/bcrypt"
)

// generateID generates a new UUID string
func generateID() string {
	return uuid.New().String()
}
func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// HashPassword hashes a plain-text password using bcrypt.
func HashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashed), err
}
// func (r *InMemoryUserRepository) GetUserByToken(token string) (*User, error) {
// 	for _, user := range r.users {
// 		if user.VerificationToken == token {
// 			return user, nil
// 		}
// 	}
// 	return nil, errors.New("token not found")
// }
// func (r *InMemoryUserRepository) UpdateUser(user *User) error {
// 	if _, exists := r.users[user.Email]; !exists {
// 		return errors.New("user does not exist")
// 	}
// 	r.users[user.Email] = user
// 	return nil
// }
