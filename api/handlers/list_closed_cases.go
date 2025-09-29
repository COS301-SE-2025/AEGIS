package handlers

import (
	"aegis-api/services_/auditlog"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *CaseHandler) ListClosedCasesHandler(c *gin.Context) {
	// Extract IDs from context
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token data"})
		return
	}

	actor := auditlog.Actor{
		ID:        userID,
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	// Fetch closed cases with multi-tenancy filtering
	cases, err := h.CaseService.ListClosedCases(userID, tenantID, teamID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_CLOSED_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type:           "closed_case_listing",
				ID:             userID,
				AdditionalInfo: map[string]string{"tenant_id": tenantID, "team_id": teamID},
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to list closed cases: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list closed cases"})
		return
	}

	// Map progress for each closed case
	for i := range cases {
		cases[i].Progress = getProgressForStage(cases[i].InvestigationStage)
	}

	// Audit successful retrieval
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_CLOSED_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type:           "closed_case_listing",
			ID:             userID,
			AdditionalInfo: map[string]string{"tenant_id": tenantID, "team_id": teamID},
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d closed cases for user %s", len(cases), userID),
	})

	c.JSON(http.StatusOK, gin.H{"closed_cases": cases})

}

// getProgressForStage returns a progress value (0-100) based on investigation stage
func getProgressForStage(stage string) int {
	switch stage {
	case "Triage":
		return 10
	case "Evidence Collection":
		return 25
	case "Analysis":
		return 40
	case "Correlation & Threat Intelligence":
		return 55
	case "Containment & Eradication":
		return 70
	case "Recovery":
		return 85
	case "Reporting & Documentation":
		return 95
	case "Case Closure & Review":
		return 100
	default:
		return 0
	}
}
