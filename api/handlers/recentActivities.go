package handlers

import (
	"aegis-api/services_/auditlog"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RecentActivityHandler struct {
	auditLogService auditlog.AuditLogReader
}

func NewRecentActivityHandler(service auditlog.AuditLogReader) *RecentActivityHandler { // ✅ use interface
	return &RecentActivityHandler{
		auditLogService: service,
	}
}
func (h *RecentActivityHandler) GetRecentActivities(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	logs, err := h.auditLogService.GetRecentUserActivities(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Recent activities retrieved",
		"data":    logs,
	})
}
