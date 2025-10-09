package handlers

import (
	"aegis-api/services_/auditlog"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ✅ Keep the receiver as CaseHandler and use the correct field name
func (h *CaseHandler) ListActiveCasesHandler(c *gin.Context) {
	userIDv, uok := c.Get("userID")
	tenantIDv, tok := c.Get("tenantID")
	teamIDv, mok := c.Get("teamID")
	rolev, _ := c.Get("userRole")
	if !(uok && tok && mok) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, tenantID, teamID := userIDv.(string), tenantIDv.(string), teamIDv.(string)
	if userID == "" || tenantID == "" || teamID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user or tenant/team ID in token"})
		return
	}

	roleStr, _ := rolev.(string)
	actor := auditlog.Actor{
		ID:        userID,
		Role:      roleStr,
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	// Fetch active cases from service
	cases, err := h.ListActiveCasesServ.ListActiveCases(userID, tenantID, teamID)
	if err != nil {
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

	// Map progress for each active case
	for i := range cases {
		cases[i].Progress = getProgressForStage(cases[i].InvestigationStage)
	}

	// Build response payload
	payload := gin.H{
		"cases": cases,
	}

	// Final response
	c.JSON(http.StatusOK, payload)

	// Audit successful request
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
}
