package metadata

import (
	"io"
	"time"

	"github.com/google/uuid"
)

// Repository interface for evidence operations
type Repository interface {
	SaveEvidence(e *Evidence) error
	FindEvidenceByID(id uuid.UUID) (*Evidence, error)
	FindEvidenceByCaseID(caseID uuid.UUID) ([]Evidence, error)
	AppendEvidenceLog(log *EvidenceLog) error
}

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
	TenantID   uuid.UUID `gorm:"type:uuid;not null" json:"tenant_id"`
	TeamID     uuid.UUID `gorm:"type:uuid;not null" json:"team_id"`
	Filename   string    `gorm:"not null" json:"filename"`
	FileType   string    `gorm:"not null" json:"file_type"`
	IpfsCID    string    `gorm:"column:ipfs_cid;not null" json:"ipfs_cid"`
	FileSize   int64     `json:"file_size"`
	Checksum   string    `gorm:"not null" json:"checksum"`
	Metadata   string    `json:"metadata"`
	UploadedAt time.Time `gorm:"autoCreateTime" json:"uploaded_at"`
}

// EvidenceLog represents an append-only log of evidence actions
type EvidenceLog struct {
	ID           uuid.UUID `gorm:"primaryKey" json:"id"`
	EvidenceID   uuid.UUID `gorm:"type:uuid;not null" json:"evidence_id"`
	Sha256       string    `gorm:"not null" json:"sha256"`
	Sha512       string    `gorm:"not null" json:"sha512"`
	Action       string    `gorm:"not null" json:"action"`
	Result       bool      `json:"result"`
	Timestamp    time.Time `json:"timestamp"`
	Details      string    `json:"details"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	PreviousHash string    `gorm:"column:previous_hash" json:"previous_hash"`
}

func (EvidenceLog) TableName() string {
	return "evidence_log"
}

// GORM table name override
func (Evidence) TableName() string {
	return "evidence"
}
