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
	// parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form", "details": err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	// Example: you could also parse additional fields
	caseID := c.PostForm("caseId")
	log.Printf("📦 Uploading %d files for case ID: %s", len(files), caseID)

	uploaded := []string{}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file", "details": err.Error()})
			return
		}
		defer file.Close()

		cid, err := h.service.UploadFile(file)
		if err != nil {
			log.Printf("❌ Upload to IPFS failed: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to IPFS", "details": err.Error()})
			return
		}

		log.Printf("✅ Uploaded %s to IPFS CID: %s", fileHeader.Filename, cid)
		uploaded = append(uploaded, cid)

		// You could also trigger metadata saving here:
		// h.metadataService.SaveEvidenceMetadata(caseID, fileHeader.Filename, cid, fileHeader.Size)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"cids":    uploaded,
	})
}
