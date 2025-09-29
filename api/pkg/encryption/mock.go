package encryption

import (
    "context"
    "encoding/base64"
    "fmt"
    "strings"
)

// MockEncryptionService for testing without Vault
type MockEncryptionService struct{}

func NewMockEncryptionService() *MockEncryptionService {
    return &MockEncryptionService{}
}

func (m *MockEncryptionService) Encrypt(ctx context.Context, plaintext []byte) (*EncryptedData, error) {
    // Simple base64 encoding for testing (NOT secure for production)
    ciphertext := base64.StdEncoding.EncodeToString(plaintext)
    return &EncryptedData{
        CipherText: "mock::" + ciphertext,
        KeyVersion: 1,
    }, nil
}

func (m *MockEncryptionService) Decrypt(ctx context.Context, ciphertext string) ([]byte, error) {
    // Remove mock prefix and decode
    if !strings.HasPrefix(ciphertext, "mock::") {
        return nil, fmt.Errorf("invalid mock ciphertext format")
    }
    
    ciphertext = ciphertext[6:] // Remove "mock::" prefix
    return base64.StdEncoding.DecodeString(ciphertext)
}