package handlers

import (
	"aegis-api/services_/auditlog"
	update_case "aegis-api/services_/case/case_update"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *CaseHandler) UpdateCaseHandler(c *gin.Context) {
	var req update_case.UpdateCaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Extract Case ID from URL
	req.CaseID = c.Param("case_id")

	// Extract multi-tenant info from JWT context
	userIDCtx, userExists := c.Get("userID")
	tenantIDCtx, tenantExists := c.Get("tenantID")
	teamIDCtx, teamExists := c.Get("teamID")
	userRole, _ := c.Get("userRole")

	if !userExists || !tenantExists || !teamExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Convert safely to string
	userID := userIDCtx.(string)
	tenantID := tenantIDCtx.(string)
	teamID := teamIDCtx.(string)

	// Attach Tenant & Team IDs to the request
	req.TenantID = tenantID
	req.TeamID = teamID

	// Call service to update case
	res, err := h.UpdateCaseService.UpdateCaseDetails(c.Request.Context(), &req)
	if err != nil {
		// Log failed attempt
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_CASE",
			Actor: auditlog.Actor{
				ID:        userID,
				Role:      userRole.(string),
				IPAddress: c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
			},
			Target: auditlog.Target{
				Type:           "case",
				ID:             req.CaseID,
				AdditionalInfo: map[string]string{"tenant_id": tenantID, "team_id": teamID},
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Case update failed: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Log success
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPDATE_CASE",
		Actor: auditlog.Actor{
			ID:        userID,
			Role:      userRole.(string),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		},
		Target: auditlog.Target{
			Type:           "case",
			ID:             req.CaseID,
			AdditionalInfo: map[string]string{"tenant_id": tenantID, "team_id": teamID},
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: "Case details updated successfully",
	})

	c.JSON(http.StatusOK, res)
}
