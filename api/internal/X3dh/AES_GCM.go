package x3dh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

type AESGCMCryptoService struct {
	key []byte // must be 16, 24, or 32 bytes for AES-128/192/256
}

func NewAESGCMCryptoService(secretKey []byte) (*AESGCMCryptoService, error) {
	if len(secretKey) != 32 {
		return nil, fmt.Errorf("key must be 32 bytes for AES-256")
	}
	return &AESGCMCryptoService{key: secretKey}, nil
}

func (s *AESGCMCryptoService) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *AESGCMCryptoService) Decrypt(ciphertext string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(raw) < nonceSize {
		return "", fmt.Errorf("invalid ciphertext")
	}

	nonce, ciphertextOnly := raw[:nonceSize], raw[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertextOnly, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
