package encryption

import (
    "context"
    "encoding/base64"
    "fmt"
    "encoding/json"
    "github.com/hashicorp/vault/api"
)

type VaultEncryptionService struct {
    client  *api.Client
    keyName string
}

func NewVaultEncryptionService() (*VaultEncryptionService, error) {
    cfg := LoadConfig()
    
    clientConfig := api.DefaultConfig()
    clientConfig.Address = cfg.VaultAddress
    
    client, err := api.NewClient(clientConfig)
    if err != nil {
        return nil, fmt.Errorf("failed to create Vault client: %w", err)
    }
    
    client.SetToken(cfg.VaultToken)
    
    // Test connection
    _, err = client.Sys().Health()
    if err != nil {
        return nil, fmt.Errorf("vault health check failed: %w", err)
    }
    
    return &VaultEncryptionService{
        client:  client,
        keyName: cfg.KeyName,
    }, nil
}

func (v *VaultEncryptionService) Encrypt(ctx context.Context, plaintext []byte) (*EncryptedData, error) {
    plaintextB64 := base64.StdEncoding.EncodeToString(plaintext)
    
    data := map[string]interface{}{
        "plaintext": plaintextB64,
    }
    
    secret, err := v.client.Logical().WriteWithContext(ctx,
        fmt.Sprintf("transit/encrypt/%s", v.keyName),
        data,
    )
    if err != nil {
        return nil, fmt.Errorf("encryption failed: %w", err)
    }
    
    if secret == nil || secret.Data == nil {
        return nil, fmt.Errorf("empty response from Vault")
    }
    
    ciphertext, ok := secret.Data["ciphertext"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid ciphertext in Vault response")
    }
    
    // Extract key version if available
    keyVersion := 1
    if kv, ok := secret.Data["key_version"]; ok {
        if kvInt, ok := kv.(json.Number); ok {
            if v, err := kvInt.Int64(); err == nil {
                keyVersion = int(v)
            }
        }
    }
    
    return &EncryptedData{
        CipherText: ciphertext,
        KeyVersion: keyVersion,
    }, nil
}

func (v *VaultEncryptionService) Decrypt(ctx context.Context, ciphertext string) ([]byte, error) {
    data := map[string]interface{}{
        "ciphertext": ciphertext,
    }
    
    secret, err := v.client.Logical().WriteWithContext(ctx,
        fmt.Sprintf("transit/decrypt/%s", v.keyName),
        data,
    )
    if err != nil {
        return nil, fmt.Errorf("decryption failed: %w", err)
    }
    
    if secret == nil || secret.Data == nil {
        return nil, fmt.Errorf("empty response from Vault")
    }
    
    plaintextB64, ok := secret.Data["plaintext"].(string)
    if !ok {
        return nil, fmt.Errorf("invalid plaintext in Vault response")
    }
    
    plaintext, err := base64.StdEncoding.DecodeString(plaintextB64)
    if err != nil {
        return nil, fmt.Errorf("base64 decode failed: %w", err)
    }
    
    return plaintext, nil
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
