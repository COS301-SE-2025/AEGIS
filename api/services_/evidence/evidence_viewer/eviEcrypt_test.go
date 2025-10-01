package evidence_viewer

import (
	"testing"

	"aegis-api/pkg/encryption"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	// Create table manually for SQLite compatibility
	err = db.Exec(`
		CREATE TABLE evidence_dtos (
			id TEXT PRIMARY KEY,
			case_id TEXT NOT NULL,
			uploaded_by TEXT NOT NULL,
			filename TEXT NOT NULL,
			file_type TEXT NOT NULL,
			ip_fsc_id TEXT NOT NULL,
			file_size INTEGER NOT NULL,
			checksum TEXT NOT NULL,
			metadata TEXT,
			uploaded_at DATETIME
		)
	`).Error
	assert.NoError(t, err)

	return db
}

func TestEvidenceDTOEncryption(t *testing.T) {
	err := encryption.InitializeService()
	assert.NoError(t, err)

	db := setupTestDB(t)

	t.Run("encrypts on create", func(t *testing.T) {
		evidence := &EvidenceDTO{
			ID:         uuid.New().String(),
			CaseID:     "550e8400-e29b-41d4-a716-446655440000",
			UploadedBy: "user-123",
			Filename:   "evidence.pdf",
			FileType:   "application/pdf",
			IPFSCID:    "QmTest123ABC",
			FileSize:   2048,
			Checksum:   "sha256:abc123def456",
			Metadata:   `{"description":"Important evidence"}`,
		}

		err := db.Create(evidence).Error
		assert.NoError(t, err)

		var retrieved EvidenceDTO
		err = db.First(&retrieved, "id = ?", evidence.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "QmTest123ABC", retrieved.IPFSCID)
		assert.Equal(t, "sha256:abc123def456", retrieved.Checksum)
		assert.Equal(t, `{"description":"Important evidence"}`, retrieved.Metadata)
	})

	t.Run("prevents double encryption on update", func(t *testing.T) {
		evidence := &EvidenceDTO{
			ID:         uuid.New().String(),
			CaseID:     "660e8400-e29b-41d4-a716-446655440000",
			UploadedBy: "user-456",
			Filename:   "document.docx",
			FileType:   "application/docx",
			IPFSCID:    "QmTest789XYZ",
			FileSize:   4096,
			Checksum:   "sha256:xyz789ghi012",
			Metadata:   `{"category":"financial"}`,
		}

		err := db.Create(evidence).Error
		assert.NoError(t, err)

		evidence.Filename = "updated_document.docx"
		err = db.Save(evidence).Error
		assert.NoError(t, err)

		var updated EvidenceDTO
		err = db.First(&updated, "id = ?", evidence.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "QmTest789XYZ", updated.IPFSCID)
		assert.Equal(t, "sha256:xyz789ghi012", updated.Checksum)
		assert.Equal(t, "updated_document.docx", updated.Filename)
	})

	t.Run("handles empty metadata", func(t *testing.T) {
		evidence := &EvidenceDTO{
			ID:         uuid.New().String(),
			CaseID:     "770e8400-e29b-41d4-a716-446655440000",
			UploadedBy: "user-789",
			Filename:   "photo.jpg",
			FileType:   "image/jpeg",
			IPFSCID:    "QmPhotoHash",
			FileSize:   1024,
			Checksum:   "sha256:photo123",
			Metadata:   "",
		}

		err := db.Create(evidence).Error
		assert.NoError(t, err)

		var retrieved EvidenceDTO
		err = db.First(&retrieved, "id = ?", evidence.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "QmPhotoHash", retrieved.IPFSCID)
		assert.Equal(t, "", retrieved.Metadata)
	})

	t.Run("bulk operations decrypt correctly", func(t *testing.T) {
		caseID := "880e8400-e29b-41d4-a716-446655440000"

		evidences := []*EvidenceDTO{
			{
				ID:         uuid.New().String(),
				CaseID:     caseID,
				UploadedBy: "user-1",
				Filename:   "file1.txt",
				FileType:   "text/plain",
				IPFSCID:    "QmFile1",
				FileSize:   512,
				Checksum:   "check1",
				Metadata:   `{"order":1}`,
			},
			{
				ID:         uuid.New().String(),
				CaseID:     caseID,
				UploadedBy: "user-2",
				Filename:   "file2.txt",
				FileType:   "text/plain",
				IPFSCID:    "QmFile2",
				FileSize:   1024,
				Checksum:   "check2",
				Metadata:   `{"order":2}`,
			},
		}

		for _, e := range evidences {
			err := db.Create(e).Error
			assert.NoError(t, err)
		}

		var results []EvidenceDTO
		err := db.Where("case_id = ?", caseID).Find(&results).Error
		assert.NoError(t, err)
		assert.Len(t, results, 2)

		assert.Equal(t, "QmFile1", results[0].IPFSCID)
		assert.Equal(t, "QmFile2", results[1].IPFSCID)
	})
}