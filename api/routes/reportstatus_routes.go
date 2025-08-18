package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

// RegisterReportStatusRoutes registers routes for updating report status.
func RegisterReportStatusRoutes(router *gin.RouterGroup, handler *handlers.ReportStatusHandler) {
	reportStatus := router.Group("/reports")
	{
		
		reportStatus.PUT("/:id/status", handler.UpdateStatus)
	}
}
