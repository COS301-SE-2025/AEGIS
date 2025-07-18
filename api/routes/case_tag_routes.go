package routes

import (
	"aegis-api/handlers"
	"aegis-api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterCaseTagRoutes(rg *gin.RouterGroup, h *handlers.CaseTagHandler, checker middleware.PermissionChecker) {
	group := rg.Group("/cases/:case_id/tags")
	{
		group.GET("", h.GetTags)

		group.POST("",
			middleware.RequirePermission("case:tag", checker),
			h.TagCase,
		)

		group.DELETE("",
			middleware.RequirePermission("case:remove_tag", checker),
			h.UntagCase,
		)
	}
}

