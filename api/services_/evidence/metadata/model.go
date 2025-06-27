package metadata

import (
	"time"

	"github.com/google/uuid"
)

// UploadEvidenceRequest is the input for uploading a file + metadata
type UploadEvidenceRequest struct {
	ID         uuid.UUID         `json:"id"`
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
	ID         uuid.UUID `gorm:"primaryKey" json:"id"`
	CaseID     uuid.UUID `gorm:"type:uuid;not null" json:"case_id"`
	UploadedBy uuid.UUID `gorm:"type:uuid;not null" json:"uploaded_by"`
	Filename   string    `gorm:"not null" json:"filename"`
	FileType   string    `gorm:"not null" json:"file_type"`
	IpfsCID    string    `gorm:"column:ipfs_cid;not null" json:"ipfs_cid"` // Fix here
	FileSize   int64     `json:"file_size"`
	Checksum   string    `gorm:"not null" json:"checksum"`
	Metadata   string    `json:"metadata"` // GORM will treat this as TEXT in SQLite// SQLite will treat this as TEXT
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploaded_at"`
}

// GORM table name override
func (Evidence) TableName() string {
	return "evidence"
}
