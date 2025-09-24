package integration_test

import (
	"aegis-api/handlers"
	routesPkg "aegis-api/routes"
	report "aegis-api/services_/report"
	reportai "aegis-api/services_/report/report_ai_assistance"

	"github.com/gin-gonic/gin"
)

// This runs before TestMain’s buildRouter() due to package init order.
func init() {
	RegisterRoutes(func(root *gin.RouterGroup) {
		// Use the globals from bootstrap: pgDB, mongoColl
		pgRepo := report.NewReportRepository(pgDB)
		mRepo := report.NewReportMongoRepo(mongoColl)
		sectionRepo := reportai.NewGormReportSectionRepo(pgDB) // Use the correct constructor for sectionRepo
		svc := report.NewReportService(pgRepo, mRepo, sectionRepo)
		h := handlers.NewReportHandler(svc)

		// Reuse your real routes
		routesPkg.RegisterReportRoutes(root, h)
	})
}
