package case_assign

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminChecker interface {
	IsAdminFromContext(ctx *gin.Context) (bool, error)
}

type CaseAssignmentRepoInterface interface {
	AssignRole(userID, caseID uuid.UUID, role string, tenantID, teamID uuid.UUID) error
	UnassignRole(userID, caseID uuid.UUID) error
	GetCaseByID(caseID uuid.UUID, caseDetails *Case) error
}

type UserRepo interface {
	GetUserByID(userID uuid.UUID) (*User, error)
}
