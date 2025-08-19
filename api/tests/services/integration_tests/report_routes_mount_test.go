package integration_test

import (
	"aegis-api/handlers"
	routesPkg "aegis-api/routes"
	report "aegis-api/services_/report"

	"github.com/gin-gonic/gin"
)

// This runs before TestMain’s buildRouter() due to package init order.
func init() {
	RegisterRoutes(func(root *gin.RouterGroup) {
		// Use the globals from bootstrap: pgDB, mongoColl
		pgRepo := report.NewReportRepository(pgDB)
		mRepo := report.NewReportMongoRepo(mongoColl)
		svc := report.NewReportService(pgRepo, mRepo)
		h := handlers.NewReportHandler(svc)

		// Reuse your real routes
		routesPkg.RegisterReportRoutes(root, h)
	})
}
