package routes

import (
	"github.com/gin-gonic/gin"
	"aegis-api/handlers"
	"aegis-api/middleware"
)

func RegisterEvidenceRoutes(
	r *gin.RouterGroup,
	viewerHandler *handlers.EvidenceViewerHandler,
	tagHandler *handlers.EvidenceTagHandler,
	permChecker middleware.PermissionChecker,
) {
	// ─── Evidence Viewer ──────────────
	evidence := r.Group("/evidence")
	evidence.Use(middleware.RequirePermission("evidence:view", permChecker))
	evidence.GET("/case/:case_id", viewerHandler.GetEvidenceByCaseID)
	evidence.GET("/:evidence_id", viewerHandler.GetEvidenceByID)
	evidence.GET("/search", viewerHandler.SearchEvidence)
	evidence.POST("/case/:case_id/filter", viewerHandler.GetFilteredEvidence)

	// ─── Evidence Tags ────────────────
	// All tagging requires evidence:tag permission
	tags := r.Group("/evidence-tags")
	tags.Use(middleware.RequirePermission("evidence:tag", permChecker))
	tags.POST("/tag", tagHandler.TagEvidence)
	tags.POST("/untag", tagHandler.UntagEvidence)

	// Viewing tags just needs view permission
	tags.GET("/:evidence_id",
		middleware.RequirePermission("evidence:view", permChecker),
		tagHandler.GetEvidenceTags)
}
