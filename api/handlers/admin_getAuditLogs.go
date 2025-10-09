package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"aegis-api/services_/auditlog"
	"aegis-api/structs"

	"github.com/gin-gonic/gin"
)

// Assumes you have imported necessary packages: "aegis-api/services_/auditlog", "strconv", "github.com/google/uuid" (already there)
func (s *AdminService) GetAuditLogs(c *gin.Context) {
	// Auth check: Ensure admin role (from context)
	userRole, exists := c.Get("userRole")
	if !exists || (userRole.(string) != "Admin" && userRole.(string) != "DFIR Admin") { // Adjust roles as needed
		// Log the unauthorized access attempt
		userID, _ := c.Get("userID")
		actor := auditlog.Actor{
			ID:        userID.(string),
			Role:      userRole.(string),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Email:     "", // Fetch if available
		}
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_AUDIT_LOGS",
			Actor:       actor,
			Target:      auditlog.Target{Type: "audit_logs", ID: ""},
			Service:     "admin",
			Status:      "FAILED",
			Description: "Unauthorized access to audit logs",
		})

		c.JSON(http.StatusForbidden, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Admin access required",
		})
		return
	}

	// Parse filters from query params
	filter := auditlog.AuditLogFilter{
		Status:  c.DefaultQuery("status", "ALL"),
		Action:  c.Query("action"),  // e.g., "EXTRACT_IOCS" for IOC retrievals
		Service: c.Query("service"), // e.g., "timelineai"
		Limit:   atoiOrDefault(c.Query("limit"), 100),
	}

	logs, err := s.auditLogService.GetAuditLogs(c.Request.Context(), filter)
	if err != nil {
		log.Printf("[GetAuditLogs] Failed to retrieve logs: %v\n", err)

		// Log the failure
		userID, _ := c.Get("userID")
		actor := auditlog.Actor{
			ID:        userID.(string),
			Role:      userRole.(string),
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Email:     "", // Fetch if available
		}
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action:      "GET_AUDIT_LOGS",
			Actor:       actor,
			Target:      auditlog.Target{Type: "audit_logs", ID: ""},
			Service:     "admin",
			Status:      "FAILED",
			Description: "Failed to retrieve audit logs: " + err.Error(),
		})

		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to retrieve logs: " + err.Error(),
		})
		return
	}

	// Log success
	userID, _ := c.Get("userID")
	actor := auditlog.Actor{
		ID:        userID.(string),
		Role:      userRole.(string),
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Email:     "", // Fetch if available
	}
	s.auditLogger.Log(c, auditlog.AuditLog{
		Action:      "GET_AUDIT_LOGS",
		Actor:       actor,
		Target:      auditlog.Target{Type: "audit_logs", ID: ""},
		Service:     "admin",
		Status:      "SUCCESS",
		Description: fmt.Sprintf("Retrieved %d audit logs with filter %+v", len(logs), filter),
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Data: gin.H{
			"logs":   logs,
			"filter": filter, // Optional: echo back filter
		},
		Message: "Audit logs retrieved successfully",
	})
}

// Helper function (add to this file or utils)
func atoiOrDefault(s string, def int) int {
	if i, err := strconv.Atoi(s); err == nil && i > 0 {
		return i
	}
	return def
}
