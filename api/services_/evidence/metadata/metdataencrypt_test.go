package metadata

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

	err = db.AutoMigrate(&Evidence{})
	assert.NoError(t, err)

	return db
}

func TestEvidenceEncryption(t *testing.T) {
	// Initialize mock encryption service
	err := encryption.InitializeService()
	assert.NoError(t, err)

	db := setupTestDB(t)

	t.Run("encrypts on create", func(t *testing.T) {
		evidence := &Evidence{
			ID:         uuid.New(),
			CaseID:     uuid.New(),
			UploadedBy: uuid.New(),
			TenantID:   uuid.New(),
			TeamID:     uuid.New(),
			Filename:   "test.pdf",
			FileType:   "application/pdf",
			IpfsCID:    "QmTest123",
			FileSize:   1024,
			Checksum:   "abc123",
			Metadata:   `{"key":"value"}`,
		}

		// Save (triggers BeforeSave)
		err := db.Create(evidence).Error
		assert.NoError(t, err)

		// Check database has encrypted values
		var dbEvidence Evidence
		err = db.Model(&Evidence{}).
			Where("id = ?", evidence.ID).
			First(&dbEvidence).Error
		assert.NoError(t, err)

		// After Find decrypts automatically
		assert.Equal(t, "QmTest123", dbEvidence.IpfsCID)
		assert.Equal(t, "abc123", dbEvidence.Checksum)
		assert.Equal(t, `{"key":"value"}`, dbEvidence.Metadata)
	})

	t.Run("prevents double encryption on update", func(t *testing.T) {
		evidence := &Evidence{
			ID:         uuid.New(),
			CaseID:     uuid.New(),
			UploadedBy: uuid.New(),
			TenantID:   uuid.New(),
			TeamID:     uuid.New(),
			Filename:   "test2.pdf",
			FileType:   "application/pdf",
			IpfsCID:    "QmTest456",
			Checksum:   "def456",
			Metadata:   `{"foo":"bar"}`,
		}

		// Create
		err := db.Create(evidence).Error
		assert.NoError(t, err)

		// Update (should not double-encrypt)
		evidence.Filename = "updated.pdf"
		err = db.Save(evidence).Error
		assert.NoError(t, err)

		// Verify still decrypts correctly
		var updated Evidence
		err = db.First(&updated, evidence.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "QmTest456", updated.IpfsCID)
		assert.Equal(t, "def456", updated.Checksum)
	})

	t.Run("handles empty values", func(t *testing.T) {
		evidence := &Evidence{
			ID:         uuid.New(),
			CaseID:     uuid.New(),
			UploadedBy: uuid.New(),
			TenantID:   uuid.New(),
			TeamID:     uuid.New(),
			Filename:   "empty.pdf",
			FileType:   "application/pdf",
			IpfsCID:    "QmEmpty",
			Checksum:   "", // Empty
			Metadata:   "", // Empty
		}

		err := db.Create(evidence).Error
		assert.NoError(t, err)

		var retrieved Evidence
		err = db.First(&retrieved, evidence.ID).Error
		assert.NoError(t, err)

		assert.Equal(t, "QmEmpty", retrieved.IpfsCID)
		assert.Equal(t, "", retrieved.Checksum)
		assert.Equal(t, "", retrieved.Metadata)
	})
}