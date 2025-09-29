package sharecase

import (
	"time"

	"github.com/google/uuid"
)

// Token struct (already exists)
type Token struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID    uuid.UUID  `gorm:"type:uuid;not null"`
	CaseID    *uuid.UUID `gorm:"type:uuid"` // 👈 only used for CASE_SHARE
	Token     string     `gorm:"uniqueIndex;not null"`
	Type      string     `gorm:"type:VARCHAR(30);not null"`
	ExpiresAt *time.Time
	Used      bool `gorm:"default:false"`
	//Uses      int  `gorm:"default:0"`
	//MaxUses   *int
	CreatedAt time.Time
}

// CaseCollaborator represents a user linked to a case
type CaseCollaborator struct {
	ID        uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	CaseID    uuid.UUID `gorm:"type:uuid;not null"`
	UserID    uuid.UUID `gorm:"type:uuid;not null"`
	Role      string    `gorm:"type:user_role;not null"` // enum user_role
	InvitedBy uuid.UUID `gorm:"type:uuid;not null"`
	InvitedAt time.Time `gorm:"default:now()"`
	ExpiresAt *time.Time
	Status    string `gorm:"type:VARCHAR(20);default:'active'"` // active, expired, revoked
}
type Case struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title              string    `gorm:"column:title;not null" json:"title"`
	Description        string    `gorm:"column:description" json:"description"`
	Status             string    `gorm:"column:status;type:case_status;default:'open'" json:"status"`
	Priority           string    `gorm:"column:priority;type:case_priority;default:'medium'" json:"priority"`
	InvestigationStage string    `gorm:"column:investigation_stage;type:investigation_stage;default:'analysis'" json:"investigation_stage"`
	CreatedBy          uuid.UUID `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	TeamName           string    `gorm:"column:team_name;type:text;not null" json:"team_name"`
	TenantID           uuid.UUID `gorm:"column:tenant_id;type:uuid;not null" json:"tenant_id"`
	TeamID             uuid.UUID `gorm:"column:team_id;type:uuid" json:"team_id"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

type Permission struct {
	id          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	name        string    `gorm:"unique;not null"`
	description string    `gorm:"not null"`
}
