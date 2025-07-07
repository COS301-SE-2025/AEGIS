package handlers

import (
	"aegis-api/services_/case/ListActiveCases"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Assuming your ListCases service looks like this:
type ListCasesService interface {
	ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error)
}

func (s *CaseServices) ListActiveCases(userID string) ([]ListActiveCases.ActiveCase, error) {
	return s.listCase.ListActiveCases(userID)
}

func (h *CaseHandler) ListActiveCasesHandler(c *gin.Context) {
	// Extract userID from query or path
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Validate UUID format
	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
		return
	}

	// Call service
	cases, err := h.CaseService.ListActiveCases(userID)
	if err != nil {
		fmt.Printf("Error listing active cases: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not list active cases"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}
