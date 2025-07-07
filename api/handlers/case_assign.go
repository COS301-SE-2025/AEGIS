package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (h *CaseHandler) AssignUserToCase(c *gin.Context) {
	assignerRole, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing role in token"})
		return
	}

	var req struct {
		AssigneeID string `json:"assignee_id"`
		CaseID     string `json:"case_id"`
		Role       string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	assigneeID, err := uuid.Parse(req.AssigneeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid assignee id"})
		return
	}
	caseID, err := uuid.Parse(req.CaseID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case id"})
		return
	}

	if err := h.CaseService.AssignUserToCase(
		assignerRole.(string), assigneeID, caseID, req.Role,
	); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user assigned to case successfully"})
}
