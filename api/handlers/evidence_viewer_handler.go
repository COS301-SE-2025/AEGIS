
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"aegis-api/services_/evidence/evidence_viewer"
)

type EvidenceViewerHandler struct {
	Service *evidence_viewer.EvidenceService
}



func NewEvidenceViewerHandler(svc *evidence_viewer.EvidenceService) *EvidenceViewerHandler {
	return &EvidenceViewerHandler{Service: svc}
}

// GET /evidence/case/:case_id
func (h *EvidenceViewerHandler) GetEvidenceByCaseID(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing case ID"})
		return
	}

	files, err := h.Service.GetEvidenceFilesByCaseID(caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evidence files by case"})
		return
	}

	if len(files) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No evidence files found"})
		return
	}

	
	c.JSON(http.StatusOK, gin.H{"files": files})
}

// GET /evidence/:evidence_id
func (h *EvidenceViewerHandler) GetEvidenceByID(c *gin.Context) {
	evidenceID := c.Param("evidence_id")
	if evidenceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing evidence ID"})
		return
	}

	file, err := h.Service.GetEvidenceFileByID(evidenceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get evidence file by ID"})
		return
	}

	if file == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Evidence file not found"})
		return
	}

	
	c.Data(http.StatusOK, "application/octet-stream", file.Data)
}

// GET /evidence/search?query=
func (h *EvidenceViewerHandler) SearchEvidence(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing search query"})
		return
	}

	files, err := h.Service.SearchEvidenceFiles(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search evidence files"})
		return
	}

	if len(files) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching evidence files found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}

// POST /evidence/case/:case_id/filter
type FilterRequest struct {
	Filters   map[string]interface{} `json:"filters"`
	SortField string                 `json:"sort_field"`
	SortOrder string                 `json:"sort_order"`
}

func (h *EvidenceViewerHandler) GetFilteredEvidence(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing case ID"})
		return
	}

	var req FilterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	files, err := h.Service.GetFilteredEvidenceFiles(caseID, req.Filters, req.SortField, req.SortOrder)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to filter evidence files"})
		return
	}

	if len(files) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No matching evidence files found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": files})
}
