package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// EnhanceSummary godoc
// @Summary Enhance a summary using AI
// @Description Rewrite a summary in simple English using AI
// @Tags reports, ai
// @Accept json
// @Produce json
// @Param body body map[string]string true "Text to enhance"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /reports/ai/enhance-summary [post]
func (h *ReportAIHandler) EnhanceSummary(c *gin.Context) {
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	// Call AI client to enhance summary with context
	result, err := h.Service.EnhanceSummary(c.Request.Context(), payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"enhanced": result})
}
