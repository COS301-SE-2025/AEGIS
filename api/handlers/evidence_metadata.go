package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/evidence/metadata"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MetadataHandler struct {
	service     metadata.MetadataService
	auditLogger *auditlog.AuditLogger
}

func NewMetadataHandler(svc metadata.MetadataService, logger *auditlog.AuditLogger) *MetadataHandler {
	return &MetadataHandler{
		service:     svc,
		auditLogger: logger,
	}
}

func (h *MetadataHandler) UploadEvidence(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	caseIDStr := c.PostForm("caseId")
	uploadedByStr := c.PostForm("uploadedBy")
	fileType := c.PostForm("fileType") // optional

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPLOAD_EVIDENCE_METADATA",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence_metadata_upload",
				ID:   caseIDStr,
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid caseId format: " + err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid caseId format", "details": err.Error()})
		return
	}

	uploadedBy, err := uuid.Parse(uploadedByStr)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPLOAD_EVIDENCE_METADATA",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence_metadata_upload",
				ID:   uploadedByStr,
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid uploadedBy format: " + err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uploadedBy format", "details": err.Error()})
		return
	}

	log.Println("[DEBUG] POST form caseId:", caseIDStr)
	log.Println("[DEBUG] POST form uploadedBy:", uploadedByStr)

	form, err := c.MultipartForm()
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPLOAD_EVIDENCE_METADATA",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence_metadata_upload",
				ID:   caseID.String(),
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid multipart form: " + err.Error(),
		})
		log.Printf("[ERROR] MultipartForm parse failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form", "details": err.Error()})
		return
	}
	log.Printf("[DEBUG] Form file keys: %v", form.File)

	files := form.File["files"]
	if len(files) == 0 {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPLOAD_EVIDENCE_METADATA",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence_metadata_upload",
				ID:   caseID.String(),
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "No files uploaded",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPLOAD_EVIDENCE_METADATA",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "evidence_metadata_upload",
					ID:   caseID.String(),
					AdditionalInfo: map[string]string{
						"filename": fileHeader.Filename,
					},
				},
				Service:     "evidence",
				Status:      "FAILED",
				Description: "Failed to open uploaded file: " + err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file", "details": err.Error()})
			return
		}

		req := metadata.UploadEvidenceRequest{
			CaseID:     caseID,
			UploadedBy: uploadedBy,
			Filename:   fileHeader.Filename,
			FileType:   fileType,
			FileSize:   fileHeader.Size,
			FileData:   file,
		}

		if err := h.service.UploadEvidence(req); err != nil {
			h.auditLogger.Log(c, auditlog.AuditLog{
				Action: "UPLOAD_EVIDENCE_METADATA",
				Actor:  actor,
				Target: auditlog.Target{
					Type: "evidence_metadata_upload",
					ID:   caseID.String(),
					AdditionalInfo: map[string]string{
						"filename": fileHeader.Filename,
					},
				},
				Service:     "evidence",
				Status:      "FAILED",
				Description: "Failed to upload evidence metadata: " + err.Error(),
			})
			log.Printf("❌ UploadEvidence failed: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to upload evidence",
				"details": err.Error(),
			})
			file.Close()
			return
		}
		file.Close()
		log.Printf("✅ Successfully uploaded evidence file: %s for case: %s", fileHeader.Filename, caseID)
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPLOAD_EVIDENCE_METADATA",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "evidence_metadata_upload",
			ID:   caseID.String(),
			AdditionalInfo: map[string]string{
				"file_count": fmt.Sprintf("%d", len(files)),
			},
		},
		Service:     "evidence",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Uploaded %d evidence files for case %s", len(files), caseID),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Evidence uploaded successfully"})
}
