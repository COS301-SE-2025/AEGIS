package handlers

import (
	"aegis-api/middleware"
	"aegis-api/services/ListCases"
	"aegis-api/structs"
	"aegis-core/services/GetUpdate_Users"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserService struct {
	updateUser *GetUpdate_Users.UserService
	userCases  *ListCases.Service
}

func NewUserService(
	updateUser *GetUpdate_Users.UserService,
	userCases *ListCases.Service,
) *UserService {
	return &UserService{
		updateUser: updateUser,
		userCases:  userCases,
	}
}

// @Summary Get user profile
// @Description Retrieves a user profile. Admins can access any user's profile, regular users can only access their own.
// @Tags Users, Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string false "User ID (required for admin access to other users)"
// @Success 200 {object} structs.SuccessResponse{data=structs.User} "User profile retrieved successfully"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/users [get]
// @Router /api/v1/admin/users/{user_id} [get]
func (m UserService) GetProfile(c *gin.Context) {

	//check which path was used to access the user profile
	/*targetUserID := c.Param("user_id") move to middleware
	role, _ := c.Get("userRole")
	var userID string
	if targetUserID != "" && role == "Admin" {
		userID = targetUserID //admin to view any user profile
	} else {
		currUserID, exists := c.Get("userID") //user to view own profile
		if !exists {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "unauthorized",
				Message: "User not authenticated",
			})
			return
		}
		userID = currUserID.(string)
	}*/
	userID, ok := middleware.GetTargetUserID(c)
	if !ok {
		return
	}

	userProfile, err := m.updateUser.GetProfile(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "user_info_failed",
			Message: "Could not retrieve user information",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User information retrieved successfully",
		Data:    userProfile,
	})
}

// @Summary Update user profile
// @Description Updates a user's profile information. Admins can update any user's profile, regular users can only update their own.
// @Tags Users, Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string false "User ID (required for admin access to other users)"
// @Param request body structs.UpdateProfileRequest true "Profile update information"
// @Success 200 {object} structs.SuccessResponse "Profile updated successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden"
// @Failure 404 {object} structs.ErrorResponse "User not found"
// @Failure 409 {object} structs.ErrorResponse "Email already exists"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/users [put]
// @Router /api/v1/admin/users/{user_id}/profile [put]
func (m UserService) UpdateProfile(c *gin.Context) {

	userID, ok := middleware.GetTargetUserID(c)
	if !ok {
		return
	}

	var req structs.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})

	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No valid fields to update",
		})
		return
	}

	err := m.updateUser.UpdateProfile(userID, updates) //restricting the updates to the email and the role ONLY
	if err != nil {
		if err.Error() == "user not found" {
			c.JSON(http.StatusNotFound, structs.ErrorResponse{
				Error:   "user_not_found",
				Message: "The specified user does not exist",
				Details: err.Error(),
			})
			return
		}
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, structs.ErrorResponse{
				Error:   "email_already_exists",
				Message: "The provided email is already associated with another account",
				Details: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "profile_update_failed",
			Message: "Could not update user profile",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// @Summary Get user cases
// @Description Retrieves cases associated with a user. Admins can access any user's cases, regular users can only access their own.
// @Tags Users, Admin
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path string false "User ID (required for admin access to other users)"
// @Success 200 {object} structs.SuccessResponse{data=[]structs.Case} "User cases retrieved successfully"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 403 {object} structs.ErrorResponse "Forbidden"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/users/cases [get]
// @Router /api/v1/admin/users/{user_id}/cases [get]
func (m UserService) GetUserCases(c *gin.Context) {
	userID, ok := middleware.GetTargetUserID(c)
	if !ok {
		return
	}

	cases, err := m.userCases.GetCasesByUser(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "cases_fetch_failed",
			Message: "Could not fetch user cases",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User cases retrieved successfully",
		Data:    cases,
	})
}
