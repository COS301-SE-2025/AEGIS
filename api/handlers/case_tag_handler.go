package handlers

import (
	"aegis-api/services_/case/case_tags"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type tagRequest struct {
	Tags []string `json:"tags" binding:"required"`
}

type CaseTagHandler struct {
	Service case_tags.CaseTagService
}

// POST /cases/:case_id/tags
func (h *CaseTagHandler) TagCase(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	caseID, err := uuid.Parse(c.Param("case_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID missing in context"})
		return
	}

	if err := h.Service.TagCase(c.Request.Context(), uuid.MustParse(userID.(string)), caseID, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to tag case", "details": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// DELETE /cases/:case_id/tags
func (h *CaseTagHandler) UntagCase(c *gin.Context) {
	var req tagRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "details": err.Error()})
		return
	}

	caseID, err := uuid.Parse(c.Param("case_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	userID, ok := c.Get("userID")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "userID missing in context"})
		return
	}

	if err := h.Service.UntagCase(c.Request.Context(), uuid.MustParse(userID.(string)), caseID, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to untag case", "details": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /cases/:case_id/tags
func (h *CaseTagHandler) GetTags(c *gin.Context) {
	caseID, err := uuid.Parse(c.Param("case_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid case ID"})
		return
	}

	tags, err := h.Service.GetTags(c.Request.Context(), caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get tags", "details": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
