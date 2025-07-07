package handlers

import (
	"aegis-api/services_/evidence/metadata"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MetadataHandler struct {
	service metadata.MetadataService
}

func NewMetadataHandler(svc metadata.MetadataService) *MetadataHandler {
	return &MetadataHandler{service: svc}
}

func (h *MetadataHandler) UploadEvidence(c *gin.Context) {
	caseIDStr := c.PostForm("caseId")
	uploadedByStr := c.PostForm("uploadedBy")
	fileType := c.PostForm("fileType") // optional metadata field

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid caseId format", "details": err.Error()})
		return
	}
	uploadedBy, err := uuid.Parse(uploadedByStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uploadedBy format", "details": err.Error()})
		return
	}

	log.Println("[DEBUG] POST form caseId:", c.PostForm("caseId"))
	log.Println("[DEBUG] POST form uploadedBy:", c.PostForm("uploadedBy"))

	form, err := c.MultipartForm()
	if err != nil {
		log.Printf("[ERROR] MultipartForm parse failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form", "details": err.Error()})
		return
	}
	log.Printf("[DEBUG] Form file keys: %v", form.File)

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file", "details": err.Error()})
			return
		}

		req := metadata.UploadEvidenceRequest{
			CaseID:     caseID,
			UploadedBy: uploadedBy,
			Filename:   fileHeader.Filename,
			FileType:   fileType,
			FileSize:   fileHeader.Size,
			FileData:   file, // pass io.Reader directly
		}

		if err := h.service.UploadEvidence(req); err != nil {
			log.Printf("❌ UploadEvidence failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to upload evidence",
				"details": err.Error(),
			})
			return
		}
		file.Close()
		log.Printf("✅ Successfully uploaded evidence file: %s for case: %s", fileHeader.Filename, caseID)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Evidence uploaded successfully"})
}
