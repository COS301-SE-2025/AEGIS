package get_collaborators

import (
	"time"

	"github.com/google/uuid"
)

type Collaborator struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	AssignedAt time.Time `json:"assigned_at"`
	TeamID     string    `json:"team_id"`
	TenantID   string    `json:"tenant_id"` // Added for multi-tenancy
}
