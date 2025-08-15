package handlers

import (
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
	tenantID := c.Query("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tenantId is required"})
		return
	}

	count, err := h.service.GetEvidenceCount(tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evidence count"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"count": count})
}
