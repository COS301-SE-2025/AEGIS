package handlers

import (
	"context"
	"net/http"

	"aegis-api/services_/case/case_deletion"

	"github.com/gin-gonic/gin"
)

type CaseDeletionHandler struct {
	Service *case_deletion.Service
}

func NewCaseDeletionHandler(service *case_deletion.Service) *CaseDeletionHandler {
	return &CaseDeletionHandler{Service: service}
}

// ArchiveCaseHandler sets status to archived for a case
func (h *CaseDeletionHandler) ArchiveCaseHandler(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing case_id"})
		return
	}
	ctx := context.Background()
	if err := h.Service.ArchiveCase(ctx, caseID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Case archived successfully"})
}
