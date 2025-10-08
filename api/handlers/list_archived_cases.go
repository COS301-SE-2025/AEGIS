package handlers

import (
	"aegis-api/services_/case/listArchiveCases"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ArchiveCaseServiceInterface defines the interface for archive case operations
type ArchiveCaseServiceInterface interface {
	ListArchivedCases(userID, tenantID, teamID string) ([]listArchiveCases.ArchivedCase, error)
}

// ListArchivedCasesHandler handles requests to list archived cases
func ListArchivedCasesHandler(service ArchiveCaseServiceInterface) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate that the required fields exist and are strings
		userIDVal, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "userID is required"})
			return
		}
		userID, ok := userIDVal.(string)
		if !ok || userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "userID must be a valid string"})
			return
		}

		tenantIDVal, exists := c.Get("tenantID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "tenantID is required"})
			return
		}
		tenantID, ok := tenantIDVal.(string)
		if !ok || tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "tenantID must be a valid string"})
			return
		}

		teamIDVal, exists := c.Get("teamID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "teamID is required"})
			return
		}
		teamID, ok := teamIDVal.(string)
		if !ok || teamID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "teamID must be a valid string"})
			return
		}

		cases, err := service.ListArchivedCases(userID, tenantID, teamID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"archived_cases": cases})
	}
}
