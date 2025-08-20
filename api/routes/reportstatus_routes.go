package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)


func RegisterReportStatusRoutes(router *gin.RouterGroup, handler *handlers.ReportStatusHandler) {
	reportStatus := router.Group("/reports")
	{

		reportStatus.PUT("/:reportID/status", handler.UpdateStatus)
	}
}
