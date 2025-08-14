package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterReportRoutes registers routes for managing reports and their sections.
func RegisterReportRoutes(router *gin.RouterGroup, handler *handlers.ReportHandler) {
	report := router.Group("/reports")
	{
		// Report-level endpoints
		report.POST("/cases/:caseID", handler.GenerateReport)    // Generate report for case
		report.GET("/cases/:caseID", handler.GetReportsByCaseID) // Get all reports for a case
		//report.GET("/evidence/:evidenceID", handler.GetReportsByEvidenceID)  // Get all reports for evidence
		report.GET("/:reportID", handler.GetReportByID) // Get a specific report
		//report.PUT("/:reportID", handler.UpdateReport)                       // Update a report
		report.DELETE("/:reportID", handler.DeleteReport) // Delete a report

		// Download endpoints
		report.GET("/:reportID/download/pdf", handler.DownloadReportPDF)   // Download PDF
		report.GET("/:reportID/download/json", handler.DownloadReportJSON) // Download JSON

		// Section-level endpoints
		report.POST("/:reportID/sections", handler.AddSection)                             // Add custom section
		report.PUT("/:reportID/sections/:sectionID/content", handler.UpdateSectionContent) // Update section content
		report.PUT("/:reportID/sections/:sectionID/title", handler.UpdateSectionTitle)     // Update section title
		report.PUT("/:reportID/sections/:sectionID/reorder", handler.ReorderSection)       // Reorder section
		report.DELETE("/:reportID/sections/:sectionID", handler.DeleteSection)             // Delete section
	}
}
