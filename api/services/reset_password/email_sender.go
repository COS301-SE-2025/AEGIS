package reset_password

import (
	"fmt"
	"log"
)

type MockEmailSender struct{}

func NewMockEmailSender() *MockEmailSender {
	return &MockEmailSender{}
}

// func (m *MockEmailSender) SendPasswordResetEmail(email string, token string) error {
// 	log.Printf("Sending password reset email to %s with token %s", email, token)
// 	// Replace with real email sending logic here
// 	return nil
// }

func (m *MockEmailSender) SendPasswordResetEmail(email string, token string) error {
	resetURL := fmt.Sprintf("http://localhost:8080/reset-password?token=%s", token)

	log.Printf("📧 [DEV] Simulated reset email to %s", email)
	log.Printf("🔗 [DEV] Reset Password Link: %s", resetURL)

	return nil
}
