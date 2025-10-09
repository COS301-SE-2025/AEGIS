package handlers

import (
	"aegis-api/services_/auditlog"
	download "aegis-api/services_/evidence/evidence_download"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// DownloadService defines the interface for the download service
type DownloadService interface {
	DownloadEvidence(evidenceID uuid.UUID) (string, io.ReadCloser, string, error)
}

// AuditLogger defines the interface for audit logging
type AuditLogger interface {
	Log(c *gin.Context, log auditlog.AuditLog) error
}

type DownloadHandler struct {
	service     DownloadService
	auditLogger AuditLogger
}

// NewDownloadHandler creates a new download handler with concrete types
func NewDownloadHandler(svc *download.Service, auditLogger *auditlog.AuditLogger) *DownloadHandler {
	return &DownloadHandler{
		service:     svc,
		auditLogger: auditLogger,
	}
}

// NewDownloadHandlerWithInterfaces creates a new download handler with interface types (for testing)
func NewDownloadHandlerWithInterfaces(svc DownloadService, auditLogger AuditLogger) *DownloadHandler {
	return &DownloadHandler{
		service:     svc,
		auditLogger: auditLogger,
	}
}

func (h *DownloadHandler) Download(c *gin.Context) {
	idParam := c.Param("id")
	evidenceID, err := uuid.Parse(idParam)
	if err != nil {
		if logErr := h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "DOWNLOAD_EVIDENCE",
			Actor:       auditlog.MakeActor(c),
			Target:      auditlog.Target{Type: "evidence", ID: idParam},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid UUID format for evidence ID",
		}); logErr != nil {
			log.Printf("Failed to log audit: %v", logErr)
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidence ID"})
		return
	}

	filename, stream, filetype, err := h.service.DownloadEvidence(evidenceID)
	if err != nil {
		log.Printf("❌ Download failed: %v\n", err)

		if logErr := h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "DOWNLOAD_EVIDENCE",
			Actor:       auditlog.MakeActor(c),
			Target:      auditlog.Target{Type: "evidence", ID: evidenceID.String()},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Download failed: " + err.Error(),
		}); logErr != nil {
			log.Printf("Failed to log audit: %v", logErr)
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download evidence", "details": err.Error()})
		return
	}
	defer stream.Close()

	if logErr := h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "DOWNLOAD_EVIDENCE",
		Actor:       auditlog.MakeActor(c),
		Target:      auditlog.Target{Type: "evidence", ID: evidenceID.String()},
		Service:     "evidence",
		Status:      "SUCCESS",
		Description: "Evidence downloaded successfully: " + filename,
	}); logErr != nil {
		log.Printf("Failed to log audit: %v", logErr)
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", filetype)
	c.Status(http.StatusOK)
	io.Copy(c.Writer, stream)
}
