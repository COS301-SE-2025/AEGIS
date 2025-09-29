package unit_tests

import (
	"encoding/base64"
	"testing"

	x3 "aegis-api/internal/x3dh" // Update with your actual module path

	"github.com/stretchr/testify/assert"
)

func TestAESGCM_EncryptReturnsBase64(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("thisis32byteslongpassphrase123456")) // dummy 32-byte key

	crypto, err := x3.NewAESGCMCryptoService(key)
	assert.NoError(t, err)

	encrypted, err := crypto.Encrypt("hello world")
	assert.NoError(t, err)
	assert.NotEmpty(t, encrypted)

	_, err = base64.StdEncoding.DecodeString(encrypted)
	assert.NoError(t, err)
}

func TestAESGCM_EncryptProducesDifferentOutput(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("thisis32byteslongpassphrase123456"))

	crypto, _ := x3.NewAESGCMCryptoService(key)

	enc1, _ := crypto.Encrypt("hello world")
	enc2, _ := crypto.Encrypt("hello world")

	assert.NotEqual(t, enc1, enc2, "Encrypt should be non-deterministic due to nonce")
}

func TestAESGCM_DecryptReversesEncryption(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("thisis32byteslongpassphrase123456"))

	crypto, _ := x3.NewAESGCMCryptoService(key)

	plaintext := "secure message"
	encrypted, err := crypto.Encrypt(plaintext)
	assert.NoError(t, err)

	decrypted, err := crypto.Decrypt(encrypted)
	assert.NoError(t, err)
	assert.Equal(t, plaintext, decrypted)
}

func TestAESGCM_DecryptFailsOnInvalidBase64(t *testing.T) {
	key := make([]byte, 32)
	copy(key, []byte("thisis32byteslongpassphrase123456"))

	crypto, _ := x3.NewAESGCMCryptoService(key)

	_, err := crypto.Decrypt("!!!not_base64!!!")
	assert.Error(t, err)
}

func TestAESGCM_DecryptFailsWithWrongKey(t *testing.T) {
	key1 := make([]byte, 32)
	copy(key1, []byte("thisis32byteslongpassphrase123456"))
	key2 := make([]byte, 32)
	copy(key2, []byte("another32bytepassphraseisdifferent"))

	crypto1, _ := x3.NewAESGCMCryptoService(key1)
	crypto2, _ := x3.NewAESGCMCryptoService(key2)

	encrypted, _ := crypto1.Encrypt("secret")
	_, err := crypto2.Decrypt(encrypted)

	assert.Error(t, err)
}
