package case_assign

import "github.com/google/uuid"

type AdminChecker interface {
	IsAdmin(userID uuid.UUID) (bool, error)
}

type CaseAssignmentRepoInterface interface {
	AssignRole(userID, caseID uuid.UUID, role string) error
	UnassignRole(userID, caseID uuid.UUID) error
	IsAdmin(userID uuid.UUID) (bool, error)
}
