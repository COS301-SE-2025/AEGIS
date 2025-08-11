// routes/coc_routes.go
package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

// Protected group is already wrapped with AuthMiddleware() in SetUpRouter.
// Just mount these under it.
func RegisterCoCRoutes(rg *gin.RouterGroup, h *handlers.CoCHandler) {
	coc := rg.Group("/coc")
	{
		// Log a CoC entry (upload, download, archive, view)
		coc.POST("/log", h.Log)

		// List entries for an evidence item
		coc.GET("/:evidenceId", h.ListByEvidence)

		// Optional: export CSV of CoC entries for an evidence item
		coc.GET("/:evidenceId/export.csv", h.ExportCSV)
	}
}
