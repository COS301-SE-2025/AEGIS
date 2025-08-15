package handlers

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"aegis-api/services_/report"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ReportHandler handles HTTP requests for reports.
type ReportHandler struct {
	ReportService report.ReportService
}

func NewReportHandler(s report.ReportService) *ReportHandler {
	return &ReportHandler{ReportService: s}
}

// GenerateReport creates a new report for a case.
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	caseIDStr := c.Param("caseID")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authorized"})
		return
	}

	examinerUUID, err := uuid.Parse(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID format"})
		return
	}

	report, err := h.ReportService.GenerateReport(c.Request.Context(), caseID, examinerUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reportID": report.ID,
		"status":   "Report generated successfully",
	})
}

// GetReportByID retrieves a report with metadata and content.
func (h *ReportHandler) GetReportByID(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	report, err := h.ReportService.DownloadReport(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	c.JSON(http.StatusOK, report)
}

// Error writer with a stable shape
func writeError(c *gin.Context, status int, code, msg string) {
	c.AbortWithStatusJSON(status, gin.H{
		"error": gin.H{
			"code":    code,
			"message": msg,
		},
	})
}

func logWithCtx(level, msg string, c *gin.Context, kv map[string]any) {
	log.Printf("%s %s path=%s method=%s ip=%s ctx=%v",
		strings.ToUpper(level), msg, c.FullPath(), c.Request.Method, c.ClientIP(), kv)
}

func (h *ReportHandler) UpdateSectionContent(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		logWithCtx("warn", "invalid report ID", c, map[string]any{"reportID": reportIDStr, "err": err.Error()})
		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		logWithCtx("warn", "invalid section ID", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionIDStr, "err": err.Error()})
		writeError(c, http.StatusBadRequest, "invalid_section_id", "invalid section ID")
		return
	}

	var req struct {
		Content string `json:"content"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Content) == "" {
		logWithCtx("warn", "invalid body", c, map[string]any{
			"reportID":  reportUUID.String(),
			"sectionID": sectionID.Hex(),
			"err":       err,
		})
		writeError(c, http.StatusBadRequest, "invalid_body", "content is required")
		return
	}

	if err := h.ReportService.UpdateCustomSectionContent(c.Request.Context(), reportUUID, sectionID, req.Content); err != nil {
		// Prefer sentinel errors from your service/repo (see note below).
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			logWithCtx("info", "report not found", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
			return
		case errors.Is(err, report.ErrSectionNotFound):
			logWithCtx("info", "section not found", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
			writeError(c, http.StatusNotFound, "section_not_found", "section not found")
			return
		default:
			// Fallback if you haven't added sentinel errors yet:
			low := strings.ToLower(err.Error())
			if strings.Contains(low, "not found") {
				logWithCtx("info", "resource not found", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
				writeError(c, http.StatusNotFound, "not_found", "resource not found")
				return
			}
			logWithCtx("error", "update section failed", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex(), "err": err.Error()})
			writeError(c, http.StatusInternalServerError, "update_failed", "failed to update section content")
			return
		}
	}

	logWithCtx("info", "section content updated", c, map[string]any{"reportID": reportUUID.String(), "sectionID": sectionID.Hex()})
	c.Status(http.StatusNoContent)
}

// DownloadReportPDF returns the report as PDF.
func (h *ReportHandler) DownloadReportPDF(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	pdfBytes, err := h.ReportService.DownloadReportAsPDF(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate PDF"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=report_"+reportIDStr+".pdf")
	c.Data(http.StatusOK, "application/pdf", pdfBytes)
}

// DownloadReportJSON returns the report as JSON.
func (h *ReportHandler) DownloadReportJSON(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	jsonBytes, err := h.ReportService.DownloadReportAsJSON(c.Request.Context(), reportID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate JSON"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=report_"+reportIDStr+".json")
	c.Data(http.StatusOK, "application/json", jsonBytes)
}

// DeleteReport deletes a report by ID.
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	if err := h.ReportService.DeleteReportByID(c.Request.Context(), reportID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete report"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "report deleted"})
}

// AddSection handles adding a new custom section to a report
func (h *ReportHandler) AddSection(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Order   int    `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err = h.ReportService.AddCustomSection(c.Request.Context(), reportUUID, req.Title, req.Content, req.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add section"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "section added successfully"})
}

// DeleteSection handles deleting a section from a report
func (h *ReportHandler) DeleteSection(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid section ID"})
		return
	}

	err = h.ReportService.DeleteCustomSection(c.Request.Context(), reportUUID, sectionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete section"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "section deleted successfully"})
}

func (h *ReportHandler) GetReportsByCaseID(c *gin.Context) {
	caseIDStr := c.Param("caseID")
	caseUUID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	reports, err := h.ReportService.GetReportsByCaseID(c.Request.Context(), caseUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch reports"})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// UpdateSectionContent updates the content of a specific section
// func (h *ReportHandler) UpdateSectionContent(c *gin.Context) {
// 	reportIDStr := c.Param("reportID")
// 	reportUUID, err := uuid.Parse(reportIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
// 		return
// 	}

// 	sectionIDStr := c.Param("sectionID")
// 	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
// 	if err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid section ID"})
// 		return
// 	}

// 	var req struct {
// 		Content string `json:"content"`
// 	}

// 	if err := c.ShouldBindJSON(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
// 		return
// 	}

// 	err = h.ReportService.UpdateReportSection(c.Request.Context(), reportUUID, sectionID, req.Content)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update section content"})
// 		return
// 	}

// 	c.JSON(http.StatusOK, gin.H{"status": "section content updated successfully"})
// }

// UpdateSectionTitle updates the title of a specific section
func (h *ReportHandler) UpdateSectionTitle(c *gin.Context) {
	// Parse report UUID
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	// Parse section ID (Mongo ObjectID)
	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid section ID"})
		return
	}

	// Bind request body
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Call service method
	err = h.ReportService.UpdateSectionTitle(c.Request.Context(), reportUUID, sectionID, req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update section title"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "section title updated successfully"})
}

// ReorderSection updates the order of a section
func (h *ReportHandler) ReorderSection(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid section ID"})
		return
	}

	var req struct {
		NewOrder int `json:"order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	err = h.ReportService.ReorderSection(c.Request.Context(), reportUUID, sectionID, req.NewOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to reorder section"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "section reordered successfully"})
}
