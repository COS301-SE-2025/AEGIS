package handlers

import (
	"aegis-api/services_/annotation_threads/messages"
	"aegis-api/services_/auth/login"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/auth/reset_password"
	"aegis-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService          *login.AuthService
	passwordResetService *reset_password.PasswordResetService
	userRepo             registration.UserRepository
}

func NewAuthHandler(
	authService *login.AuthService,
	resetService *reset_password.PasswordResetService,
	userRepo registration.UserRepository,
) *AuthHandler {
	return &AuthHandler{
		authService:          authService,
		passwordResetService: resetService,
		userRepo:             userRepo,
	}
}

type Handler struct {
	AdminService            AdminServiceInterface
	AuthService             AuthServiceInterface
	CaseService             CaseServiceInterface
	EvidenceService         EvidenceServiceInterface
	UserService             UserServiceInterface
	CaseHandler             *CaseHandler
	UploadHandler           *UploadHandler
	DownloadHandler         *DownloadHandler
	MetadataHandler         *MetadataHandler
	MessageService          messages.MessageService
	AnnotationThreadHandler *AnnotationThreadHandler
}

func NewHandler(
	adminSvc AdminServiceInterface,
	authSvc AuthServiceInterface,
	caseSvc CaseServiceInterface,
	evidenceSvc EvidenceServiceInterface,
	userSvc UserServiceInterface,
	caseHandler *CaseHandler,
	uploadHandler *UploadHandler,
	downloadHandler *DownloadHandler, // Optional, if you have a download handler
	metadataHandler *MetadataHandler, // Optional, if you have a metadata handler
	messageService messages.MessageService,
	annotationThreadHandler *AnnotationThreadHandler,
) *Handler {
	return &Handler{
		AdminService:            adminSvc,
		AuthService:             authSvc,
		CaseService:             caseSvc,
		EvidenceService:         evidenceSvc,
		UserService:             userSvc,
		CaseHandler:             caseHandler,
		UploadHandler:           uploadHandler,
		DownloadHandler:         downloadHandler,
		MetadataHandler:         metadataHandler,
		MessageService:          messageService,
		AnnotationThreadHandler: annotationThreadHandler,
	}
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	resp, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid email or password",
		})
		return
	}

	// if !resp.IsVerified {
	// 	c.JSON(http.StatusForbidden, structs.ErrorResponse{
	// 		Error:   "email_not_verified",
	// 		Message: "Please verify your email before logging in.",
	// 	})
	// 	return
	// }

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Login successful",
		Data:    resp,
	})
}

func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Error:   "user_not_found",
			Message: "No user found with that email",
		})
		return
	}

	uid := user.ID
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "internal_error",
			Message: "Invalid user ID format",
		})
		return
	}

	err = h.passwordResetService.RequestPasswordReset(uid, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "reset_failed",
			Message: "Failed to send password reset email",
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Password reset link sent if account exists",
	})
}

func (h *AuthHandler) ResetPasswordHandler(c *gin.Context) {
	var req struct {
		Email       string `json:"email" binding:"required,email"`
		NewPassword string `json:"newPassword" binding:"required"`
		Token       string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err := h.passwordResetService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Password reset successful",
	})
}
