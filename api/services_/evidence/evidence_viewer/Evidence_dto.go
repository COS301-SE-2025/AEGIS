package evidence_viewer

import (
	"context"
	"time"

	"aegis-api/pkg/encryption"
	"gorm.io/gorm"
)

type EvidenceDTO struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	CaseID     string    `gorm:"type:uuid;not null" json:"case_id"`
	UploadedBy string    `gorm:"not null" json:"uploaded_by"`
	Filename   string    `gorm:"not null" json:"filename"`
	FileType   string    `gorm:"not null" json:"file_type"`
	IPFSCID    string    `gorm:"not null" json:"ipfs_cid"` // ✅ Encrypted
	FileSize   int64     `gorm:"not null" json:"file_size"`
	Checksum   string    `gorm:"not null" json:"checksum"` // ✅ Encrypted
	Metadata   string    `gorm:"type:jsonb" json:"metadata"` // ✅ Encrypted
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploaded_at"`
}

type EvidenceFile struct {
	ID   string `json:"id"`
	Data []byte `json:"data"`
}


// BeforeSave encrypts sensitive fields before INSERT/UPDATE
func (e *EvidenceDTO) BeforeSave(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	// Encrypt IPFSCID
	if e.IPFSCID != "" && !encryption.IsEncryptedFormat(e.IPFSCID) {
		encrypted, err := encryption.EncryptString(ctx, service, e.IPFSCID)
		if err != nil {
			return err
		}
		e.IPFSCID = encrypted
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
func (e *EvidenceDTO) AfterFind(tx *gorm.DB) error {
	ctx := context.Background()
	service := encryption.GetService()

	// Decrypt IPFSCID
	if e.IPFSCID != "" {
		decrypted, err := encryption.DecryptString(ctx, service, e.IPFSCID)
		if err != nil {
			return err
		}
		e.IPFSCID = decrypted
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

// AfterCreate decrypts after INSERT
func (e *EvidenceDTO) AfterCreate(tx *gorm.DB) error {
	return e.AfterFind(tx)
}