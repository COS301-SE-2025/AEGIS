package unit_tests

import (
	"aegis-api/handlers"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/evidence/metadata"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ─── Mock Metadata Service ─────────────────────────────
type MockMetadataService struct {
	mock.Mock
}

func (m *MockMetadataService) UploadEvidence(req metadata.UploadEvidenceRequest) error {
	args := m.Called(req)
	return args.Error(0)
}

func (m *MockMetadataService) GetEvidenceByCaseID(caseID uuid.UUID) ([]metadata.Evidence, error) {
	args := m.Called(caseID)
	return args.Get(0).([]metadata.Evidence), args.Error(1)
}

func (m *MockMetadataService) FindEvidenceByID(id uuid.UUID) (*metadata.Evidence, error) {
	args := m.Called(id)
	return args.Get(0).(*metadata.Evidence), args.Error(1)
}

// ─── Fake Mongo Logger for Audit ───────────────────────
type FakeMongoLogger struct{}

func (f *FakeMongoLogger) Log(_ *gin.Context, _ auditlog.AuditLog) error {

	return nil
}

// ─── Setup Test Router ────────────────────────────────
func setupTestRouter(service *MockMetadataService) (*gin.Engine, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	rec := httptest.NewRecorder()

	// Use the real auditlog.NewAuditLogger with a fake Mongo logger
	fakeMongo := &FakeMongoLogger{}
	zapLogger := auditlog.NewZapLogger()
	auditLogger := auditlog.NewAuditLogger(fakeMongo, zapLogger)

	handler := handlers.NewMetadataHandler(service, auditLogger, nil)

	router.GET("/evidence/:id", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("userRole", "Analyst")
		handler.GetEvidenceByID(c)
	})

	router.GET("/evidence/case/:case_id", func(c *gin.Context) {
		c.Set("userID", "test-user-id")
		c.Set("userRole", "Analyst")
		handler.GetEvidenceByCaseID(c)
	})

	return router, rec
}

// ─── Test Get by ID ───────────────────────────────────
func TestGetEvidenceByID_Success(t *testing.T) {
	mockService := new(MockMetadataService)
	testID := uuid.New()
	expected := &metadata.Evidence{
		ID:       testID,
		Filename: "test.pdf",
		FileType: "pdf",
	}

	mockService.On("FindEvidenceByID", testID).Return(expected, nil)
	router, rec := setupTestRouter(mockService)

	req := httptest.NewRequest("GET", "/evidence/"+testID.String(), nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var result metadata.Evidence
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	require.Equal(t, expected.ID, result.ID)
	require.Equal(t, expected.Filename, result.Filename)
}

// ─── Test Get by CaseID ───────────────────────────────
func TestGetEvidenceByCaseID_Success(t *testing.T) {
	mockService := new(MockMetadataService)
	caseID := uuid.New()
	expected := []metadata.Evidence{
		{
			ID:       uuid.New(),
			CaseID:   caseID,
			Filename: "case_file_1.txt",
		},
	}

	mockService.On("GetEvidenceByCaseID", caseID).Return(expected, nil)
	router, rec := setupTestRouter(mockService)

	req := httptest.NewRequest("GET", "/evidence/case/"+caseID.String(), nil)
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var result []metadata.Evidence
	err := json.NewDecoder(rec.Body).Decode(&result)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "case_file_1.txt", result[0].Filename)
}
