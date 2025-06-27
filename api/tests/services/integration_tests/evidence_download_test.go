package integration

import (
	"aegis-api/handlers"
	"aegis-api/services_/evidence/evidence_download"
	"aegis-api/services_/evidence/metadata"
	upload "aegis-api/services_/evidence/upload"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDownloadEvidenceIntegration(t *testing.T) {
	// --- Setup ---
	filePath := "tests/services/unit_tests/evidence_upload_file.md"
	content := "This is a sample evidence file for integration testing.\n"
	_ = os.WriteFile(filePath, []byte(content), 0644)

	db, _ := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	_ = db.AutoMigrate(&metadata.Evidence{})
	repo := metadata.NewGormRepository(db)
	ipfsClient := upload.NewIPFSClient("http://localhost:5001")
	metaService := metadata.NewService(repo, ipfsClient)

	// Upload test file first
	//evidenceID := uuid.New()
	req := metadata.UploadEvidenceRequest{
		CaseID:     uuid.MustParse("08bffdb7-a74c-47c8-8bbf-f4df30b6bd54"),
		UploadedBy: uuid.MustParse("27031538-2795-4095-9adf-59bb7bd3fc19"),
		Filename:   "sample.txt",
		FileType:   "text/plain",
		FilePath:   filePath,
		FileSize:   int64(len(content)),
		Metadata: map[string]string{
			"test": "true",
		},
	}
	err := metaService.UploadEvidence(req)
	require.NoError(t, err)

	// Fetch saved evidence to get CID
	var saved metadata.Evidence
	err = db.First(&saved, "filename = ?", "sample.txt").Error
	require.NoError(t, err)

	// --- Setup handler and router ---
	downloadService := evidence_download.NewService(repo, ipfsClient)
	handler := handlers.NewDownloadHandler(downloadService)
	router := gin.Default()
	router.GET("/download/:id", handler.Download)

	// --- Perform GET ---
	w := httptest.NewRecorder()
	reqURL := "/download/" + saved.ID.String()
	reqHTTP, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, reqHTTP)

	// --- Validate Response ---
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	require.True(t, strings.Contains(w.Body.String(), "This is a sample evidence file"))
}
