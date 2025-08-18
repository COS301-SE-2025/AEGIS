package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)


func RegisterReportStatusRoutes(router *gin.RouterGroup, handler *handlers.ReportStatusHandler) {
	reportStatus := router.Group("/reports")
	{
		
		reportStatus.PUT("/:id/status", handler.UpdateStatus)
	}
}
