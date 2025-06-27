package handlers

// import (
// 	"aegis-api/middleware"
// 	"aegis-api/services/ListCases"
// 	"aegis-api/structs"
// 	"aegis-core/services/GetUpdate_Users"
// 	"github.com/gin-gonic/gin"
// 	"net/http"
// )

// type UserService struct {
// 	updateUser *GetUpdate_Users.UserService
// 	userCases  *ListCases.Service
// }

// func NewUserService(
// 	updateUser *GetUpdate_Users.UserService,
// 	userCases *ListCases.Service,
// ) *UserService {
// 	return &UserService{
// 		updateUser: updateUser,
// 		userCases:  userCases,
// 	}
// }

// // @Summary Get user profile
// // @Description Retrieves the profile of the current user or a specific user for admins
// // @Tags Users, Admin
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param user_id path string false "User ID (required for admin access to other users)"
// // @Success 200 {object} structs.SuccessResponse{data=GetUpdate_Users.UserProfile} "User profile retrieved successfully"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/users/profile [get]
// // @Router /api/v1/admin/users/{user_id} [get]
// func (m UserService) GetProfile(c *gin.Context) {
// 	userID, ok := middleware.GetTargetUserID(c)
// 	if !ok {
// 		return
// 	}

// 	userProfile, err := m.updateUser.GetProfile(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "user_info_failed",
// 			Message: "Could not retrieve user information",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User information retrieved successfully",
// 		Data:    userProfile,
// 	})
// }

// // @Summary Update user profile
// // @Description Updates the profile of the current user or a specific user for admins
// // @Tags Users, Admin
// // @Accept json
// // @Accept x-www-form-urlencoded
// // @Accept multipart/form-data
// // @Produce json
// // @Security ApiKeyAuth
// // @Param user_id path string false "User ID (required for admin access to other users)"
// // @Param request body structs.UpdateProfileRequest true "Profile Update Request"
// // @Success 200 {object} structs.SuccessResponse "Profile updated successfully"
// // @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden"
// // @Failure 409 {object} structs.ErrorResponse "Email already exists"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/users/profile [put]
// // @Router /api/v1/admin/users/{user_id}/profile [put]
// func (m UserService) UpdateProfile(c *gin.Context) {
// 	userID, ok := middleware.GetTargetUserID(c)
// 	if !ok {
// 		return
// 	}

// 	var req structs.UpdateProfileRequest
// 	if err := c.ShouldBind(&req); err != nil {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "Invalid request body",
// 			Details: err.Error(),
// 		})

// 		return
// 	}

// 	updates := make(map[string]interface{})

// 	if req.FullName != "" {
// 		updates["full_name"] = req.FullName
// 	}
// 	if req.Email != "" {
// 		updates["email"] = req.Email
// 	}

// 	if len(updates) == 0 {
// 		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
// 			Error:   "invalid_request",
// 			Message: "No valid fields to update",
// 		})
// 		return
// 	}

// 	err := m.updateUser.UpdateProfile(userID, updates) //restricting the updates to the email and the role ONLY
// 	if err != nil {
// 		if err.Error() == "user not found" {
// 			c.JSON(http.StatusNotFound, structs.ErrorResponse{
// 				Error:   "user_not_found",
// 				Message: "The specified user does not exist",
// 				Details: err.Error(),
// 			})
// 			return
// 		}
// 		if err.Error() == "email already exists" {
// 			c.JSON(http.StatusConflict, structs.ErrorResponse{
// 				Error:   "email_already_exists",
// 				Message: "The provided email is already associated with another account",
// 				Details: err.Error(),
// 			})
// 			return
// 		}

// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "profile_update_failed",
// 			Message: "Could not update user profile",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "Profile updated successfully",
// 	})
// }

// // @Summary Get user roles
// // @Description Retrieves all roles associated with a user. Admins can access any user's roles, regular users can only access their own.
// // @Tags Users, Admin
// // @Accept json
// // @Produce json
// // @Security ApiKeyAuth
// // @Param user_id path string false "User ID (required for admin access to other users)"
// // @Success 200 {object} structs.SuccessResponse{data=[]string} "User roles retrieved successfully"
// // @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// // @Failure 403 {object} structs.ErrorResponse "Forbidden"
// // @Failure 500 {object} structs.ErrorResponse "Internal server error"
// // @Router /api/v1/users/roles [get]
// // @Router /api/v1/admin/users/{user_id}/roles [get]
// func (m UserService) GetUserRoles(c *gin.Context) {
// 	userID, ok := middleware.GetTargetUserID(c) //either admin or user accessing their own roles
// 	if !ok {
// 		return
// 	}

// 	roles, err := m.updateUser.GetUserRoles(userID)
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
// 			Error:   "roles_fetch_failed",
// 			Message: "Could not fetch user roles",
// 			Details: err.Error(),
// 		})
// 		return
// 	}

// 	c.JSON(http.StatusOK, structs.SuccessResponse{
// 		Success: true,
// 		Message: "User roles retrieved successfully",
// 		Data:    roles,
// 	})
// }
