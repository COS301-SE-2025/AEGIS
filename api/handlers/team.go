package handlers

import (
	"fmt"
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

func (h *Handler) GetTeamByID(c *gin.Context) {
	teamIDStr := c.Param("id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}
	fmt.Println("Parsed teamID:", teamID) // 🔍

	team, err := h.TeamRepo.FindByID(teamID)
	if err != nil {
		fmt.Println("Team not found error:", err) // 🔍
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}
	fmt.Println("Fetched team ID from DB:", team.ID) // 🔍

	managerName := "N/A"
	user, err := h.UserRepo.FindByTeamIDAndRole(team.ID, "DFIR Admin")
	if err == nil && user != nil {
		managerName = user.FullName
		fmt.Println("User found:", user.FullName, "with team ID:", user.TeamID) // 🔍
	} else {
		fmt.Println("No DFIR Admin found for team ID:", team.ID, "Error:", err) // 🔍
	}

	result := TeamWithManager{
		ID:       team.ID,
		Name:     team.Name,
		TenantID: team.TenantID,
		Manager:  managerName,
	}

	c.JSON(http.StatusOK, result)
}
