package handlers

import (
	"aegis-api/structs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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
