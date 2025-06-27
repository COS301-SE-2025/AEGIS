package handlers

import (
	"aegis-api/services_/evidence/upload"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	service *upload.Service
}

func NewUploadHandler(svc *upload.Service) *UploadHandler {
	return &UploadHandler{service: svc}
}

func (h *UploadHandler) Upload(c *gin.Context) {
	type UploadRequest struct {
		Path string `json:"path"`
	}

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request", "details": err.Error()})
		return
	}

	cid, err := h.service.UploadFile(req.Path)
	if err != nil {
		// Add this log line:
		log.Printf("❌ Upload failed: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cid": cid})
}
