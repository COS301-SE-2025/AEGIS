package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/case/ListCases"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Assuming your ListCases service looks like this:
type ListCasesService interface {
	//ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error)
	GetAllCases() ([]ListCases.Case, error)
	GetCasesByUser(userID string) ([]ListCases.Case, error)
	GetFilteredCases(status, priority, createdBy, teamName, titleTerm, sortBy, order string) ([]ListCases.Case, error)
	GetCaseByID(caseID string) (*ListCases.Case, error)
}

type CaseListHandler struct {
	Service ListCasesService
}

func NewCaseListHandler(service ListCasesService) *CaseListHandler {
	return &CaseListHandler{Service: service}
}

func (h *CaseHandler) ListActiveCasesHandler(c *gin.Context) {
	// Extract user info from JWT context
	userIDCtx, exists := c.Get("userID")
	tenantIDCtx, tenantExists := c.Get("tenantID")
	teamIDCtx, teamExists := c.Get("teamID")
	userRole, _ := c.Get("userRole")

	if !exists || !tenantExists || !teamExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDCtx.(string)
	tenantID, tOk := tenantIDCtx.(string)
	teamID, tmOk := teamIDCtx.(string)

	if !ok || !tOk || !tmOk || userID == "" || tenantID == "" || teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user or tenant/team ID in token"})
		return
	}

	actor := auditlog.Actor{
		ID:        userID,
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	// ✅ Fetch active cases with multi-tenancy filtering
	cases, err := h.CaseService.ListActiveCases(userID, tenantID, teamID)
	if err != nil {
		fmt.Printf("Error listing active cases: %v\n", err)
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_ACTIVE_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "active_case_listing",
				ID:   userID,
				AdditionalInfo: map[string]string{
					"tenant_id": tenantID,
					"team_id":   teamID,
				},
			},

			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to list active cases: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list active cases"})
		return
	}

	// ✅ Audit successful request
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_ACTIVE_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "active_case_listing",
			ID:   userID,
			AdditionalInfo: map[string]string{
				"tenant_id": tenantID,
				"team_id":   teamID,
			},
		},

		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d active cases for user %s", len(cases), userID),
	})

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}
