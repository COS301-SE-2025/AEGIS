package handlers

import (
	"aegis-api/services_/case/listArchiveCases"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ListArchivedCasesHandler handles requests to list archived cases
func ListArchivedCasesHandler(service *listArchiveCases.ArchiveCaseService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		tenantID, _ := c.Get("tenantID")
		teamID, _ := c.Get("teamID")
		cases, err := service.ListArchivedCases(userID.(string), tenantID.(string), teamID.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"archived_cases": cases})
	}
}
