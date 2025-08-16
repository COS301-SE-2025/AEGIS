package handlers

import (
	"fmt"
	"net/http"

	evidencecount "aegis-api/services_/evidence/evidence_count"

	"github.com/gin-gonic/gin"
)

type EvidenceHandler struct {
	service evidencecount.EvidenceService
}

func NewEvidenceHandler(service evidencecount.EvidenceService) *EvidenceHandler {
	return &EvidenceHandler{service: service}
}

func (h *EvidenceHandler) GetEvidenceCount(c *gin.Context) {
	tenantIDFromToken, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenantID in token context"})
		return
	}
	tenantIDStr := tenantIDFromToken.(string)
	tenantID := c.Param("tenantId")
	fmt.Printf("[DEBUG] Received tenantID: %s\n", tenantID)
	if tenantID != tenantIDStr {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID mismatch"})
		return
	}
	count, err := h.service.GetEvidenceCount(tenantID)
	fmt.Printf("[DEBUG] Evidence count for tenantID %s: %d\n", tenantID, count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evidence count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
