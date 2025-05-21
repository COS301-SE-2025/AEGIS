package registration

import (
	
	"github.com/google/uuid"
	"crypto/rand"
	"encoding/hex"
	"errors"
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
func (r *InMemoryUserRepository) GetUserByToken(token string) (*UserEntity, error) {
	for _, user := range r.users {
		if user.VerificationToken == token {
			return user, nil
		}
	}
	return nil, errors.New("token not found")
}
func (r *InMemoryUserRepository) UpdateUser(user *UserEntity) error {
	if _, exists := r.users[user.Email]; !exists {
		return errors.New("user does not exist")
	}
	r.users[user.Email] = user
	return nil
}
