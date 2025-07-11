package handlers

import (
	"net/http"

	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *CaseHandler) AssignUserToCase(c *gin.Context) {
	assignerRole, exists := c.Get("userRole")
	userID, _ := c.Get("userID")

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      "",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	if exists {
		actor.Role = assignerRole.(string)
	}

	if !exists {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ASSIGN_USER_TO_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_assignment",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Missing role in token",
		})
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing role in token"})
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id"`
		CaseID     string `json:"case_id"`
		Role       string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ASSIGN_USER_TO_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_assignment",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Invalid JSON payload",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ASSIGN_USER_TO_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_assignment",
				ID:   req.AssigneeID,
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Invalid assignee ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee id"})
		return
	}

	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ASSIGN_USER_TO_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_assignment",
				ID:   req.CaseID,
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Invalid case ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}

	if err := h.CaseService.AssignUserToCase(
		assignerRole.(string), assigneeID, caseID, req.Role,
	); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ASSIGN_USER_TO_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_assignment",
				ID:   caseID.String(),
				AdditionalInfo: map[string]string{
					"assignee_id": assigneeID.String(),
					"role":        req.Role,
				},
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to assign user to case: " + err.Error(),
		})
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Log success
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "ASSIGN_USER_TO_CASE",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case_assignment",
			ID:   caseID.String(),
			AdditionalInfo: map[string]string{
				"assignee_id": assigneeID.String(),
				"role":        req.Role,
			},
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: "User assigned to case successfully",
	})

	c.JSON(http.StatusOK, gin.H{"message": "user assigned to case successfully"})
}
