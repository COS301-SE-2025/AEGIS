package handlers

import (
	"net/http"

	"aegis-api/services_/admin/delete_user"
	"aegis-api/structs"

	"github.com/gin-gonic/gin"
)

// DeleteUserHandler handles DELETE /api/v1/users/:userId
func (s *AdminService) DeleteUserHandler(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "missing_user_id",
			Message: "Missing userId in path",
		})
		return
	}

	// Extract role from context (middleware sets as 'userRole')
	role, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Role not found in context",
		})
		return
	}

	req := delete_user.DeleteUserRequest{UserID: userID}
	err := s.userDeleteService.DeleteUser(req, role.(string))
	if err != nil {
		c.JSON(http.StatusForbidden, structs.ErrorResponse{
			Error:   "delete_user_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
