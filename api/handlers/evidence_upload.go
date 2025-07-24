package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/evidence/upload"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	service     *upload.Service
	auditLogger *auditlog.AuditLogger
}

func NewUploadHandler(svc *upload.Service, logger *auditlog.AuditLogger) *UploadHandler {
	return &UploadHandler{
		service:     svc,
		auditLogger: logger,
	}
}
func (h *UploadHandler) Upload(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     c.GetHeader("X-User-Email"),
	}

	form, err := c.MultipartForm()
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPLOAD_EVIDENCE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence_upload",
				ID:   "",
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid multipart form: " + err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form", "details": err.Error()})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPLOAD_EVIDENCE",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence_upload",
				ID:   "",
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "No files uploaded",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	caseID := c.PostForm("caseId")
	log.Printf("📦 Uploading %d files for case ID: %s", len(files), caseID)

	uploaded := []string{}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPLOAD_EVIDENCE",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "evidence_upload",
					ID:   "",
				},
				Service:     "evidence",
				Status:      "FAILED",
				Description: "Failed to open uploaded file: " + err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file", "details": err.Error()})
			return
		}
		defer file.Close()

		cid, err := h.service.UploadFile(file)
		if err != nil {
			log.Printf("❌ Upload to IPFS failed: %v\n", err)
			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPLOAD_EVIDENCE",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "evidence_upload",
					ID:   "",
					AdditionalInfo: map[string]string{
						"filename": fileHeader.Filename,
					},
				},
				Service:     "evidence",
				Status:      "FAILED",
				Description: "Upload to IPFS failed: " + err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload to IPFS", "details": err.Error()})
			return
		}

		log.Printf("✅ Uploaded %s to IPFS CID: %s", fileHeader.Filename, cid)
		uploaded = append(uploaded, cid)
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPLOAD_EVIDENCE",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "evidence_upload",
			ID:   caseID,
			AdditionalInfo: map[string]string{
				"file_count": fmt.Sprintf("%d", len(uploaded)),
			},
		},
		Service:     "evidence",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Uploaded %d files to IPFS for case %s", len(uploaded), caseID),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"cids":    uploaded,
	})
}
