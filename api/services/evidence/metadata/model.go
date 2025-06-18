package metadata

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Evidence represents the metadata for an uploaded evidence file.
// It includes fields for the case ID, uploader, filename, file type, IPFS CID,
// file size, checksum, metadata, upload timestamp, and associated tags.
// The ID field is a string to accommodate SQLite's primary key requirements.
// The Tags field establishes a many-to-many relationship with the Tag model,
// allowing for flexible tagging of evidence items.

type Evidence struct {
	ID         string `gorm:"primaryKey"` // Use string for SQLite
	CaseID     string
	UploadedBy string
	Filename   string
	FileType   string
	IpfsCID    string `gorm:"column:ipfs_cid"`
	FileSize   int64
	Checksum   string
	Metadata   datatypes.JSON
	UploadedAt time.Time
	Tags       []Tag `gorm:"many2many:evidence_tags;constraint:OnDelete:CASCADE"`
}

// Tag represents a tag that can be associated with evidence.
type Tag struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"uniqueIndex"`
}

// EvidenceMetadataRequest is the request structure for uploading evidence metadata.
type EvidenceMetadataRequest struct {
	CaseID   uuid.UUID              `json:"case_id"`
	Tags     []string               `json:"tags"`
	Checksum string                 `json:"checksum"`
	Metadata map[string]interface{} `json:"metadata"`
}

// EvidenceMetadataResponse is the response structure for evidence metadata upload.
type Service struct {
	repo Repository
}
