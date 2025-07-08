package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GET /cases/all
func (h *CaseHandler) GetAllCasesHandler(c *gin.Context) {
	cases, err := h.ListCasesService.GetAllCases()
	if err != nil {
		fmt.Printf("[GetAllCasesHandler] failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve cases"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GET /cases/user/:user_id
func (h *CaseHandler) GetCasesByUserHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}
	cases, err := h.ListCasesService.GetCasesByUser(userID)
	if err != nil {
		fmt.Printf("[GetCasesByUserHandler] failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve cases for user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GET /cases/filter
func (h *CaseHandler) GetFilteredCasesHandler(c *gin.Context) {
	// Parse query params
	status := c.Query("status")
	priority := c.Query("priority")
	createdBy := c.Query("created_by")
	teamName := c.Query("team_name")
	titleTerm := c.Query("title_term")
	sortBy := c.Query("sort_by")
	order := c.Query("order")

	cases, err := h.ListCasesService.GetFilteredCases(
		status, priority, createdBy, teamName, titleTerm, sortBy, order,
	)
	if err != nil {
		fmt.Printf("[GetFilteredCasesHandler] failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not apply filters"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cases": cases})
}
