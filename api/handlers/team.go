package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamWithManager struct {
	ID       uuid.UUID  `json:"id"`
	Name     string     `json:"name"`
	TenantID *uuid.UUID `json:"tenant_id"`
	Manager  string     `json:"manager"` // DFIR Admin full name or "N/A"
}

func (h *Handler) GetTeamsByTenant(c *gin.Context) {
	tenantIDStr := c.Query("tenant_id")
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant_id"})
		return
	}

	teams, err := h.TeamRepo.FindByTenantID(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	var result []TeamWithManager
	for _, team := range teams {
		managerName := "N/A"
		user, err := h.UserRepo.FindByTeamIDAndRole(team.ID, "DFIR Admin")
		if err == nil && user != nil {
			managerName = user.FullName
		}
		result = append(result, TeamWithManager{
			ID:       team.ID,
			Name:     team.Name,
			TenantID: team.TenantID,
			Manager:  managerName,
		})
	}

	c.JSON(http.StatusOK, result)
}
