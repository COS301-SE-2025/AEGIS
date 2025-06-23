package metadata

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// UploadEvidenceRequest is the input for uploading a file + metadata
type UploadEvidenceRequest struct {
	CaseID     uuid.UUID         `json:"case_id"`
	UploadedBy uuid.UUID         `json:"uploaded_by"`
	Filename   string            `json:"filename"`
	FileType   string            `json:"file_type"`
	FilePath   string            `json:"file_path"`
	FileSize   int64             `json:"file_size"`
	Metadata   map[string]string `json:"metadata"`
}

// Evidence represents a file uploaded to the system, linked to a case and user.
type Evidence struct {
	ID         uuid.UUID         `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	CaseID     uuid.UUID         `gorm:"type:uuid;not null" json:"case_id"`
	UploadedBy uuid.UUID         `gorm:"type:uuid;not null" json:"uploaded_by"`
	Filename   string            `gorm:"not null" json:"filename"`
	FileType   string            `gorm:"not null" json:"file_type"`
	IpfsCID    string            `gorm:"not null" json:"ipfs_cid"`
	FileSize   int64             `gorm:"check:file_size >= 0" json:"file_size"`
	Checksum   string            `gorm:"not null" json:"checksum"`
	Metadata   datatypes.JSONMap `gorm:"type:jsonb" json:"metadata"` // Use this!
	UploadedAt time.Time         `gorm:"autoCreateTime" json:"uploaded_at"`
}
