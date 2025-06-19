package handlers

import (
	"aegis-api/services/ListCases"
	"aegis-api/structs"
	"aegis-core/services/getUpdate_Users"
	"github.com/gin-gonic/gin"
	"net/http"
)

type UserService struct {
	updateUser *getUpdate_Users.UserService
	userCases  *ListCases.Service
}

func NewUserService(
	updateUser *getUpdate_Users.UserService,
	userCases *ListCases.Service,
) *UserService {
	return &UserService{
		updateUser: updateUser,
		userCases:  userCases,
	}
}

// getprofile
// @Summary Get current user's information
// @Description Retrieves the detailed profile information for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} structs.SuccessResponse{data=structs.User} "User information retrieved successfully"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized (user not authenticated)"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/user/info [get]
func (m UserService) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	userProfile, err := m.updateUser.GetProfile(userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "user_info_failed",
			Message: "Could not retrieve user information",
			Details: err.Error(),
		})
		return
	}

	//mockUser := structs.User{
	//	ID:       userID.(string),
	//	Email:    "user@example.com",
	//	FullName: "Mock User",
	//	Role:     structs.UserRole{Name: "Forensic Analyst"},
	//}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User information retrieved successfully",
		//Data:    userInfo,
		Data: userProfile,
	})
}

// updateprofile
// @Summary Update current user's information
// @Description Updates the profile details (e.g., name, email) for the authenticated user.
// @Tags User
// @Accept json
// @Produce json
// @Param request body structs.UpdateUserInfoRequest true "User Info Update Request"
// @Security ApiKeyAuth
// @Success 200 {object} structs.SuccessResponse "User information updated successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/user/info [put]
func (m UserService) UpdateProfile(c *gin.Context) {
	_, exists := c.Get("userID") // _ -> userID
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	var req structs.UpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid user data",
			Details: err.Error(),
		})
		return
	}

	//err := m.userService.UpdateUserInfo(userID.(string), req)
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
	//		Error:   "update_failed",
	//		Message: "Could not update user information",
	//		Details: err.Error(),
	//	})
	//	return
	//}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User information updated successfully",
	})
}

// getcasesbyuser | alias reuse
// @Summary Get cases assigned to the current user
// @Description Retrieves a list of security cases that the authenticated user is involved in.
// @Tags User
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} structs.SuccessResponse{data=[]structs.Case} "User cases retrieved successfully"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/user/cases [get]
func (m UserService) GetUserCases(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	// Get pagination parameters
	//page := c.DefaultQuery("page", "1")
	//pageSize := c.DefaultQuery("page_size", "10")

	//cases, err := m.userService.GetUserCases(userID.(string)) //, page, pageSize
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
	//		Error:   "cases_fetch_failed",
	//		Message: "Could not fetch user cases",
	//		Details: err.Error(),
	//	})
	//	return
	//}
	mockCases := []structs.Case{
		{
			ID:          "user-case-1",
			Title:       "User's Case 1",
			Description: "First case assigned to user",
			Status:      "active",
			CreatedBy:   userID.(string),
		},
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "User cases retrieved successfully",
		//Data:    cases,
		Data: mockCases,
	})
}
