package routes

import (
	"aegis-api/handlers"

	"github.com/gin-gonic/gin"
)

// Health routes don’t need permissions (they’re usually public for monitoring).
func RegisterHealthRoutes(rg *gin.RouterGroup, h *handlers.HealthHandler) {
	group := rg.Group("/health")
	{
		group.GET("", h.Health)       // full system health
		group.GET("/ready", h.Readiness) // readiness probe
		group.GET("/live", h.Liveness)   // liveness probe
	}
}
