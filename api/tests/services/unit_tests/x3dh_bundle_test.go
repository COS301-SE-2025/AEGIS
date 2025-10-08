package unit_tests

import (
	"aegis-api/internal/x3dh"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockKeyStore2 mocks the KeyStore interface
type MockKeyStore2 struct {
	mock.Mock
}

func (m *MockKeyStore2) GetIdentityKey(ctx context.Context, userID string) (*x3dh.IdentityKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x3dh.IdentityKey), args.Error(1)
}

func (m *MockKeyStore2) GetSignedPreKey(ctx context.Context, userID string) (*x3dh.SignedPreKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x3dh.SignedPreKey), args.Error(1)
}

func (m *MockKeyStore2) ConsumeOneTimePreKey(ctx context.Context, userID string) (*x3dh.OneTimePreKey, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x3dh.OneTimePreKey), args.Error(1)
}

func (m *MockKeyStore2) StoreBundle(ctx context.Context, req x3dh.RegisterBundleRequest, crypto x3dh.CryptoService) error {
	args := m.Called(ctx, req, crypto)
	return args.Error(0)
}

func (m *MockKeyStore2) InsertOPKs(ctx context.Context, userID string, opks []x3dh.OneTimePreKeyUpload) error {
	args := m.Called(ctx, userID, opks)
	return args.Error(0)
}

func (m *MockKeyStore2) RotateSignedPreKey(ctx context.Context, userID, newSPK, signature string, expiresAt *time.Time) error {
	args := m.Called(ctx, userID, newSPK, signature, expiresAt)
	return args.Error(0)
}

func (m *MockKeyStore2) CountAvailableOPKs(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockKeyStore2) CountOPKs(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockKeyStore2) ListUsersWithOPKs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockCryptoService mocks the CryptoService interface
type MockCryptoService struct {
	mock.Mock
}

func (m *MockCryptoService) Encrypt(plaintext string) (string, error) {
	args := m.Called(plaintext)
	return args.String(0), args.Error(1)
}

func (m *MockCryptoService) Decrypt(ciphertext string) (string, error) {
	args := m.Called(ciphertext)
	return args.String(0), args.Error(1)
}

// NoOpAuditLogger implements the audit logger interface but does nothing
type NoOpAuditLogger struct{}

func (n *NoOpAuditLogger) Log(ctx context.Context, log x3dh.AuditLog) error {
	return nil
}

func newNoOpAuditor() *NoOpAuditLogger {
	return &NoOpAuditLogger{}
}

// Helper function to generate valid Ed25519 key pair and signature
func generateValidKeys() (string, string, string) {
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	spkPub, _, _ := ed25519.GenerateKey(rand.Reader)

	sig := ed25519.Sign(priv, spkPub)

	return base64.StdEncoding.EncodeToString(pub),
		base64.StdEncoding.EncodeToString(spkPub),
		base64.StdEncoding.EncodeToString(sig)
}

func TestBundleService_GetBundle_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)
	userID := "user123"
	expectedIK := &x3dh.IdentityKey{
		UserID:    userID,
		PublicKey: "ik_public_key",
	}
	expectedSPK := &x3dh.SignedPreKey{
		UserID:    userID,
		PublicKey: "spk_public_key",
		Signature: "spk_signature",
	}
	expectedOPK := &x3dh.OneTimePreKey{
		ID:        "opk1",
		UserID:    userID,
		PublicKey: "opk_public_key",
	}

	mockStore.On("GetIdentityKey", ctx, userID).Return(expectedIK, nil)
	mockStore.On("GetSignedPreKey", ctx, userID).Return(expectedSPK, nil)
	mockStore.On("ConsumeOneTimePreKey", ctx, userID).Return(expectedOPK, nil)

	bundle, err := service.GetBundle(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, bundle)
	assert.Equal(t, expectedIK.PublicKey, bundle.IdentityKey)
	assert.Equal(t, expectedSPK.PublicKey, bundle.SignedPreKey)
	assert.Equal(t, expectedSPK.Signature, bundle.SPKSignature)
	assert.Equal(t, expectedOPK.PublicKey, bundle.OneTimePreKey)
	assert.Equal(t, expectedOPK.ID, bundle.OPKID)

	mockStore.AssertExpectations(t)
}

func TestBundleService_GetBundle_NoOPKAvailable(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	expectedIK := &x3dh.IdentityKey{PublicKey: "ik_public_key"}
	expectedSPK := &x3dh.SignedPreKey{PublicKey: "spk_public_key", Signature: "sig"}

	mockStore.On("GetIdentityKey", ctx, userID).Return(expectedIK, nil)
	mockStore.On("GetSignedPreKey", ctx, userID).Return(expectedSPK, nil)
	mockStore.On("ConsumeOneTimePreKey", ctx, userID).
		Return(nil, errors.New("sql: no rows in result set"))

	bundle, err := service.GetBundle(ctx, userID)

	assert.NoError(t, err)
	assert.NotNil(t, bundle)
	assert.Equal(t, "", bundle.OneTimePreKey)
	assert.Equal(t, "", bundle.OPKID)

	mockStore.AssertExpectations(t)
}

func TestBundleService_GetBundle_IdentityKeyError(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	expectedErr := errors.New("database error")

	mockStore.On("GetIdentityKey", ctx, userID).Return(nil, expectedErr)

	bundle, err := service.GetBundle(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, bundle)
	assert.Contains(t, err.Error(), "get IK")

	mockStore.AssertExpectations(t)
}

func TestBundleService_GetBundle_SignedPreKeyError(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	expectedIK := &x3dh.IdentityKey{PublicKey: "ik_public_key"}
	expectedErr := errors.New("spk not found")

	mockStore.On("GetIdentityKey", ctx, userID).Return(expectedIK, nil)
	mockStore.On("GetSignedPreKey", ctx, userID).Return(nil, expectedErr)

	bundle, err := service.GetBundle(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, bundle)
	assert.Contains(t, err.Error(), "get SPK")

	mockStore.AssertExpectations(t)
}

func TestBundleService_GetBundle_OPKError(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	expectedIK := &x3dh.IdentityKey{PublicKey: "ik_public_key"}
	expectedSPK := &x3dh.SignedPreKey{PublicKey: "spk_public_key", Signature: "sig"}
	expectedErr := errors.New("database connection error")

	mockStore.On("GetIdentityKey", ctx, userID).Return(expectedIK, nil)
	mockStore.On("GetSignedPreKey", ctx, userID).Return(expectedSPK, nil)
	mockStore.On("ConsumeOneTimePreKey", ctx, userID).Return(nil, expectedErr)

	bundle, err := service.GetBundle(ctx, userID)

	assert.Error(t, err)
	assert.Nil(t, bundle)
	assert.Contains(t, err.Error(), "get OPK")

	mockStore.AssertExpectations(t)
}

func TestBundleService_RefillOPKs_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	opks := []x3dh.OneTimePreKeyUpload{
		{KeyID: "opk1", PublicKey: "key1"},
		{KeyID: "opk2", PublicKey: "key2"},
	}

	mockStore.On("InsertOPKs", ctx, userID, opks).Return(nil)

	err := service.RefillOPKs(ctx, userID, opks)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestBundleService_RefillOPKs_Error(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	opks := []x3dh.OneTimePreKeyUpload{{KeyID: "opk1", PublicKey: "key1"}}
	expectedErr := errors.New("insert failed")

	mockStore.On("InsertOPKs", ctx, userID, opks).Return(expectedErr)

	err := service.RefillOPKs(ctx, userID, opks)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	mockStore.AssertExpectations(t)
}

func TestBundleService_StoreBundle_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	ik, spk, sig := generateValidKeys()

	req := x3dh.RegisterBundleRequest{
		UserID:       "user123",
		IdentityKey:  ik,
		SignedPreKey: spk,
		SPKSignature: sig,
		OneTimePreKeys: []x3dh.OneTimePreKeyUpload{
			{KeyID: "opk1", PublicKey: "key1"},
		},
	}

	mockStore.On("StoreBundle", ctx, req, mockCrypto).Return(nil)

	err := service.StoreBundle(ctx, req)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestBundleService_StoreBundle_InvalidSignature(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	ik, spk, _ := generateValidKeys()
	_, _, wrongSig := generateValidKeys()

	req := x3dh.RegisterBundleRequest{
		UserID:       "user123",
		IdentityKey:  ik,
		SignedPreKey: spk,
		SPKSignature: wrongSig,
	}

	err := service.StoreBundle(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SPK signature")
	mockStore.AssertNotCalled(t, "StoreBundle")
}

func TestBundleService_RotateSPK_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	pub, priv, _ := ed25519.GenerateKey(rand.Reader)
	newSPKPub, _, _ := ed25519.GenerateKey(rand.Reader)
	sig := ed25519.Sign(priv, newSPKPub)

	ikB64 := base64.StdEncoding.EncodeToString(pub)
	newSPK := base64.StdEncoding.EncodeToString(newSPKPub)
	sigB64 := base64.StdEncoding.EncodeToString(sig)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	mockStore.On("GetIdentityKey", ctx, userID).
		Return(&x3dh.IdentityKey{PublicKey: ikB64}, nil)
	mockStore.On("RotateSignedPreKey", ctx, userID, newSPK, sigB64, &expiresAt).
		Return(nil)

	err := service.RotateSPK(ctx, userID, newSPK, sigB64, &expiresAt)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestBundleService_RotateSPK_IdentityKeyError(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	expectedErr := errors.New("user not found")

	mockStore.On("GetIdentityKey", ctx, userID).Return(nil, expectedErr)

	err := service.RotateSPK(ctx, userID, "newSPK", "sig", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fetch IK")
	mockStore.AssertExpectations(t)
}

func TestBundleService_RotateSPK_InvalidSignature(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	ik, spk, _ := generateValidKeys()
	_, _, wrongSig := generateValidKeys()

	mockStore.On("GetIdentityKey", ctx, userID).
		Return(&x3dh.IdentityKey{PublicKey: ik}, nil)

	err := service.RotateSPK(ctx, userID, spk, wrongSig, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid SPK signature")
	mockStore.AssertNotCalled(t, "RotateSignedPreKey")
}

func TestBundleService_CountAvailableOPKs_Success(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	expectedCount := 42

	mockStore.On("CountAvailableOPKs", ctx, userID).Return(expectedCount, nil)

	count, err := service.CountAvailableOPKs(ctx, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockStore.AssertExpectations(t)
}

func TestBundleService_CountAvailableOPKs_Error(t *testing.T) {
	ctx := context.Background()
	mockStore := new(MockKeyStore2)
	mockCrypto := new(MockCryptoService)
	auditor := newNoOpAuditor()

	service := x3dh.NewBundleService(mockStore, mockCrypto, auditor)

	userID := "user123"
	expectedErr := errors.New("database error")

	mockStore.On("CountAvailableOPKs", ctx, userID).Return(0, expectedErr)

	count, err := service.CountAvailableOPKs(ctx, userID)

	assert.Error(t, err)
	assert.Equal(t, 0, count)
	mockStore.AssertExpectations(t)
}
