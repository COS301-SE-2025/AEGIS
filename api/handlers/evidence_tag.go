package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"aegis-api/services_/evidence/evidence_tag"
)

// Request payload struct for POST /evidence-tags/tag
type TagEvidenceRequest struct {
	EvidenceID string   `json:"evidence_id"`
	Tags       []string `json:"tags"`
}

type EvidenceTagHandler struct {
	Service evidence_tag.EvidenceTagService
}

func (h *EvidenceTagHandler) TagEvidence(c *gin.Context) {

	var req TagEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	userIDStr := c.GetString("userID")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID in JWT claims"})
		return
	}

	evidenceID, err := uuid.Parse(req.EvidenceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidenceID"})
		return
	}

	if err := h.Service.TagEvidence(context.Background(), userID, evidenceID, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to tag evidence"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Evidence tagged successfully"})
}

func (h *EvidenceTagHandler) UntagEvidence(c *gin.Context) {
	var req TagEvidenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
		return
	}

	userIDStr := c.GetString("userID") 
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid userID in JWT claims"})
		return
	}

	evidenceID, err := uuid.Parse(req.EvidenceID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidenceID"})
		return
	}

	if err := h.Service.UntagEvidence(context.Background(), userID, evidenceID, req.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to untag evidence"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Evidence untagged successfully"})
}


func (h *EvidenceTagHandler) GetEvidenceTags(c *gin.Context) {
	evidenceIDStr := c.Param("evidence_id")

	evidenceID, err := uuid.Parse(evidenceIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid evidenceID"})
		return
	}

	tags, err := h.Service.GetEvidenceTags(context.Background(), evidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tags"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}
