package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

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
// handlers/report.go
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("caseID"))
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_case_id", "invalid case ID")
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "user not authorized")
		return
	}
	examinerID, err := uuid.Parse(userID.(string))
	if err != nil {
		writeError(c, http.StatusInternalServerError, "invalid_user_id", "invalid user ID format")
		return
	}

	tenantIDStr := c.GetString("tenantID")
	if tenantIDStr == "" {
		writeError(c, http.StatusUnauthorized, "tenant_missing", "tenant not found")
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_tenant_id", "invalid tenant id")
		return
	}

	var teamID uuid.UUID
	if s := c.GetString("teamID"); s != "" {
		if teamID, err = uuid.Parse(s); err != nil {
			writeError(c, http.StatusBadRequest, "invalid_team_id", "invalid team id")
			return
		}
	}

	rep, err := h.ReportService.GenerateReport(c.Request.Context(), caseID, examinerID, tenantID, teamID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "generate_failed", "failed to generate report")
		return
	}

	c.JSON(http.StatusOK, gin.H{"reportID": rep.ID, "status": "Report generated successfully"})
}

// GetReportByID retrieves a report with metadata and content.
func (h *ReportHandler) GetReportByID(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	rep, err := h.ReportService.DownloadReport(c.Request.Context(), reportID)
	if err != nil {
		logWithCtx("info", "report not found", c, map[string]any{"reportID": reportIDStr, "err": err.Error()})
		writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		return
	}

	c.JSON(http.StatusOK, rep)
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
	// ✅ allow empty content (users may clear a section)
	if err := c.ShouldBindJSON(&req); err != nil {
		logWithCtx("warn", "invalid body", c, map[string]any{
			"reportID":  reportUUID.String(),
			"sectionID": sectionID.Hex(),
			"err":       err,
		})
		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}

	if err := h.ReportService.UpdateCustomSectionContent(c.Request.Context(), reportUUID, sectionID, req.Content); err != nil {
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

// -----------------
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
// DownloadReportJSON returns the report as JSON.
func (h *ReportHandler) DownloadReportJSON(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	jsonBytes, err := h.ReportService.DownloadReportAsJSON(c.Request.Context(), reportID)
	if err != nil {
		logWithCtx("error", "download json failed", c, map[string]any{"reportID": reportIDStr, "err": err.Error()})
		writeError(c, http.StatusInternalServerError, "generate_json_failed", "failed to generate JSON")
		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=report_"+reportIDStr+".json")
	c.Header("Cache-Control", "no-store")
	c.Data(http.StatusOK, "application/json", jsonBytes)
}

// DeleteReport deletes a report by ID.
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	if err := h.ReportService.DeleteReportByID(c.Request.Context(), reportID); err != nil {
		// map known errors if you expose them from the repo/service
		if errors.Is(err, report.ErrReportNotFound) {
			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
			return
		}
		logWithCtx("error", "delete report failed", c, map[string]any{"reportID": reportIDStr, "err": err.Error()})
		writeError(c, http.StatusInternalServerError, "delete_failed", "failed to delete report")
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "report deleted"})
}

// AddSection handles adding a new custom section to a report.
func (h *ReportHandler) AddSection(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}

	var req struct {
		Title   string `json:"title"`
		Content string `json:"content"`
		Order   int    `json:"order"` // 1-based; service can clamp/append if <=0
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(c, http.StatusBadRequest, "invalid_title", "title is required")
		return
	}
	// Optional title length guardrail
	if len(req.Title) > 200 {
		writeError(c, http.StatusBadRequest, "invalid_title", "title is too long")
		return
	}

	if err := h.ReportService.AddCustomSection(c.Request.Context(), reportUUID, req.Title, req.Content, req.Order); err != nil {
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		case errors.Is(err, report.ErrInvalidInput):
			writeError(c, http.StatusBadRequest, "invalid_input", "invalid input")
		default:
			logWithCtx("error", "add section failed", c, map[string]any{
				"reportID": reportUUID.String(),
				"title":    req.Title,
				"order":    req.Order,
				"err":      err.Error(),
			})
			writeError(c, http.StatusInternalServerError, "add_section_failed", "failed to add section")
		}
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

func (h *ReportHandler) UpdateSectionTitle(c *gin.Context) {
	// IDs
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}
	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_section_id", "invalid section ID")
		return
	}

	// Body
	var req struct {
		Title string `json:"title"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		writeError(c, http.StatusBadRequest, "invalid_title", "title is required")
		return
	}
	if len(title) > 200 { // guardrails; adjust if you want
		writeError(c, http.StatusBadRequest, "invalid_title", "title is too long")
		return
	}

	// Service
	if err := h.ReportService.UpdateSectionTitle(c.Request.Context(), reportUUID, sectionID, title); err != nil {
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		case errors.Is(err, report.ErrSectionNotFound):
			writeError(c, http.StatusNotFound, "section_not_found", "section not found")
		case errors.Is(err, report.ErrInvalidInput):
			writeError(c, http.StatusBadRequest, "invalid_input", "invalid input")
		default:
			writeError(c, http.StatusInternalServerError, "update_failed", "failed to update section title")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "section title updated successfully"})
}

func (h *ReportHandler) ReorderSection(c *gin.Context) {
	// IDs
	reportIDStr := c.Param("reportID")
	reportUUID, err := uuid.Parse(reportIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_report_id", "invalid report ID")
		return
	}
	sectionIDStr := c.Param("sectionID")
	sectionID, err := primitive.ObjectIDFromHex(sectionIDStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_section_id", "invalid section ID")
		return
	}

	// Body
	var req struct {
		NewOrder int `json:"order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid_body", "invalid request body")
		return
	}
	if req.NewOrder < 1 {
		writeError(c, http.StatusBadRequest, "invalid_order", "order must be >= 1")
		return
	}

	// Service (rename here if your service method is ReorderCustomSection)
	if err := h.ReportService.ReorderSection(c.Request.Context(), reportUUID, sectionID, req.NewOrder); err != nil {
		switch {
		case errors.Is(err, report.ErrReportNotFound), errors.Is(err, report.ErrMongoReportNotFound):
			writeError(c, http.StatusNotFound, "report_not_found", "report not found")
		case errors.Is(err, report.ErrSectionNotFound):
			writeError(c, http.StatusNotFound, "section_not_found", "section not found")
		case errors.Is(err, report.ErrInvalidInput):
			writeError(c, http.StatusBadRequest, "invalid_input", "invalid input")
		default:
			// Optional: if you implement optimistic concurrency and return a "conflict" error,
			// map it to 409 here.
			low := strings.ToLower(err.Error())
			if strings.Contains(low, "conflict") {
				writeError(c, http.StatusConflict, "conflict", "the section was modified by someone else")
				return
			}
			writeError(c, http.StatusInternalServerError, "reorder_failed", "failed to reorder section")
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "section reordered successfully"})
}

// GET /api/v1/reports/recent?limit=6&mine=true&caseId=<uuid>&status=<string>
func (h *ReportHandler) GetRecentReports(c *gin.Context) {
	// auth: same style as GenerateReport
	userIDVal, ok := c.Get("userID")
	if !ok {
		writeError(c, http.StatusUnauthorized, "unauthorized", "user not authorized")
		return
	}
	examinerID, err := uuid.Parse(userIDVal.(string))
	if err != nil {
		writeError(c, http.StatusInternalServerError, "invalid_user_id", "invalid user id format")
		return
	}

	// query params
	limitStr := c.DefaultQuery("limit", "6")
	mineStr := c.DefaultQuery("mine", "true")
	caseIDStr := c.Query("caseId")
	statusStr := strings.TrimSpace(c.Query("status"))

	limit := 6
	if v, err := strconv.Atoi(limitStr); err == nil && v > 0 {
		limit = v
	}
	mine := strings.EqualFold(mineStr, "true")

	var caseID *uuid.UUID
	if caseIDStr != "" {
		if cid, err := uuid.Parse(caseIDStr); err == nil {
			caseID = &cid
		} else {
			writeError(c, http.StatusBadRequest, "invalid_case_id", "invalid caseId")
			return
		}
	}

	var status *string
	if statusStr != "" {
		status = &statusStr
	}

	// NEW: tenant & team from auth middleware (strings)
	tidStr := c.GetString("tenantID")
	if tidStr == "" {
		writeError(c, http.StatusUnauthorized, "tenant_missing", "tenant not found")
		return
	}
	tenantID, err := uuid.Parse(tidStr)
	if err != nil {
		writeError(c, http.StatusBadRequest, "invalid_tenant_id", "invalid tenant id")
		return
	}

	var teamIDPtr *uuid.UUID
	if teamStr := c.GetString("teamID"); teamStr != "" {
		if t, err := uuid.Parse(teamStr); err == nil {
			teamIDPtr = &t
		} else {
			writeError(c, http.StatusBadRequest, "invalid_team_id", "invalid team id")
			return
		}
	}

	opts := report.RecentReportsOptions{
		Limit:      limit,
		MineOnly:   mine,
		ExaminerID: examinerID,
		CaseID:     caseID,
		Status:     status,
		TenantID:   tenantID, // ← NEW
		TeamID:     teamIDPtr,
	}

	items, err := h.ReportService.ListRecentReports(c.Request.Context(), opts)
	if err != nil {
		logWithCtx("error", "list recent reports failed", c, map[string]any{"err": err.Error()})
		writeError(c, http.StatusInternalServerError, "recent_reports_failed", "failed to load recent reports")
		return
	}

	// shape the response to what your React expects
	type row struct {
		ID           uuid.UUID `json:"id"`
		Title        string    `json:"title"`
		Status       string    `json:"status"`
		LastModified string    `json:"lastModified"` // ISO time
	}

	resp := make([]row, len(items))
	for i, it := range items {
		resp[i] = row{
			ID:           it.ID,
			Title:        it.Title,
			Status:       it.Status,
			LastModified: it.LastModified.Format(time.RFC3339),
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ReportHandler) UpdateReportName(c *gin.Context) {
	rid, err := uuid.Parse(c.Param("reportID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid reportID"})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	updated, err := h.ReportService.UpdateReportName(c.Request.Context(), rid, req.Name)
	if err != nil {
		switch {
		case errors.Is(err, report.ErrInvalidReportName):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		case errors.Is(err, report.ErrReportNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update name"})
		}
		return
	}

	c.JSON(http.StatusOK, updated) // full Report JSON
}
