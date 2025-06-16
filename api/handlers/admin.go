package handlers

import (
	"aegis-api/services/delete_user"
	"aegis-api/services/listusers"
	"aegis-api/services/registration"
	"aegis-api/services/update_user_role"
	"aegis-api/structs"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AdminServices struct {
	registerUser   *registration.RegistrationService
	listUser       listusers.ListUserService
	updateUserRole *update_user_role.UserService

	deleteUser *delete_user.UserDeleteService
}

func NewAdminServices(
	registerUser *registration.RegistrationService,
	listUser listusers.ListUserService,
	updateUserRole *update_user_role.UserService,
	deleteUser *delete_user.UserDeleteService,
) *AdminServices {
	return &AdminServices{
		registerUser:   registerUser,
		listUser:       listUser,
		updateUserRole: updateUserRole,
		deleteUser:     deleteUser,
	}
}

// @Summary Register a new user
// @Description Registers a new user with the provided details. Only users with 'Admin' role can perform this action.
// @Tags Admin
// @Accept  json
// @Produce  json
// @Param   request body structs.RegisterUserRequest true "User Registration Request"
// @Success 201 {object} structs.SuccessResponse{data=structs.User} "User registered successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/admin/users [post]
func (m AdminServices) RegisterUser(c *gin.Context) {
	//struct to hold user data
	//binding and validation
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid input",
			Details: err.Error(),
		})
		return
	}

	//call the service function
	user, err := m.registerUser.Register(req)

	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: "Could not register user",
			Details: err.Error(),
		})
		return
	}

	//http response
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}

// @Summary List all users
// @Description Retrieves a list of all registered users. Supports filtering by role, status, and creation date range.
// @Tags Admin
// @Accept json
// @Produce json
// @Param role query string false "Filter users by role (e.g., 'Forensic Analyst')"
// @Param status query string false "Filter users by status (e.g., 'active', 'inactive')"
// @Param start_date query string false "Filter users created after this date (YYYY-MM-DD)"
// @Param end_date query string false "Filter users created before this date (YYYY-MM-DD)"
// @Success 200 {object} structs.SuccessResponse{data=[]structs.User} "Users retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/admin/users [get]
func (m AdminServices) ListUsers(c *gin.Context) {
	//binding and validation
	//var req structs.UserFilter
	//if err := c.ShouldBindQuery(&req); err != nil {
	//	c.JSON(http.StatusBadRequest, structs.ErrorResponse{
	//		Error:   "invalid_request",
	//		Message: "Invalid query parameters",
	//		Details: err.Error(),
	//	})
	//	return
	//}

	//call the service function
	users, err := m.listUser.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "list_users_failed",
			Message: "Could not retrieve users",
			Details: err.Error(),
		})
		return
	}

	//http response
	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Users retrieved successfully",
		Data:    users,
	})
}

// @Summary Get user activity
// @Description Retrieves the activity log for a specific user.
// @Tags Admin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse{data=[]structs.UserActivity} "User activity retrieved successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request (e.g., missing user ID)"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/admin/users/{user_id} [get]
func (m AdminServices) GetUserActivity(c *gin.Context) {
	// Get user ID from URL parameter
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	// Get query parameters for filtering
	//timeRange := c.Query("time_range")
	//activityType := c.Query("activity_type")

	//activities, err := m.AdminServices.GetUserActivity(userID) //call service function here
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
	//		Error:   "activity_fetch_failed",
	//		Message: "Could not fetch user activity",
	//		Details: err.Error(),
	//	})
	//	return
	//}

	mockActivity := []structs.UserActivity{
		{
			UserID:   userID,
			Action:   "login",
			Resource: "system",
		},
		{
			UserID:   userID,
			Action:   "create_case",
			Resource: "case-123",
		},
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User activity retrieved successfully",
		//Data:    activities,
		Data: mockActivity,
	})
}

// @Summary Update a user's role
// @Description Updates the role of a specific user. Only 'Admin' can perform this action.
// @Tags Admin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param request body structs.UpdateUserRoleRequest true "User Role Update Request"
// @Success 200 {object} structs.SuccessResponse "User role updated successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload or user ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/admin/users/{user_id} [put]
func (m AdminServices) UpdateUserRole(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	var req structs.UpdateUserRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid role data",
			Details: err.Error(),
		})
		return
	}

	err := m.updateUserRole.UpdateUserRole(userID, req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "role_update_failed",
			Message: "Could not update user role",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User role updated successfully",
	})
}

// @Summary Get all user roles
// @Description Retrieves a list of all available user roles and their associated permissions.
// @Tags Admin
// @Accept json
// @Produce json
// @Success 200 {object} structs.SuccessResponse{data=[]structs.UserRole} "Roles retrieved successfully"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/admin/roles [get]
func (m AdminServices) GetRoles(c *gin.Context) {
	//roles, err := m.GetRoles()
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
	//		Error:   "roles_fetch_failed",
	//		Message: "Could not fetch roles",
	//		Details: err.Error(),
	//	})
	//	return
	//}

	mockRoles := []structs.UserRole{
		{ID: "1", Name: "Incident Responder", Permissions: []string{"read_cases", "create_cases"}},
		{ID: "2", Name: "Forensic Analyst", Permissions: []string{"read_cases", "analyze_evidence"}},
		{ID: "3", Name: "DFIR Manager", Permissions: []string{"read_cases", "create_cases", "manage_users"}},
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Roles retrieved successfully",
		//Data:    roles, //arr
		Data: mockRoles,
	})
}

//DeleteUser

// @Summary Delete a user
// @Description Deletes a specific user from the system. Only 'Admin' can perform this action.
// @Tags Admin
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} structs.SuccessResponse "User deleted successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload or user ID"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/admin/users/{user_id} [delete]
func (m AdminServices) DeleteUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "User ID is required",
		})
		return
	}

	// Get requester's role from context
	role, exists := c.Get("userRole")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Role information not found",
		})
		return
	}

	// Create the delete user request
	req := delete_user.DeleteUserRequest{
		UserID: userID,
	}

	err := m.deleteUser.DeleteUser(req, role.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "user_deletion_failed",
			Message: "Could not delete user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User deleted successfully",
	})
}
