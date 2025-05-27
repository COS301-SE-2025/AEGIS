package get_collaborators

import (
	"github.com/google/uuid"
)

type Repository interface {
	GetCollaboratorsByCaseID(caseID uuid.UUID) ([]Collaborator, error)
}
