package encryption

import (
	"context"
	"time"
	"fmt"
)



// Encrypt encrypts a string using the global service (for your existing pattern)
func Encrypt(plaintext string) (string, error) {
	if globalService == nil {
		return "", fmt.Errorf("encryption service not initialized")
	}
	
	if plaintext == "" {
		return "", nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	encrypted, err := globalService.Encrypt(ctx, []byte(plaintext))
	if err != nil {
		return "", err
	}
	
	return encrypted.CipherText, nil
}


func Decrypt(ciphertext string) (string, error) {
	if globalService == nil {
		return "", fmt.Errorf("encryption service not initialized")
	}
	
	if ciphertext == "" {
		return "", nil
	}
	
	// Skip decryption if it's not encrypted (for migration scenarios)
	if !IsEncryptedFormat(ciphertext) {
		return ciphertext, nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	plaintext, err := globalService.Decrypt(ctx, ciphertext)
	if err != nil {
		return "", err
	}
	
	return string(plaintext), nil
}




func IsEncryptedFormat(value string) bool {
	if len(value) < 6 {
		return false
	}
	// Vault format: "vault:v1:..."
	// Mock format: "mock::..."
	return (len(value) > 8 && value[:7] == "vault:v") || 
		   (len(value) > 6 && value[:6] == "mock::")
}
