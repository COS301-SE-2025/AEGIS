package encryption

import "os"

type Config struct {
    VaultAddress string
    VaultToken   string
    KeyName      string
}

func LoadConfig() *Config {
    return &Config{
        VaultAddress: getEnv("VAULT_ADDR", ""),
        VaultToken:   getEnv("VAULT_TOKEN", ""),
        KeyName:      getEnv("VAULT_KEY_NAME", "app-data-key"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}