package evidencecount

import (
	"time"

	"github.com/google/uuid"
)

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
