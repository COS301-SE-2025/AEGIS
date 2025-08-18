package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"aegis-api/services_/report/update_status"
)

type UpdateStatusRequest struct {
	Status update_status.ReportStatus `json:"status" binding:"required"`
}

type ReportStatusHandler struct {
	service update_status.ReportStatusService
}

// Accepts the service instead of the repo
func NewReportStatusHandler(service update_status.ReportStatusService) *ReportStatusHandler {
	return &ReportStatusHandler{service: service}
}

func (h *ReportStatusHandler) RegisterRoutes(r *gin.Engine) {
	group := r.Group("/api/v1/reports")
	{
		group.PUT("/:id/status", h.UpdateStatus)
	}
}

func (h *ReportStatusHandler) UpdateStatus(c *gin.Context) {
	idParam := c.Param("id")
	reportID, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid report ID"})
		return
	}

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.service.UpdateStatus(c.Request.Context(), reportID, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}
