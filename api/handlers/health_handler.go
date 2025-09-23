package handlers

import (
	"net/http"

	"aegis-api/services_/health"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	Service *health.Service
}

func (h *HealthHandler) Health(c *gin.Context) {
	resp := h.Service.GetHealth()
	statusCode := http.StatusOK
	if resp.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}
	c.JSON(statusCode, resp)
}

func (h *HealthHandler) Readiness(c *gin.Context) {
	if h.Service.GetReadiness() {
		c.String(http.StatusOK, "ready")
	} else {
		c.String(http.StatusServiceUnavailable, "not ready")
	}
}

func (h *HealthHandler) Liveness(c *gin.Context) {
	if h.Service.GetLiveness() {
		c.String(http.StatusOK, "alive")
	} else {
		c.String(http.StatusServiceUnavailable, "dead")
	}
}
