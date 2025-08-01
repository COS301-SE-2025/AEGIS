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
