
package reset_password

import "log"

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
    log.Printf("📧 [MOCK] Sending reset link to %s: token=%s", email, token)
    return nil
}