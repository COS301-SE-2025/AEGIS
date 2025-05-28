package case_status_update

import (
	"fmt"

	"github.com/google/uuid"
)

type CaseStatusService struct {
	repo CaseStatusRepository
}

func NewCaseStatusService(repo CaseStatusRepository) *CaseStatusService {
	return &CaseStatusService{repo: repo}
}

// UpdateCaseStatus allows only Admins to change a case's status.
func (s *CaseStatusService) UpdateCaseStatus(req UpdateCaseStatusRequest, requesterRole string) error {
	if requesterRole != "Admin" {
		return fmt.Errorf("unauthorized: only Admins can update case status")
	}

	if _, err := uuid.Parse(req.CaseID); err != nil {
		return fmt.Errorf("invalid case UUID: %w", err)
	}

	return s.repo.UpdateStatus(req.CaseID, req.Status)
}
