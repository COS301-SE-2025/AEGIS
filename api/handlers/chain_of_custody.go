package handlers

import (
	"aegis-api/services_/auditlog"
	"aegis-api/services_/chain_of_custody"
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ChainOfCustodyHandler struct {
	service     chain_of_custody.ChainOfCustodyService
	auditLogger *auditlog.AuditLogger
}

func NewChainOfCustodyHandler(service chain_of_custody.ChainOfCustodyService, auditLogger *auditlog.AuditLogger) *ChainOfCustodyHandler {
	return &ChainOfCustodyHandler{
		service:     service,
		auditLogger: auditLogger,
	}
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

	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[AddEntry] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   "",
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	var custody chain_of_custody.ChainOfCustody
	if err := c.ShouldBindJSON(&custody); err != nil {
		fmt.Printf("[AddEntry] Bind error: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   "",
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Invalid JSON input for adding chain of custody entry: " + err.Error(),
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("[AddEntry] Parsed chain of custody struct: %+v\n", custody)

	err := h.service.AddEntry(context.Background(), &custody)
	if err != nil {
		fmt.Printf("[AddEntry] Failed to add chain of custody entry: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ADD_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   custody.ID.String(),
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Failed to add chain of custody entry: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("[AddEntry] Successfully added chain of custody entry: %+v\n", custody)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "ADD_CHAIN_OF_CUSTODY_ENTRY",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "chain_of_custody",
			ID:   custody.ID.String(),
		},
		Service:     "chain_of_custody",
		Status:      "SUCCESS",
		Description: "Chain of custody entry added successfully",
	})

	c.JSON(http.StatusCreated, custody)
}

// PUT /api/v1/chain_of_custody/:id
func (h *ChainOfCustodyHandler) UpdateEntry(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[UpdateEntry] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   "",
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		fmt.Printf("[UpdateEntry] Invalid ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   idStr,
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Invalid ID format for updating chain of custody entry: " + err.Error(),
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var custody chain_of_custody.ChainOfCustody
	if err := c.ShouldBindJSON(&custody); err != nil {
		fmt.Printf("[UpdateEntry] Bind error: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   id.String(),
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Invalid JSON input for updating chain of custody entry: " + err.Error(),
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	custody.ID = id

	err = h.service.UpdateEntry(context.Background(), &custody)
	if err != nil {
		fmt.Printf("[UpdateEntry] Failed to update chain of custody entry: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "UPDATE_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   id.String(),
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Failed to update chain of custody entry: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("[UpdateEntry] Successfully updated chain of custody entry: %+v\n", custody)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "UPDATE_CHAIN_OF_CUSTODY_ENTRY",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "chain_of_custody",
			ID:   id.String(),
		},
		Service:     "chain_of_custody",
		Status:      "SUCCESS",
		Description: "Chain of custody entry updated successfully",
	})

	c.JSON(http.StatusOK, custody)
}

// GET /api/v1/chain_of_custody/evidence/:evidenceID
func (h *ChainOfCustodyHandler) GetEntries(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GetEntries] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CHAIN_OF_CUSTODY_ENTRIES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody_entries",
				ID:   "",
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	evidenceIDStr := c.Query("evidence_id")
	fmt.Printf("[DEBUG] GetEntries: evidenceIDStr=%s\n", evidenceIDStr)
	evidenceID, err := uuid.Parse(evidenceIDStr)
	if err != nil {
		fmt.Printf("[DEBUG] GetEntries: invalid evidenceID %s\n", evidenceIDStr)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CHAIN_OF_CUSTODY_ENTRIES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence",
				ID:   evidenceIDStr,
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Invalid evidenceID format for getting chain of custody entries: " + err.Error(),
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid evidenceID"})
		return
	}

	entries, err := h.service.GetEntries(context.Background(), evidenceID)
	fmt.Printf("[DEBUG] GetEntries: entries=%+v\n", entries)
	if err != nil {
		fmt.Printf("[DEBUG] GetEntries: error=%v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CHAIN_OF_CUSTODY_ENTRIES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "evidence",
				ID:   evidenceID.String(),
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Failed to get chain of custody entries: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("[GetEntries] Successfully retrieved chain of custody entries: %+v\n", entries)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_CHAIN_OF_CUSTODY_ENTRIES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "evidence",
			ID:   evidenceID.String(),
		},
		Service:     "chain_of_custody",
		Status:      "SUCCESS",
		Description: "Chain of custody entries retrieved successfully",
	})

	c.JSON(http.StatusOK, entries)
}

// GET /api/v1/chain_of_custody/:id
func (h *ChainOfCustodyHandler) GetEntry(c *gin.Context) {
	// Grab user details from context
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set

	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this header set
	}

	_, exists := c.Get("userID")
	if !exists {
		fmt.Printf("[GetEntry] Missing userID in context\n")

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   "",
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		fmt.Printf("[GetEntry] Invalid ID: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   idStr,
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Invalid ID format for getting chain of custody entry: " + err.Error(),
		})

		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	entry, err := h.service.GetEntry(context.Background(), id)
	if err != nil {
		fmt.Printf("[GetEntry] Failed to get chain of custody entry: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CHAIN_OF_CUSTODY_ENTRY",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "chain_of_custody",
				ID:   id.String(),
			},
			Service:     "chain_of_custody",
			Status:      "FAILED",
			Description: "Failed to get chain of custody entry: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("[GetEntry] Successfully retrieved chain of custody entry: %+v\n", entry)

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_CHAIN_OF_CUSTODY_ENTRY",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "chain_of_custody",
			ID:   id.String(),
		},
		Service:     "chain_of_custody",
		Status:      "SUCCESS",
		Description: "Chain of custody entry retrieved successfully",
	})

	c.JSON(http.StatusOK, entry)
}
