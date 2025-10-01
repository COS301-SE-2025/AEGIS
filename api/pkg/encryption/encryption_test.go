package encryption

import (
    "context"
    "os"
    "testing"
	"strings"
)

func TestMockEncryptionService(t *testing.T) {
    service := NewMockEncryptionService()
    ctx := context.Background()
    
    // Test encryption
    plaintext := "Hello, World!"
    encrypted, err := service.Encrypt(ctx, []byte(plaintext))
    if err != nil {
        t.Fatalf("Encryption failed: %v", err)
    }
    
    if encrypted.CipherText == "" {
        t.Fatal("Ciphertext is empty")
    }
    
    if encrypted.KeyVersion != 1 {
        t.Fatalf("Expected key version 1, got %d", encrypted.KeyVersion)
    }
    
    // Test decryption
    decrypted, err := service.Decrypt(ctx, encrypted.CipherText)
    if err != nil {
        t.Fatalf("Decryption failed: %v", err)
    }
    
    if string(decrypted) != plaintext {
        t.Fatalf("Expected %s, got %s", plaintext, string(decrypted))
    }
}

func TestVaultEncryptionService(t *testing.T) {
    // Skip if Vault is not available
    vaultAddr := os.Getenv("VAULT_ADDR")
    vaultToken := os.Getenv("VAULT_TOKEN")
    
    if vaultAddr == "" || vaultToken == "" {
        t.Skip("Skipping Vault test - VAULT_ADDR or VAULT_TOKEN not set")
    }
    
    service, err := NewVaultEncryptionService()
    if err != nil {
        t.Fatalf("Failed to create Vault service: %v", err)
    }
    
    ctx := context.Background()
    
    // Test encryption
    plaintext := "Hello, Vault!"
    encrypted, err := service.Encrypt(ctx, []byte(plaintext))
    if err != nil {
        t.Fatalf("Encryption failed: %v", err)
    }
    
    if encrypted.CipherText == "" {
        t.Fatal("Ciphertext is empty")
    }
    
    // Test decryption
    decrypted, err := service.Decrypt(ctx, encrypted.CipherText)
    if err != nil {
        t.Fatalf("Decryption failed: %v", err)
    }
    
    if string(decrypted) != plaintext {
        t.Fatalf("Expected %s, got %s", plaintext, string(decrypted))
    }
}

func TestNewService(t *testing.T) {
    // Test with no Vault configured (should return mock)
    originalAddr := os.Getenv("VAULT_ADDR")
    originalToken := os.Getenv("VAULT_TOKEN")
    
    os.Setenv("VAULT_ADDR", "")
    os.Setenv("VAULT_TOKEN", "")
    
    service, err := NewService()
    if err != nil {
        t.Fatalf("NewService failed: %v", err)
    }
    
    if service == nil {
        t.Fatal("Service is nil")
    }
    
    // Test that it's actually the mock service
    ctx := context.Background()
    encrypted, err := service.Encrypt(ctx, []byte("test"))
    if err != nil {
        t.Fatalf("Mock encryption failed: %v", err)
    }
    
    if !strings.HasPrefix(encrypted.CipherText, "mock::") {
        t.Fatal("Expected mock service, but got different service")
    }
    
    // Restore original environment
    if originalAddr != "" {
        os.Setenv("VAULT_ADDR", originalAddr)
    }
    if originalToken != "" {
        os.Setenv("VAULT_TOKEN", originalToken)
    }
}

func TestEncryptDecryptString(t *testing.T) {
    service := NewMockEncryptionService()
    ctx := context.Background()
    
    plaintext := "Test String"
    
    // Test EncryptString
    ciphertext, err := EncryptString(ctx, service, plaintext)
    if err != nil {
        t.Fatalf("EncryptString failed: %v", err)
    }
    
    // Test DecryptString
    decrypted, err := DecryptString(ctx, service, ciphertext)
    if err != nil {
        t.Fatalf("DecryptString failed: %v", err)
    }
    
    if decrypted != plaintext {
        t.Fatalf("Expected %s, got %s", plaintext, decrypted)
    }
}
