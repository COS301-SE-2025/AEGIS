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

type DownloadHandler struct {
	service     *download.Service
	auditLogger *auditlog.AuditLogger
}

func NewDownloadHandler(svc *download.Service, auditLogger *auditlog.AuditLogger) *DownloadHandler {
	return &DownloadHandler{
		service:     svc,
		auditLogger: auditLogger,
	}
}

func (h *DownloadHandler) Download(c *gin.Context) {
	idParam := c.Param("id")
	evidenceID, err := uuid.Parse(idParam)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "DOWNLOAD_EVIDENCE",
			Actor:       auditlog.MakeActor(c),
			Target:      auditlog.Target{Type: "evidence", ID: idParam},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Invalid UUID format for evidence ID",
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidence ID"})
		return
	}

	filename, stream, filetype, err := h.service.DownloadEvidence(evidenceID)
	if err != nil {
		log.Printf("❌ Download failed: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "DOWNLOAD_EVIDENCE",
			Actor:       auditlog.MakeActor(c),
			Target:      auditlog.Target{Type: "evidence", ID: evidenceID.String()},
			Service:     "evidence",
			Status:      "FAILED",
			Description: "Download failed: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to download evidence", "details": err.Error()})
		return
	}
	defer stream.Close()

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "DOWNLOAD_EVIDENCE",
		Actor:       auditlog.MakeActor(c),
		Target:      auditlog.Target{Type: "evidence", ID: evidenceID.String()},
		Service:     "evidence",
		Status:      "SUCCESS",
		Description: "Evidence downloaded successfully: " + filename,
	})

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", filetype)
	c.Status(http.StatusOK)
	io.Copy(c.Writer, stream)
}
