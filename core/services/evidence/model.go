package evidence

import (
	"time"
	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Evidence struct {
	ID         uuid.UUID         `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CaseID     uuid.UUID         `gorm:"type:uuid;not null"`
	UploadedBy uuid.UUID         `gorm:"type:uuid;not null"`
	Filename   string            `gorm:"not null"`
	FileType   string            `gorm:"not null"`
	IpfsCID    string            `gorm:"column:ipfs_cid;not null"`
	FileSize   int64             `gorm:"check:file_size >= 0"`
	Checksum   string            `gorm:"not null"`
	Metadata   datatypes.JSONMap `gorm:"type:jsonb"` // Flexible JSONB storage
	UploadedAt time.Time         `gorm:"autoCreateTime"`
	Tags       []Tag             `gorm:"many2many:evidence_tags;constraint:OnDelete:CASCADE"` // Association
}

type EvidenceFile struct {
	Filename string
	FileType string
	IpfsCID  string
	Content  []byte
}


func (Evidence) TableName() string {
	return "evidence"
}
