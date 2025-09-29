package encryption

import (
	"context"
	"fmt"
	"time"
)

// MultiDBEncryption handles encryption across different database types
type MultiDBEncryption struct {
	service Service
}

// NewMultiDBEncryption creates a new multi-database encryption helper
func NewMultiDBEncryption(service Service) *MultiDBEncryption {
	return &MultiDBEncryption{service: service}
}

// EncryptForPostgres encrypts data for PostgreSQL storage
func (m *MultiDBEncryption) EncryptForPostgres(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	encrypted, err := m.service.Encrypt(ctx, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("postgres encryption failed: %w", err)
	}
	
	return encrypted.CipherText, nil
}

// DecryptFromPostgres decrypts data from PostgreSQL
func (m *MultiDBEncryption) DecryptFromPostgres(ciphertext string) (string, error) {
	return m.decrypt(ciphertext, "postgres")
}

// EncryptForMongo encrypts data for MongoDB storage  
func (m *MultiDBEncryption) EncryptForMongo(plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	encrypted, err := m.service.Encrypt(ctx, []byte(plaintext))
	if err != nil {
		return "", fmt.Errorf("mongo encryption failed: %w", err)
	}
	
	return encrypted.CipherText, nil
}

// DecryptFromMongo decrypts data from MongoDB
func (m *MultiDBEncryption) DecryptFromMongo(ciphertext string) (string, error) {
	return m.decrypt(ciphertext, "mongo")
}

// EncryptForIPFS encrypts data before storing in IPFS
func (m *MultiDBEncryption) EncryptForIPFS(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second) // Longer timeout for potentially large files
	defer cancel()
	
	encrypted, err := m.service.Encrypt(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("ipfs encryption failed: %w", err)
	}
	
	// For IPFS, we might want to store as JSON with metadata
	return []byte(encrypted.CipherText), nil
}

// DecryptFromIPFS decrypts data retrieved from IPFS
func (m *MultiDBEncryption) DecryptFromIPFS(encryptedData []byte) ([]byte, error) {
	if len(encryptedData) == 0 {
		return nil, nil
	}
	
	ciphertext := string(encryptedData)
	if !IsEncryptedFormat(ciphertext) {
		return encryptedData, nil // Return as-is if not encrypted
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	plaintext, err := m.service.Decrypt(ctx, ciphertext)
	if err != nil {
		return nil, fmt.Errorf("ipfs decryption failed: %w", err)
	}
	
	return plaintext, nil
}

// decrypt is a common decryption helper
func (m *MultiDBEncryption) decrypt(ciphertext, source string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}
	
	if !IsEncryptedFormat(ciphertext) {
		return ciphertext, nil // Return as-is if not encrypted (for migration)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	plaintext, err := m.service.Decrypt(ctx, ciphertext)
	if err != nil {
		return "", fmt.Errorf("%s decryption failed: %w", source, err)
	}
	
	return string(plaintext), nil
}
