package handlers

import (
	"aegis-api/services/login/auth"
	"aegis-api/services/reset_password"
	"aegis-api/structs"
	"github.com/gin-gonic/gin"
	"net/http"
)

type AuthServices struct {
	authService          *auth.AuthService
	passwordReset        *reset_password.PasswordResetService
	passwordResetRequest *reset_password.PasswordResetService
}

func NewAuthHandler(
	authService *auth.AuthService,
	passwordReset *reset_password.PasswordResetService,
	passwordResetRequest *reset_password.PasswordResetService,

) *AuthServices {
	return &AuthServices{
		authService:          authService,
		passwordReset:        passwordReset,
		passwordResetRequest: passwordResetRequest,
	}
}

type EmailSender interface {
	SendPasswordResetEmail(email string, token string) error
}

// @Summary User login
// @Description Authenticates a user and returns a JWT token and user details upon successful login
// @Tags Authentication
// @Accept json
// @Accept x-www-form-urlencoded
// @Accept multipart/form-data
// @Produce json
// @Param request body structs.LoginRequest true "User login credentials (email and password)"
// @Success 200 {object} structs.SuccessResponse{data=auth.LoginResponse} "Login successful"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 401 {object} structs.ErrorResponse "Authentication failed (invalid credentials)"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/auth/login [post]
func (m AuthServices) Login(c *gin.Context) {
	var req structs.LoginRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid input",
			Details: err.Error(),
		})
		return
	}

	response, err := m.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "login_failed",
			Message: "Invalid credentials",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    response,
	})
}

// @Summary User logout
// @Description Logs out the currently authenticated user by invalidating their session token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} structs.SuccessResponse "Logged out successfully"
// @Failure 401 {object} structs.ErrorResponse "Unauthorized (user not authenticated)"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/auth/logout [delete]
func (m AuthServices) Logout(c *gin.Context) { //COME BACK TO THIS

	//_, exists := c.Get("userID") //_ -> userID
	//if !exists {
	//	c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
	//		Error:   "unauthorized",
	//		Message: "User not authenticated",
	//	})
	//	return
	//}
	//err := m.authService.Logout(userID.(string))
	//if err != nil {
	//	c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
	//		Error:   "logout_failed",
	//		Message: "Could not log out user",
	//		Details: err.Error(),
	//	})
	//	return
	//}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

//requestpasswordreset -- UNDER REVIEW
/*
// @Summary Request password reset
// @Description Requests a password reset email to be sent to the provided email address
// @Tags Authentication
// @Accept json
// @Accept x-www-form-urlencoded
// @Accept multipart/form-data
// @Produce json
// @Param request body structs.PasswordResetRequestBody true "Password reset request with email"
// @Success 200 {object} structs.SuccessResponse "Reset email sent (returns success regardless of whether email exists for security)"
// @Failure 400 {object} structs.ErrorResponse "Invalid request payload"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/auth/reset-password/request [post]
func (m AuthServices) RequestPasswordReset(c *gin.Context) {
	var req structs.ResetPasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid reset password data",
			Details: err.Error(),
		})
		return
	}

	err := m.passwordResetRequest.RequestPasswordReset(userID, req.Email) //should not take in a userID, just email
	if err != nil {                                                       //log failure
		c.JSON(http.StatusOK, structs.SuccessResponse{
			Success: true,
			Message: "If an account with that email exists, a password reset link has been sent.",
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "If an account with that email exists, a password reset link has been sent.",
	})
	return
}
*/

// @Summary Reset password
// @Description Resets a user's password using a valid reset token
// @Tags Authentication
// @Accept json
// @Accept x-www-form-urlencoded
// @Accept multipart/form-data
// @Produce json
// @Param request body structs.PasswordResetBody true "Password reset with token and new password"
// @Success 200 {object} structs.SuccessResponse "Password reset successfully"
// @Failure 400 {object} structs.ErrorResponse "Invalid token or password requirements not met"
// @Failure 500 {object} structs.ErrorResponse "Internal server error"
// @Router /api/v1/auth/reset-password [post]
func (m AuthServices) ResetPassword(c *gin.Context) {
	var req structs.PasswordResetBody
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Invalid reset password data",
			Details: err.Error(),
		})
		return
	}

	err := m.passwordReset.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "invalid or expired token" || err.Error() == "token has expired" {
			status = http.StatusBadRequest
		}

		c.JSON(status, structs.ErrorResponse{
			Error:   "reset_failed",
			Message: "Could not reset password",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Password reset successfully",
	})
}
