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

// GetEvidenceByID retrieves evidence metadata by its ID.
func (h *MetadataHandler) GetEvidenceByID(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_EVIDENCE_BY_ID",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence",
				ID:   idStr,
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid evidence ID format: " + err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidence ID format", "details": err.Error()})
		return
	}

	evidence, err := h.service.FindEvidenceByID(id)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_EVIDENCE_BY_ID",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence",
				ID:   idStr,
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Evidence not found or retrieval failed: " + err.Error(),
		})
		c.JSON(http.StatusNotFound, gin.H{"error": "Evidence not found", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_EVIDENCE_BY_ID",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "evidence",
			ID:   evidence.ID.String(),
		},
		Service:     "evidence",
		Status:      "SUCCESS",
		Description: "Successfully retrieved evidence by ID",
	})

	c.JSON(http.StatusOK, evidence)
}

// / GetEvidenceByCaseID retrieves all evidence records for a given case.
func (h *MetadataHandler) GetEvidenceByCaseID(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	caseIDStr := c.Param("case_id")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_EVIDENCE_BY_CASE_ID",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseIDStr,
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid case ID format: " + err.Error(),
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid case ID format", "details": err.Error()})
		return
	}

	evidences, err := h.service.GetEvidenceByCaseID(caseID)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_EVIDENCE_BY_CASE_ID",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case",
				ID:   caseID.String(),
			},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Failed to retrieve evidence list: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve evidence", "details": err.Error()})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_EVIDENCE_BY_CASE_ID",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case",
			ID:   caseID.String(),
			AdditionalInfo: map[string]string{
				"result_count": fmt.Sprintf("%d", len(evidences)),
			},
		},
		Service:     "evidence",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Successfully retrieved %d evidence records for case", len(evidences)),
	})

	c.JSON(http.StatusOK, evidences)
}
