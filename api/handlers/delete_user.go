package handlers

import (
	"fmt"
	"net/http"

	"aegis-api/services_/admin/delete_user"
	"aegis-api/services_/auditlog"
	"aegis-api/structs"

	"github.com/gin-gonic/gin"
)

// DeleteUserHandler handles DELETE /api/v1/users/:userId
func (s *AdminService) DeleteUserHandler(c *gin.Context) {
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
		fmt.Printf("[DeleteUserHandler] Missing userID in context\n")

		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_USER",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "user",
				ID:   "",
			},
			Service:     "user",
			Status:      "FAILED",
			Description: "Unauthorized: missing userID in context",
		})

		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: missing userID in context"})
		return
	}

	userIDToDelete := c.Param("userId")
	if userIDToDelete == "" {
		fmt.Printf("[DeleteUserHandler] Missing userId in path\n")

		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_USER",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "user",
				ID:   "",
			},
			Service:     "user",
			Status:      "FAILED",
			Description: "Missing userId in path",
		})

		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "missing_user_id",
			Message: "Missing userId in path",
		})
		return
	}

	// Extract role from context (middleware sets as 'userRole')
	role, exists := c.Get("userRole")
	if !exists {
		fmt.Printf("[DeleteUserHandler] Role not found in context\n")

		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_USER",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "user",
				ID:   userIDToDelete,
			},
			Service:     "user",
			Status:      "FAILED",
			Description: "Role not found in context",
		})

		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Role not found in context",
		})
		return
	}

	req := delete_user.DeleteUserRequest{UserID: userIDToDelete}
	err := s.userDeleteService.DeleteUser(req, role.(string))
	if err != nil {
		fmt.Printf("[DeleteUserHandler] Failed to delete user: %v\n", err)

		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "DELETE_USER",
			Actor:  actor,
			Target: auditlog.Target{
				Type: "user",
				ID:   userIDToDelete,
			},
			Service:     "user",
			Status:      "FAILED",
			Description: "Delete user failed: " + err.Error(),
		})

		c.JSON(http.StatusForbidden, structs.ErrorResponse{
			Error:   "delete_user_failed",
			Message: err.Error(),
		})
		return
	}
	fmt.Printf("[DeleteUserHandler] Successfully deleted user: %s\n", userIDToDelete)

	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "DELETE_USER",
		Actor:  actor,
		Target: auditlog.Target{
			Type: "user",
			ID:   userIDToDelete,
		},
		Service:     "user",
		Status:      "SUCCESS",
		Description: "User deleted successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
