package handlers

import (
	download "aegis-api/services_/evidence/evidence_download"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DownloadHandler struct {
	service *download.Service
}

func NewDownloadHandler(svc *download.Service) *DownloadHandler {
	return &DownloadHandler{service: svc}
}

func (h *DownloadHandler) Download(c *gin.Context) {
	idParam := c.Param("id")
	evidenceID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidence ID"})
		return
	}

	filename, stream, filetype, err := h.service.DownloadEvidence(evidenceID)
	if err != nil {
		log.Printf("❌ Download failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download evidence", "details": err.Error()})
		return
	}
	defer stream.Close()

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", filetype)
	c.Status(http.StatusOK)
	io.Copy(c.Writer, stream)
}
