package encryption

import (
    "context"
)

// Service defines the interface for encryption operations
type Service interface {
    Encrypt(ctx context.Context, plaintext []byte) (*EncryptedData, error)
    Decrypt(ctx context.Context, ciphertext string) ([]byte, error)
}

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
    CipherText string `json:"ciphertext"`
    KeyVersion int    `json:"key_version"`
}

// EncryptString is a convenience method for encrypting strings
func EncryptString(ctx context.Context, service Service, plaintext string) (string, error) {
    encrypted, err := service.Encrypt(ctx, []byte(plaintext))
    if err != nil {
        return "", err
    }
    return encrypted.CipherText, nil
}

// DecryptString is a convenience method for decrypting to strings
func DecryptString(ctx context.Context, service Service, ciphertext string) (string, error) {
    plaintext, err := service.Decrypt(ctx, ciphertext)
    if err != nil {
        return "", err
    }
    return string(plaintext), nil
}

// Global service instance (initialize once in main)
var globalService Service



// InitializeService sets up the global encryption service
func InitializeService() error {
    service, err := NewService()
    if err != nil {
        return err
    }
    globalService = service
    return nil
}

// GetService returns the global service instance
func GetService() Service {
    return globalService
}

// NewService creates a new encryption service (Vault or Mock)
func NewService() (Service, error) {
    cfg := LoadConfig()
    
    // Use mock service for development/testing
    if cfg.VaultAddress == "" || cfg.VaultToken == "" {
        return NewMockEncryptionService(), nil
    }
    
    return NewVaultEncryptionService()
}