package handlers

import (
	"aegis-api/services_/chain_of_custody"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChainOfCustodyHandler struct {
	service chain_of_custody.ChainOfCustodyService
}

func NewChainOfCustodyHandler(service chain_of_custody.ChainOfCustodyService) *ChainOfCustodyHandler {
	return &ChainOfCustodyHandler{service: service}
}

// POST /api/v1/chain_of_custody
func (h *ChainOfCustodyHandler) AddEntry(c *gin.Context) {
	if h == nil {
		fmt.Println("ERROR: Handler is nil")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Handler not initialized"})
		return
	}
	if h.service == nil {
		fmt.Println("ERROR: Service is nil in ChainOfCustodyHandler")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Service not initialized"})
		return
	}

	var custody chain_of_custody.ChainOfCustody
	if err := c.ShouldBindJSON(&custody); err != nil {
		fmt.Printf("Bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Parsed chain of custody struct: %+v\n", custody)

	err := h.service.AddEntry(context.Background(), &custody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, custody)
}

// PUT /api/v1/chain_of_custody/:id
func (h *ChainOfCustodyHandler) UpdateEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var custody chain_of_custody.ChainOfCustody
	if err := c.ShouldBindJSON(&custody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	custody.ID = id

	err = h.service.UpdateEntry(context.Background(), &custody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, custody)
}

// GET /api/v1/chain_of_custody/evidence/:evidenceID
func (h *ChainOfCustodyHandler) GetEntries(c *gin.Context) {
	evidenceIDStr := c.Query("evidence_id")
	fmt.Printf("[DEBUG] GetEntries: evidenceIDStr=%s\n", evidenceIDStr)
	evidenceID, err := uuid.Parse(evidenceIDStr)
	if err != nil {
		fmt.Printf("[DEBUG] GetEntries: invalid evidenceID %s\n", evidenceIDStr)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid evidenceID"})
		return
	}
	entries, err := h.service.GetEntries(context.Background(), evidenceID)
	fmt.Printf("[DEBUG] GetEntries: entries=%+v\n", entries)
	if err != nil {
		fmt.Printf("[DEBUG] GetEntries: error=%v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entries)
}

// GET /api/v1/chain_of_custody/:id
func (h *ChainOfCustodyHandler) GetEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	entry, err := h.service.GetEntry(context.Background(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, entry)
}
