package evidence_viewer

import (
	"time"
)

type EvidenceDTO struct {
	ID         string    `gorm:"primaryKey;type:uuid;default:gen_random_uuid()" json:"id"`
	CaseID     string    `gorm:"type:uuid;not null" json:"case_id"`
	UploadedBy string    `gorm:"not null" json:"uploaded_by"`
	Filename   string    `gorm:"not null" json:"filename"`
	FileType   string    `gorm:"not null" json:"file_type"`
	IPFSCID    string    `gorm:"not null" json:"ipfs_cid"`
	FileSize   int64     `gorm:"not null" json:"file_size"`
	Checksum   string    `gorm:"not null" json:"checksum"`
	Metadata   string    `gorm:"type:jsonb" json:"metadata"`
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploaded_at"`
}
