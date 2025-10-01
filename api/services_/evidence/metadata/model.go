package metadata

import (
	"io"
	"time"

	"github.com/google/uuid"
	gorm "gorm.io/gorm"
	"aegis-api/pkg/encryption"
	"context"
)

// UploadEvidenceRequest is the input for uploading a file + metadata

type UploadEvidenceRequest struct {
	CaseID     uuid.UUID
	UploadedBy uuid.UUID
	TenantID   uuid.UUID // ✅ Added
	TeamID     uuid.UUID // ✅ Added
	Filename   string
	FileType   string
	FileSize   int64
	FileData   io.Reader
	Metadata   map[string]string
}

// Evidence represents a file uploaded to the system, linked to a case and user.
type Evidence struct {
	ID         uuid.UUID `gorm:"primaryKey" json:"id"`
	CaseID     uuid.UUID `gorm:"type:uuid;not null" json:"case_id"`
	UploadedBy uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	TenantID   uuid.UUID `gorm:"type:uuid;not null" json:"tenant_id"` // ✅ Added
	TeamID     uuid.UUID `gorm:"type:uuid;not null" json:"team_id"`   // ✅ Added
	Filename   string    `gorm:"not null" json:"filename"`
	FileType   string    `gorm:"not null" json:"file_type"`
	IpfsCID    string    `gorm:"column:ipfs_cid;not null" json:"ipfs_cid"`
	FileSize   int64     `json:"file_size"`
	Checksum   string    `gorm:"not null" json:"checksum"`
	Metadata   string    `json:"metadata"`
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploaded_at"`
}

// GORM table name override
func (Evidence) TableName() string {
	return "evidence"
}


func (e *Evidence) BeforeSave(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	// Encrypt IpfsCID
	if e.IpfsCID != "" && !encryption.IsEncryptedFormat(e.IpfsCID) {
		encrypted, err := encryption.EncryptString(ctx, service, e.IpfsCID)
		if err != nil {
			return err
		}
		e.IpfsCID = encrypted
	}

	// Encrypt Checksum
	if e.Checksum != "" && !encryption.IsEncryptedFormat(e.Checksum) {
		encrypted, err := encryption.EncryptString(ctx, service, e.Checksum)
		if err != nil {
			return err
		}
		e.Checksum = encrypted
	}

	// Encrypt Metadata
	if e.Metadata != "" && !encryption.IsEncryptedFormat(e.Metadata) {
		encrypted, err := encryption.EncryptString(ctx, service, e.Metadata)
		if err != nil {
			return err
		}
		e.Metadata = encrypted
	}

	return nil
}

// AfterFind decrypts sensitive fields after SELECT
func (e *Evidence) AfterFind(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	// Decrypt IpfsCID
	if e.IpfsCID != "" {
		decrypted, err := encryption.DecryptString(ctx, service, e.IpfsCID)
		if err != nil {
			return err
		}
		e.IpfsCID = decrypted
	}

	// Decrypt Checksum
	if e.Checksum != "" {
		decrypted, err := encryption.DecryptString(ctx, service, e.Checksum)
		if err != nil {
			return err
		}
		e.Checksum = decrypted
	}

	// Decrypt Metadata
	if e.Metadata != "" {
		decrypted, err := encryption.DecryptString(ctx, service, e.Metadata)
		if err != nil {
			return err
		}
		e.Metadata = decrypted
	}

	return nil
}

// AfterCreate decrypts after INSERT (if you return the created object)
func (e *Evidence) AfterCreate(tx *gorm.DB) error {
	return e.AfterFind(tx)
}