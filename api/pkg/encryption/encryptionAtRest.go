package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
)

var masterKey []byte


func Init() error {
    key := os.Getenv("ENCRYP_REST_MASTER_KEY") 
    if len(key) != 32 {
        return errors.New("ENCRYP_REST_MASTER_KEY must be 32 characters long")
    }
    masterKey = []byte(key)
    return nil
}


func Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}


func Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ct := data[:nonceSize], data[nonceSize:]

	// Decrypt → []byte
	plaintext, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}