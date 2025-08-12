package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterReportRoutes registers routes for managing reports.
func RegisterReportRoutes(router *gin.RouterGroup, handler *handlers.ReportHandler) {
	report := router.Group("/reports")
	{
		report.POST("/cases/:caseID", handler.GenerateReport)               // Generate report for case
		report.GET("/cases/:caseID", handler.GetReportsByCaseID)            // Get all reports for a case
		report.GET("/evidence/:evidenceID", handler.GetReportsByEvidenceID) // Get all reports for evidence
		report.GET("/:reportID", handler.GetReportByID)                     // Get a specific report
		report.PUT("/:reportID", handler.UpdateReport)                      // Update a report
	}
}
