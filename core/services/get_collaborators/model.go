package get_collaborators

import (
	"github.com/google/uuid"
	"time"
)

type Collaborator struct {
	ID         uuid.UUID `json:"id"`
	FullName   string    `json:"full_name"`
	Email      string    `json:"email"`
	Role       string    `json:"role"`
	AssignedAt time.Time `json:"assigned_at"`
}