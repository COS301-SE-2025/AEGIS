package handlers

import (
	"net/http"

	graphicalmapping "aegis-api/services_/GraphicalMapping"

	"github.com/gin-gonic/gin"
)

type IOCHandler struct {
	service graphicalmapping.IOCService
}

func NewIOCHandler(service graphicalmapping.IOCService) *IOCHandler {
	return &IOCHandler{service: service}
}

// Handler for whole network of cases (tenant-wide)
func (h *IOCHandler) GetTenantIOCGraph(c *gin.Context) {
	tenantIDFromToken, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenantID in token context"})
		return
	}
	tenantIDStr := tenantIDFromToken.(string)

	// Check URL param matches token tenant ID
	tenantIDParam := c.Param("tenantId")
	if tenantIDParam != tenantIDStr {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID mismatch"})
		return
	}

	nodes, edges, err := h.service.BuildIOCGraph(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	elements := graphicalmapping.ConvertToCytoscapeElements(nodes, edges)
	c.JSON(http.StatusOK, elements)
}

func (h *IOCHandler) GetCaseIOCGraph(c *gin.Context) {
	tenantIDFromToken, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenantID in token context"})
		return
	}
	tenantIDStr := tenantIDFromToken.(string)

	tenantIDParam := c.Param("tenantId")
	if tenantIDParam != tenantIDStr {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID mismatch"})
		return
	}

	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing case ID"})
		return
	}

	nodes, edges, err := h.service.BuildIOCGraphByCase(tenantIDStr, caseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	elements := graphicalmapping.ConvertToCytoscapeElements(nodes, edges)
	c.JSON(http.StatusOK, elements)
}

// GET /cases/:caseId/iocs
func (h *IOCHandler) GetIOCsByCase(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing case ID"})
		return
	}

	iocs, err := h.service.ListIOCsByCase(caseID) // <-- call correct method name
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, iocs)
}

// POST /cases/:caseId/iocs
func (h *IOCHandler) AddIOCToCase(c *gin.Context) {
	caseID := c.Param("case_id")
	if caseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing case ID"})
		return
	}

	var req struct {
		Type  string `json:"type" binding:"required"`
		Value string `json:"value" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tenantIDFromToken, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing tenantID in token"})
		return
	}

	ioc := &graphicalmapping.IOC{
		CaseID:   caseID,
		TenantID: tenantIDFromToken.(string),
		Type:     req.Type,
		Value:    req.Value,
	}

	createdIOC, err := h.service.AddIOC(ioc) // <-- pass IOC struct, get created IOC + error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdIOC)
}
