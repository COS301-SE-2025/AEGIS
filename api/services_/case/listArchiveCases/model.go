package listArchiveCases

// ArchivedCase represents the structure of an archived case
import (
	"time"

	"github.com/google/uuid"
)

// ArchivedCase represents a case that is archived
type ArchivedCase struct {
	ID                 uuid.UUID `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	Title              string    `gorm:"column:title;type:varchar(255);not null" json:"title"`
	Description        string    `gorm:"column:description;type:text" json:"description"`
	Status             string    `gorm:"column:status;type:varchar(50);not null" json:"status"`
	InvestigationStage string    `gorm:"column:investigation_stage;type:varchar(50);not null" json:"investigation_stage"`
	Priority           string    `gorm:"column:priority;type:varchar(50);not null" json:"priority"`
	TeamName           string    `gorm:"column:team_name;type:text;not null" json:"team_name"`
	TeamID             uuid.UUID `gorm:"column:team_id;type:uuid;not null" json:"team_id"`
	TenantID           uuid.UUID `gorm:"column:tenant_id;type:uuid;not null" json:"tenant_id"`
	CreatedBy          uuid.UUID `gorm:"column:created_by;type:uuid;not null" json:"created_by"`
	CreatedAt          time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Progress           int       `gorm:"column:progress;type:int" json:"progress"`
	ArchivedAt         time.Time `gorm:"column:archived_at" json:"archived_at"`
}

// ResponseDTO represents the response for listing archived cases
type ResponseDTO struct {
	Cases []ArchivedCase `json:"cases"`
}
