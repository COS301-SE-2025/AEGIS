package integration

import (
	"aegis-api/services_/evidence/metadata"
	upload "aegis-api/services_/evidence/upload"
	"testing"

	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRealDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)
	err = db.AutoMigrate(&metadata.Evidence{})
	require.NoError(t, err)
	return db
}
func ensureTestFile(t *testing.T, path, content string) {
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	require.NoError(t, err)

	err = os.WriteFile(path, []byte(content), 0644)
	require.NoError(t, err)
}
func TestUploadEvidence_WithRealIPFS(t *testing.T) {
	db := setupRealDB(t)
	ipfsClient := upload.NewIPFSClient("http://localhost:5001")
	service := metadata.NewService(metadata.NewGormRepository(db), ipfsClient)

	filePath := "tests/services/unit_tests/evidence_upload_file.md"
	ensureTestFile(t, filePath, "This is a sample evidence file for integration testing.")

	// Generate unique values to avoid conflicts
	uniqueID := uuid.New()
	uniqueFilename := "sample_evidence_" + uniqueID.String() + ".txt"

	req := metadata.UploadEvidenceRequest{
		ID:         uniqueID, // Assign a unique ID explicitly
		CaseID:     uuid.MustParse("08bffdb7-a74c-47c8-8bbf-f4df30b6bd54"),
		UploadedBy: uuid.MustParse("27031538-2795-4095-9adf-59bb7bd3fc19"),
		Filename:   uniqueFilename,
		FileType:   "text/plain",
		FilePath:   filePath,
		FileSize:   128,
		Metadata: map[string]string{
			"source": "unit_test",
		},
	}

	err := service.UploadEvidence(req)
	require.NoError(t, err)

	var ev metadata.Evidence
	err = db.First(&ev, "filename = ?", uniqueFilename).Error
	require.NoError(t, err)
	require.Equal(t, "text/plain", ev.FileType)
	require.NotEmpty(t, ev.IpfsCID)
	require.Contains(t, ev.Metadata, "unit_test")
}
