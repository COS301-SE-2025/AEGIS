package handlers

import (
	"fmt"
	"log"
	"net/http"

	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *CaseHandler) AssignUserToCase(c *gin.Context) {
	assignerRole, exists := c.Get("userRole")
	userID, _ := c.Get("userID")
	email, _ := c.Get("email")

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      "",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string),
	}

	if exists {
		actor.Role = assignerRole.(string)
	}

	if !exists {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:  "ASSIGN_USER_TO_CASE",
			Actor:   actor,
			Target:  auditlog.Target{Type: "case_assignment", ID: ""},
			Service: "case", Status: "FAILED", Description: "Missing role in token",
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
			Action:  "ASSIGN_USER_TO_CASE",
			Actor:   actor,
			Target:  auditlog.Target{Type: "case_assignment", ID: ""},
			Service: "case", Status: "FAILED", Description: "Invalid JSON payload",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:  "ASSIGN_USER_TO_CASE",
			Actor:   actor,
			Target:  auditlog.Target{Type: "case_assignment", ID: req.AssigneeID},
			Service: "case", Status: "FAILED", Description: "Invalid assignee ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee id"})
		return
	}

	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:  "ASSIGN_USER_TO_CASE",
			Actor:   actor,
			Target:  auditlog.Target{Type: "case_assignment", ID: req.CaseID},
			Service: "case", Status: "FAILED", Description: "Invalid case ID",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}

	//  Extract assigner's tenant ID
	assignerTenantID, ok := c.Get("tenantID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenant ID in token"})
		return
	}

	//  Check if the assignee belongs to the same tenant
	assigneeUser, err := h.UserRepo.GetUserByID(assigneeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch assignee"})
		return
	}

	if assigneeUser.Role == "Tenant Admin" || assigneeUser.Role == "DFIR Admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot assign users with elevated roles"})
		return
	}
	fmt.Println("Assignee Tenant:", assigneeUser.TenantID)
	fmt.Println("Assigner Tenant:", assignerTenantID)

	assignerTenantIDStr, ok := assignerTenantID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid tenant ID in context"})
		return
	}

	assignerTenantUUID, err := uuid.Parse(assignerTenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tenant ID format"})
		return
	}
	log.Printf("AssignerTenantID: %q (%T)", assignerTenantUUID.String(), assignerTenantUUID)
	log.Printf("AssigneeTenantID: %q (%T)", assigneeUser.TenantID.String(), assigneeUser.TenantID)
	if assigneeUser.TenantID == uuid.Nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "assignee does not have a tenant"})
		return
	}

	// if assigneeUser.TenantID != assignerTenantUUID {
	// 	c.JSON(http.StatusForbidden, gin.H{"error": "cannot assign users from different tenants"})
	// 	return
	// }
	log.Printf("Comparing Tenant IDs: %s == %s ? %v",
		assigneeUser.TenantID.String(),
		assignerTenantUUID.String(),
		assigneeUser.TenantID == assignerTenantUUID)

	userIDRaw, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user ID missing from context"})
		return
	}

	assignerID, err := uuid.Parse(userIDRaw.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	teamIDVal, ok := c.Get("teamID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing team ID in token"})
		return
	}
	teamIDStr, ok := teamIDVal.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid team ID in context"})
		return
	}
	teamUUID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid team ID format"})
		return
	}
	// Proceed with assignment
	if err := h.CaseService.AssignUserToCase(
		assignerRole.(string),
		assigneeID,
		caseID,
		assignerID,
		req.Role,
		assignerTenantUUID,
		teamUUID,
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

func (h *CaseHandler) UnassignUserFromCase(c *gin.Context) {
	// Extract actor metadata for audit log
	assignerIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: missing user ID in token"})
		return
	}

	assignerID, err := uuid.Parse(assignerIDVal.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}
	Email, exists := c.Get("email")

	actor := auditlog.Actor{
		ID:        assignerID.String(),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     Email.(string),
	}

	// Bind request JSON
	var req struct {
		AssigneeID string `json:"assignee_id" binding:"required,uuid"`
		CaseID     string `json:"case_id" binding:"required,uuid"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "UNASSIGN_USER_FROM_CASE",
			Actor:       actor,
			Target:      auditlog.Target{Type: "case_assignment"},
			Service:     "case",
			Status:      "FAILED",
			Description: "Invalid JSON payload",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Parse UUIDs
	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee ID"})
		return
	}

	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	// 🚨 Updated: Pass c to the service method (instead of assignerID)
	if err := h.CaseService.UnassignUserFromCase(c, assigneeID, caseID); err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UNASSIGN_USER_FROM_CASE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_assignment",
				ID:   caseID.String(),
				AdditionalInfo: map[string]string{
					"assignee_id": assigneeID.String(),
				},
			},
			Service:     "case",
			Status:      "FAILED",
			Description: err.Error(),
		})
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// Success log
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UNASSIGN_USER_FROM_CASE",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case_assignment",
			ID:   caseID.String(),
			AdditionalInfo: map[string]string{
				"assignee_id": assigneeID.String(),
			},
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: "User unassigned from case successfully",
	})

	c.JSON(http.StatusOK, gin.H{
		"message":     "user unassigned from case successfully",
		"case_id":     caseID,
		"assignee_id": assigneeID,
	})
}
