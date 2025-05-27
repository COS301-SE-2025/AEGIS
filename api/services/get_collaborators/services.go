package get_collaborators

import (
	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCollaborators(caseID uuid.UUID) ([]Collaborator, error) {
	return s.repo.GetCollaboratorsByCaseID(caseID)
}