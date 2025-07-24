package handlers

import (
	"fmt"
	"net/http"

	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
)

// GET /cases/all
func (h *CaseHandler) GetAllCasesHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this set

	}

	cases, err := h.ListCasesService.GetAllCases()
	if err != nil {
		fmt.Printf("[GetAllCasesHandler] failed: %v\n", err)

		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_ALL_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_listing",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to list all cases: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve cases"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_ALL_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case_listing",
			ID:   "",
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d cases", len(cases)),
	})

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GET /cases/user/:user_id
func (h *CaseHandler) GetCasesByUserHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this set

	}
	paramUserID := c.Param("user_id")
	if paramUserID == "" {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_USER_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_listing_by_user",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Missing user_id parameter",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	cases, err := h.ListCasesService.GetCasesByUser(paramUserID)
	if err != nil {
		fmt.Printf("[GetCasesByUserHandler] failed: %v\n", err)
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_USER_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_listing_by_user",
				ID:   paramUserID,
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to get cases for user: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve cases for user"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_USER_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case_listing_by_user",
			ID:   paramUserID,
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d cases for user %s", len(cases), paramUserID),
	})

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GET /cases/filter
func (h *CaseHandler) GetFilteredCasesHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this set

	}

	status := c.Query("status")
	priority := c.Query("priority")
	createdBy := c.Query("created_by")
	teamName := c.Query("team_name")
	titleTerm := c.Query("title_term")
	sortBy := c.Query("sort_by")
	order := c.Query("order")

	cases, err := h.ListCasesService.GetFilteredCases(
		status, priority, createdBy, teamName, titleTerm, sortBy, order,
	)
	if err != nil {
		fmt.Printf("[GetFilteredCasesHandler] failed: %v\n", err)
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LIST_FILTERED_CASES",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_filtered_listing",
				ID:   "",
				AdditionalInfo: map[string]string{
					"status":    status,
					"priority":  priority,
					"createdBy": createdBy,
					"teamName":  teamName,
					"titleTerm": titleTerm,
					"sortBy":    sortBy,
					"order":     order,
				},
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to apply case filters: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not apply filters"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LIST_FILTERED_CASES",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case_filtered_listing",
			ID:   "",
			AdditionalInfo: map[string]string{
				"status":    status,
				"priority":  priority,
				"createdBy": createdBy,
				"teamName":  teamName,
				"titleTerm": titleTerm,
				"sortBy":    sortBy,
				"order":     order,
			},
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d filtered cases", len(cases)),
	})

	c.JSON(http.StatusOK, gin.H{"cases": cases})
}

// GET /cases/:case_id
func (h *CaseHandler) GetCaseByIDHandler(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")
	email, _ := c.Get("email") // Optional, if you have this set
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     email.(string), // Optional, if you have this set

	}

	caseID := c.Param("case_id")
	if caseID == "" {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CASE_BY_ID",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_details",
				ID:   "",
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Missing case_id parameter",
		})
		c.JSON(http.StatusBadRequest, gin.H{"error": "case_id is required"})
		return
	}

	caseDetails, err := h.ListCasesService.GetCaseByID(caseID)
	if err != nil {
		fmt.Printf("[GetCaseByIDHandler] failed: %v\n", err)
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "GET_CASE_BY_ID",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "case_details",
				ID:   caseID,
			},
			Service:     "case",
			Status:      "FAILED",
			Description: "Failed to retrieve case by ID: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not retrieve case"})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "GET_CASE_BY_ID",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "case_details",
			ID:   caseID,
		},
		Service:     "case",
		Status:      "SUCCESS",
		Description: "Retrieved case details successfully",
	})

	c.JSON(http.StatusOK, gin.H{"case": caseDetails})
}
