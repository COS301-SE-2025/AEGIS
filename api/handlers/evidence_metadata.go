package handlers

import (
	"aegis-api/cache"
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
	cacheClient cache.Client
}

func NewMetadataHandler(svc metadata.MetadataService, logger *auditlog.AuditLogger, c cache.Client) *MetadataHandler {
	return &MetadataHandler{
		service:     svc,
		auditLogger: logger,
		cacheClient: c,
	}
}

func (h *MetadataHandler) UploadEvidence(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	tenantID, tenantExists := c.Get("tenantID")
	teamID, teamExists := c.Get("teamID")

	if !tenantExists || !teamExists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Tenant or Team context missing"})
		return
	}

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	}

	// Extract form data
	caseIDStr := c.PostForm("caseId")
	uploadedByStr := c.PostForm("uploadedBy")
	fileType := c.PostForm("fileType") // optional

	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid caseId format"})
		return
	}

	uploadedBy, err := uuid.Parse(uploadedByStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid uploadedBy format"})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid multipart form"})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	for _, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to open uploaded file"})
			return
		}

		req := metadata.UploadEvidenceRequest{
			CaseID:     caseID,
			UploadedBy: uploadedBy,
			Filename:   fileHeader.Filename,
			FileType:   fileType,
			FileSize:   fileHeader.Size,
			FileData:   file,
			TenantID:   uuid.MustParse(tenantID.(string)),
			TeamID:     uuid.MustParse(teamID.(string)),
		}

		if err := h.service.UploadEvidence(req); err != nil {
			file.Close()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		file.Close()
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPLOAD_EVIDENCE_METADATA",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "evidence_metadata_upload",
			ID:   caseID.String(),
			AdditionalInfo: map[string]string{
				"file_count": fmt.Sprintf("%d", len(files)),
				"tenant_id":  tenantID.(string),
				"team_id":    teamID.(string),
			},
		},
		Service:     "evidence",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Uploaded %d evidence files for case %s", len(files), caseID),
	})
	// after all files are successfully saved
	ctx := c.Request.Context()
	tenantStr := tenantID.(string)
	caseIDStr = caseID.String()

	// Blow all list variants for that case (different q=sha)
	cache.InvalidateEvidenceListsForCase(ctx, h.cacheClient, tenantStr, caseIDStr)

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
	var emailStr string
	if emailVal, ok := c.Get("email"); ok && emailVal != nil {
		emailStr, _ = emailVal.(string)
	}
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     emailStr,
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

// VerifyEvidenceChain verifies the hash chain integrity for a given evidence ID.
func (h *MetadataHandler) VerifyEvidenceChain(c *gin.Context) {
	log.Printf("[DEBUG] Handler entered for verify-chain")

	idStr := c.Param("evidence_id")
	log.Printf("[DEBUG] Received verify-chain request for evidence_id: %s\n", idStr)

	evidenceID, err := uuid.Parse(idStr)
	if err != nil {
		log.Printf("[ERROR] Invalid evidence ID format: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidence ID format"})
		return
	}

	log.Printf("[DEBUG] Calling service.VerifyEvidenceLogChain for evidenceID: %s\n", evidenceID.String())
	ok, details, err := h.service.VerifyEvidenceLogChain(evidenceID)
	if err != nil {
		log.Printf("[ERROR] Verification failed: %v | Details: %s\n", err, details)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Verification failed",
			"details": err.Error(),
		})
		return
	}
	log.Printf("[DEBUG] Verification result: valid=%v, details=%s\n", ok, details)
	fmt.Printf("[DEBUG] Verification result: valid=%v, details=%s\n", ok, details)
	c.JSON(http.StatusOK, gin.H{
		"evidence_id": evidenceID.String(),
		"valid":       ok,
		"details":     details,
	})
}
