package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/case/ListCases"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Assuming your ListCases service looks like this:
type ListCasesService interface {
	//ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error)
	GetAllCases(tenantID string) ([]ListCases.Case, error)
	GetCasesByUser(userID string, tenantID string) ([]ListCases.Case, error)
	GetFilteredCases(TenantID, status, priority, createdBy, teamName, titleTerm, sortBy, order string) ([]ListCases.Case, error)
	GetCaseByID(caseID string, tenantID string) (*ListCases.Case, error)
}

type CaseListHandler struct {
	Service ListCasesService
}

func NewCaseListHandler(service ListCasesService) *CaseListHandler {
	return &CaseListHandler{Service: service}
}

func (h *CaseHandler) ListActiveCasesHandler(c *gin.Context) {
	userIDCtx, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	actor := auditlog.Actor{
		ID:        userIDCtx.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	userID := c.Query("user_id")
	if userID == "" {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_ACTIVE_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "active_case_listing",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Missing user_id parameter",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_ACTIVE_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "active_case_listing",
				ID:   userID,
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Invalid user_id format: " + err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	cases, err := h.CaseService.ListActiveCases(userID)
	if err != nil {
		fmt.Printf("Error listing active cases: %v\n", err)
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_ACTIVE_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "active_case_listing",
				ID:   userID,
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to list active cases: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list active cases"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_ACTIVE_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "active_case_listing",
			ID:   userID,
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d active cases for user %s", len(cases), userID),
	})

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}
