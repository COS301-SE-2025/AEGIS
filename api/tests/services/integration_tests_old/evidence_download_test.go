package integration

import (
	"aegis-api/handlers"
	"aegis-api/services_/evidence/evidence_download"
	"aegis-api/services_/evidence/metadata"
	upload "aegis-api/services_/evidence/upload"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDownloadEvidenceIntegration(t *testing.T) {
	// --- Setup database and repository ---
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&metadata.Evidence{})
	require.NoError(t, err)

	repo := metadata.NewGormRepository(db)

	// --- Setup IPFS client (mock or your actual implementation with local IPFS) ---
	ipfsClient := upload.NewIPFSClient("http://localhost:5001")
	metaService := metadata.NewService(repo, ipfsClient)

	// --- Upload evidence using new API with FileData (io.Reader) ---
	fileContent := "This is a sample evidence file for integration testing.\n"
	fileReader := bytes.NewReader([]byte(fileContent))

	uploadReq := metadata.UploadEvidenceRequest{
		CaseID:     uuid.MustParse("08bffdb7-a74c-47c8-8bbf-f4df30b6bd54"),
		UploadedBy: uuid.MustParse("27031538-2795-4095-9adf-59bb7bd3fc19"),
		Filename:   "sample.txt",
		FileType:   "text/plain",
		FileSize:   int64(len(fileContent)),
		FileData:   fileReader,
		Metadata: map[string]string{
			"test": "true",
		},
	}

	err = metaService.UploadEvidence(uploadReq)
	require.NoError(t, err)

	// --- Fetch saved evidence to get the ID ---
	var saved metadata.Evidence
	err = db.First(&saved, "filename = ?", "sample.txt").Error
	require.NoError(t, err)

	// --- Setup router with download handler ---
	downloadService := evidence_download.NewService(repo, ipfsClient)
	handler := handlers.NewDownloadHandler(downloadService, nil)

	router := gin.Default()
	router.GET("/download/:id", handler.Download)

	// --- Perform HTTP GET request ---
	w := httptest.NewRecorder()
	reqURL := "/download/" + saved.ID.String()
	reqHTTP, _ := http.NewRequest("GET", reqURL, nil)
	router.ServeHTTP(w, reqHTTP)

	// --- Validate response ---
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, "text/plain", w.Header().Get("Content-Type"))
	require.Contains(t, w.Body.String(), "This is a sample evidence file")
}
