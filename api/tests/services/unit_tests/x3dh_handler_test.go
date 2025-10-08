package unit_tests

import (
	"aegis-api/internal/x3dh"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockBundleService implements x3dh.BundleServiceInterface
type MockBundleService struct {
	mock.Mock
}

func (m *MockBundleService) GetBundle(ctx context.Context, userID string) (*x3dh.BundleResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*x3dh.BundleResponse), args.Error(1)
}

func (m *MockBundleService) StoreBundle(ctx context.Context, req x3dh.RegisterBundleRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockBundleService) RefillOPKs(ctx context.Context, userID string, opks []x3dh.OneTimePreKeyUpload) error {
	args := m.Called(ctx, userID, opks)
	return args.Error(0)
}

func (m *MockBundleService) RotateSPK(ctx context.Context, userID, newSPK, signature string, expiresAt *time.Time) error {
	args := m.Called(ctx, userID, newSPK, signature, expiresAt)
	return args.Error(0)
}

func (m *MockBundleService) CountAvailableOPKs(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

// ============ GET /bundle/:user_id Tests ============

func TestHandler_GetBundle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockBundleService)

	expectedBundle := &x3dh.BundleResponse{
		IdentityKey:   "ik_pub_key",
		SignedPreKey:  "spk_pub_key",
		SPKSignature:  "spk_sig",
		OneTimePreKey: "opk_pub_key",
		OPKID:         "opk123",
	}

	mockService.On("GetBundle", mock.Anything, "user123").Return(expectedBundle, nil)

	router := gin.New()

	// ✅ THIS IS KEY: Use the ACTUAL RegisterX3DHHandlers from x3dh package
	x3dhGroup := router.Group("/x3dh")
	x3dh.RegisterX3DHHandlers(x3dhGroup, mockService)

	req := httptest.NewRequest("GET", "/x3dh/bundle/user123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_GetBundle_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockBundleService)

	mockService.On("GetBundle", mock.Anything, "nonexistent").
		Return(nil, errors.New("user not found"))

	router := gin.New()
	x3dhGroup := router.Group("/x3dh")
	x3dh.RegisterX3DHHandlers(x3dhGroup, mockService) // Actual handlers

	req := httptest.NewRequest("GET", "/x3dh/bundle/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// ============ POST /register-bundle Tests ============

func TestHandler_RegisterBundle_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockBundleService)

	reqBody := x3dh.RegisterBundleRequest{
		UserID:       "user123",
		IdentityKey:  "ik_pub",
		SignedPreKey: "spk_pub",
		SPKSignature: "spk_sig",
		OneTimePreKeys: []x3dh.OneTimePreKeyUpload{
			{KeyID: "opk1", PublicKey: "key1"},
		},
	}

	mockService.On("StoreBundle", mock.Anything, reqBody).Return(nil)

	router := gin.New()
	x3dhGroup := router.Group("/x3dh")
	x3dh.RegisterX3DHHandlers(x3dhGroup, mockService) // Actual handlers

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/x3dh/register-bundle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestHandler_RegisterBundle_DuplicateKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockBundleService)

	reqBody := x3dh.RegisterBundleRequest{
		UserID:       "user123",
		IdentityKey:  "ik_pub",
		SignedPreKey: "spk_pub",
		SPKSignature: "spk_sig",
	}

	mockService.On("StoreBundle", mock.Anything, reqBody).
		Return(errors.New("duplicate key violation"))

	router := gin.New()
	x3dhGroup := router.Group("/x3dh")
	x3dh.RegisterX3DHHandlers(x3dhGroup, mockService) // Actual handlers

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/x3dh/register-bundle", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}

// ============ GET /opk-count/:user_id Tests ============

func TestHandler_CountOPKs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockBundleService)

	mockService.On("CountAvailableOPKs", mock.Anything, "user123").Return(42, nil)

	router := gin.New()
	x3dhGroup := router.Group("/x3dh")
	x3dh.RegisterX3DHHandlers(x3dhGroup, mockService) // Actual handlers

	req := httptest.NewRequest("GET", "/x3dh/opk-count/user123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ============ POST /refill-opks Tests ============

func TestHandler_RefillOPKs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockBundleService)

	opks := []x3dh.OneTimePreKeyUpload{
		{KeyID: "opk1", PublicKey: "key1"},
		{KeyID: "opk2", PublicKey: "key2"},
	}

	mockService.On("RefillOPKs", mock.Anything, "user123", opks).Return(nil)

	router := gin.New()
	x3dhGroup := router.Group("/x3dh")
	x3dh.RegisterX3DHHandlers(x3dhGroup, mockService) // Actual handlers

	reqBody := x3dh.RefillOPKRequest{
		UserID: "user123",
		OPKs:   opks,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/x3dh/refill-opks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// ============ POST /rotate-spk Tests ============

func TestHandler_RotateSPK_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockBundleService)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)

	mockService.On("RotateSPK", mock.Anything, "user123", "new_spk", "signature",
		mock.MatchedBy(func(t *time.Time) bool {
			if t == nil {
				return false
			}
			return t.Truncate(time.Second).Equal(expiresAt.Truncate(time.Second))
		})).Return(nil)

	router := gin.New()
	x3dhGroup := router.Group("/x3dh")
	x3dh.RegisterX3DHHandlers(x3dhGroup, mockService) // Actual handlers

	expiresAtStr := expiresAt.Format(time.RFC3339)
	reqBody := x3dh.RotateSPKRequest{
		UserID:    "user123",
		NewSPK:    "new_spk",
		Signature: "signature",
		ExpiresAt: &expiresAtStr,
	}

	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/x3dh/rotate-spk", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// Add more test cases for error scenarios, invalid JSON, etc.
