package handlers

import (
	"net/http"
	"aegis-api/services_/case/case_evidence_totals"

	"github.com/gin-gonic/gin"
	"strings"
)

type CaseEvidenceTotalsHandler struct {
	DashboardService case_evidence_totals.DashboardService
}

func NewCaseEvidenceTotalsHandler(service case_evidence_totals.DashboardService) *CaseEvidenceTotalsHandler {
	return &CaseEvidenceTotalsHandler{
		DashboardService: service,
	}
}



func (h *CaseEvidenceTotalsHandler) GetDashboardTotals(c *gin.Context) {
	statusQuery := c.DefaultQuery("statuses", "open,ongoing,closed")
	statuses := strings.Split(statusQuery, ",")

	userID := c.GetString("userID") // from middleware
	if userID == "" {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	caseCount, evidenceCount, err := h.DashboardService.GetCounts(userID, statuses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve dashboard totals"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"case_count":     caseCount,
		"evidence_count": evidenceCount,
	})
}
