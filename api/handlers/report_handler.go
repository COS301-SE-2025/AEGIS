package handlers

import (
	"aegis-api/services_/report"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ReportHandler is the handler for managing reports.
type ReportHandler struct {
	ReportService report.ReportInterface // Correct the type to ReportService
}

// NewReportHandler creates a new instance of ReportHandler.
func NewReportHandler(reportService report.ReportInterface) *ReportHandler {
	return &ReportHandler{
		ReportService: reportService,
	}
}

// GenerateReport creates a new report for a case.
func (h *ReportHandler) GenerateReport(c *gin.Context) {
	// Extract the case ID from the URL parameters
	caseIDStr := c.Param("caseID")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	// Extract examinerID from the context (this assumes the userID is stored in context middleware)
	examinerID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authorized"})
		return
	}

	// Pass the context and other arguments to GenerateReport
	report, err := h.ReportService.GenerateReport(c.Request.Context(), caseID, examinerID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate report"})
		return
	}

	// Return the response with the generated report details
	c.JSON(http.StatusOK, gin.H{
		"reportID": report.ID,
		"status":   "Report generated successfully",
	})
}

// GetReportByID retrieves a report by ID.
func (h *ReportHandler) GetReportByID(c *gin.Context) {
	// Extract the report ID from the URL parameters
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	// Get the report using the ReportService
	report, err := h.ReportService.GetReportByID(c.Request.Context(), reportID) // Pass context here
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	// Return the report details
	c.JSON(http.StatusOK, gin.H{
		"reportID":     report.ID,
		"scope":        report.Scope,
		"objectives":   report.Objectives,
		"status":       report.Status,
		"dateExamined": report.DateExamined,
	})
}

// UpdateReport updates an existing report.
func (h *ReportHandler) UpdateReport(c *gin.Context) {
	// Bind the incoming JSON request to a Report struct
	var report report.Report
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Update the report using the ReportService
	err := h.ReportService.UpdateReport(c.Request.Context(), &report) // Pass context here
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update report"})
		return
	}

	// Return the success message after updating the report
	c.JSON(http.StatusOK, gin.H{"status": "Report updated successfully"})
}

// GetAllReports retrieves all reports.
func (h *ReportHandler) GetAllReports(c *gin.Context) {
	// Retrieve all reports using the ReportService
	reports, err := h.ReportService.GetAllReports(c.Request.Context()) // Pass context here
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve reports"})
		return
	}

	// Return the reports in the response
	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
	})
}

// GetReportsByCaseID retrieves all reports associated with a specific case.

// GetReportsByCaseID retrieves all reports for a specific case.
func (h *ReportHandler) GetReportsByCaseID(c *gin.Context) {
	caseIDStr := c.Param("caseID")
	caseID, err := uuid.Parse(caseIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	reports, err := h.ReportService.GetReportsByCaseID(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve reports"})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// GetReportsByEvidenceID retrieves all reports for a specific evidence.
func (h *ReportHandler) GetReportsByEvidenceID(c *gin.Context) {
	evidenceIDStr := c.Param("evidenceID")
	evidenceID, err := uuid.Parse(evidenceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid evidence ID"})
		return
	}

	reports, err := h.ReportService.GetReportsByEvidenceID(c.Request.Context(), evidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve reports"})
		return
	}

	c.JSON(http.StatusOK, reports)
}

// DeleteReport handles the request to delete a report.
func (h *ReportHandler) DeleteReport(c *gin.Context) {
	// Extract the report ID from the URL parameters
	reportIDStr := c.Param("reportID")
	reportID, err := uuid.Parse(reportIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	// Extract user information from context

	// Delete the report from the repository
	err = h.ReportService.DeleteReportByID(c.Request.Context(), reportID.String())
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "report not found"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{"message": "Report deleted successfully"})
}
