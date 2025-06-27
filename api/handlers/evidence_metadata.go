// File: handlers/metadata_handler.go

package handlers

import (
	"aegis-api/services_/evidence/metadata"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MetadataHandler struct {
	service metadata.MetadataService
}

func NewMetadataHandler(svc metadata.MetadataService) *MetadataHandler {
	return &MetadataHandler{service: svc}
}

func (h *MetadataHandler) UploadEvidence(c *gin.Context) {
	var req metadata.UploadEvidenceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	if err := h.service.UploadEvidence(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload evidence", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Evidence uploaded successfully"})
}
