package handlers

import (
	"aegis-api/structs"
	"log"
	"net/http"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GET /api/v1/tenants/:tenant_id/users

func (s *AdminService) ListUsersByTenant(c *gin.Context) {
	tenantIDStr := c.Param("tenantId")
	if tenantIDStr == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "missing_tenant_id",
			Message: "Missing tenant_id in path",
		})
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_tenant_id",
			Message: "Invalid tenant_id format",
		})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	users, total, err := s.listUserService.ListUsersByTenant(c.Request.Context(), tenantID, page, pageSize)
	if err != nil {
		log.Printf("ListUsersByTenant error: %v", err)
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "list_users_by_tenant_failed",
			Message: "Failed to list users for tenant",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Users for tenant retrieved successfully",
		"data":    users,
		"total":   total,
	})
}

// GET /api/v1/users
func (s *AdminService) ListUsers(c *gin.Context) {
	users, err := s.listUserService.ListUsers(c.Request.Context())
	if err != nil {
		log.Printf("ListUsers error: %v", err)
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "list_users_failed",
			Message: "Failed to list users",
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}
