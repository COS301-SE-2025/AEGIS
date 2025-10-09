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
		//report.POST("/:reportID/download/pdf", handler.DownloadReportPDF) 

		// Section-level endpoints
		report.POST("/:reportID/sections", handler.AddSection)                             // Add custom section
		report.PUT("/:reportID/sections/:sectionID/content", handler.UpdateSectionContent) // Update section content
		report.PUT("/:reportID/sections/:sectionID/title", handler.UpdateSectionTitle)     // Update section title
		report.PUT("/:reportID/sections/:sectionID/reorder", handler.ReorderSection)       // Reorder section
		report.DELETE("/:reportID/sections/:sectionID", handler.DeleteSection)

		// Context autofill endpoint
		report.GET("/:reportID/sections/:sectionID/context", handler.GetSectionContext)

		// Recent reports endpoint
		report.GET("/recent", handler.GetRecentReports) // List recent reports

		// Report name update endpoint
		report.PUT("/:reportID/name", handler.UpdateReportName) // Update report name
		// 🔹 Team-scoped list — add this near the top
		report.GET("/teams/:teamID", handler.GetReportsForTeam)

	}
}

// RegisterReportAIRoutes registers routes for AI assistance on reports
func RegisterReportAIRoutes(router *gin.RouterGroup, handler *handlers.ReportAIHandler) {
	reportAI := router.Group("/reports/ai")
	{
		// Generate AI suggestion for a section (GET and POST)
		reportAI.GET("/:reportID/sections/:sectionID/suggest", handler.SuggestSection)
		reportAI.POST("/:reportID/sections/:sectionID/suggest", handler.SuggestSectionPOST)

		// Submit feedback on AI suggestion
		reportAI.POST("/sections/:sectionID/feedback", handler.SubmitFeedback)

		// Optionally, generate AI references for a section
		reportAI.GET("/:reportID/sections/:sectionID/references", handler.GenerateReferences)

		// Enhance summary endpoint
		reportAI.POST("/enhance-summary", handler.EnhanceSummary)
	}
}
