package handlers

import (
	reportshared "aegis-api/services_/report/shared"
	"net/http"

	"aegis-api/services_/report"
	"aegis-api/services_/report/report_ai_assistance"

	"github.com/gin-gonic/gin"
)

// SuggestSectionPOST godoc
// @Summary Get AI suggestion for a section (POST)
// @Description Generate an AI suggestion for a specific report section using provided context
// @Tags reports, ai
// @Accept json
// @Produce json
// @Param reportID path string true "Report ID"
// @Param sectionID path string true "Report Section ID"
// @Param body body map[string]interface{} true "Context payload"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports/ai/{reportID}/sections/{sectionID}/suggest [post]
func (h *ReportAIHandler) SuggestSectionPOST(c *gin.Context) {
	sectionIDStr := c.Param("sectionID")
	reportIDStr := c.Param("reportID")
	if sectionIDStr == "" || reportIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing reportID or sectionID"})
		return
	}

	// Validate ObjectId: must be 24 hex characters
	if len(sectionIDStr) != 24 || len(reportIDStr) != 24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDs must be 24-character hex strings (ObjectId)"})
		return
	}
	for _, cHex := range sectionIDStr + reportIDStr {
		if !((cHex >= '0' && cHex <= '9') || (cHex >= 'a' && cHex <= 'f') || (cHex >= 'A' && cHex <= 'F')) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IDs must be valid hex strings"})
			return
		}
	}

	var contextPayload map[string]interface{}
	if err := c.ShouldBindJSON(&contextPayload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid context payload"})
		return
	}

	// Pass contextPayload to the AI service (update service to accept context if needed)
	suggestion, err := h.Service.GenerateSectionSuggestion(c.Request.Context(), reportIDStr, sectionIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestion": suggestion})
}

// ReportAIHandler holds reference to the services
type ReportAIHandler struct {
	Service       report_ai_assistance.ReportAIService
	ReportService report.ReportService // injected so we can fetch reports
}

// NewReportAIHandler creates a new handler
func NewReportAIHandler(service report_ai_assistance.ReportAIService, reportService report.ReportService) *ReportAIHandler {
	return &ReportAIHandler{
		Service:       service,
		ReportService: reportService,
	}
}

// SuggestSection godoc
// @Summary Get AI suggestion for a section
// @Description Generate an AI suggestion for a specific report section
// @Tags reports, ai
// @Accept json
// @Produce json
// @Param sectionID path string true "Report Section ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports/ai/sections/{sectionID}/suggest [get]
func (h *ReportAIHandler) SuggestSection(c *gin.Context) {
	sectionIDStr := c.Param("sectionID")
	reportIDStr := c.Param("reportID")
	if sectionIDStr == "" || reportIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing reportID or sectionID"})
		return
	}

	// Validate ObjectId: must be 24 hex characters
	if len(sectionIDStr) != 24 || len(reportIDStr) != 24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IDs must be 24-character hex strings (ObjectId)"})
		return
	}

	// Optionally, check if they're valid hex
	for _, cHex := range sectionIDStr + reportIDStr {
		if !((cHex >= '0' && cHex <= '9') || (cHex >= 'a' && cHex <= 'f') || (cHex >= 'A' && cHex <= 'F')) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "IDs must be valid hex strings"})
			return
		}
	}

	suggestion, err := h.Service.GenerateSectionSuggestion(c.Request.Context(), reportIDStr, sectionIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"suggestion": suggestion})
}

// SubmitFeedback godoc
// @Summary Submit feedback on an AI suggestion
// @Description Submit user feedback (accept/reject/edit) for an AI suggestion
// @Tags reports, ai
// @Accept json
// @Produce json
// @Param sectionID path string true "Report Section ID"
// @Param body body report_ai_assistance.AIFeedback true "Feedback payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports/ai/sections/{sectionID}/feedback [post]
func (h *ReportAIHandler) SubmitFeedback(c *gin.Context) {
	var feedback report_ai_assistance.AIFeedback
	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	if err := h.Service.SaveFeedback(c.Request.Context(), feedback); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "feedback recorded"})
}

// GenerateReferences godoc
// @Summary Generate references for a report section
// @Description Use AI to suggest references for a given section of a report
// @Tags reports, ai
// @Accept json
// @Produce json
// @Param reportID path string true "Report ID"
// @Param sectionID path string true "Report Section ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports/ai/{reportID}/sections/{sectionID}/references [get]
func (h *ReportAIHandler) GenerateReferences(c *gin.Context) {
	reportIDStr := c.Param("reportID")
	sectionIDStr := c.Param("sectionID")

	if reportIDStr == "" || sectionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing reportID or sectionID"})
		return
	}

	// Fetch the report (if needed, you may need to update this logic for MongoDB)
	// If your ReportService expects a hex string, update its signature accordingly
	reportObj, err := h.ReportService.GetReportByID(c.Request.Context(), reportIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate references using the AI service
	// Convert *report.Report to *reportshared.Report
	reportShared := &reportshared.Report{
		ID:        reportObj.ID.String(),
		Title:     reportObj.Name, // Use Name if Title is not present
		CreatedAt: reportObj.CreatedAt,
		// Add other fields as needed
	}
	refs, err := h.Service.GenerateSectionReferences(c.Request.Context(), sectionIDStr, reportShared)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"references": refs})
}
