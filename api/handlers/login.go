package handlers

import (
	"aegis-api/middleware"
	"aegis-api/pkg/websocket"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/auth/login"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/auth/reset_password"
	"aegis-api/services_/notification"
	"aegis-api/structs"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService          *login.AuthService
	passwordResetService *reset_password.PasswordResetService
	userRepo             registration.UserRepository
	auditLogger          *auditlog.AuditLogger
}

func NewAuthHandler(
	authService *login.AuthService,
	resetService *reset_password.PasswordResetService,
	userRepo registration.UserRepository,
	auditLogger *auditlog.AuditLogger,
) *AuthHandler {
	return &AuthHandler{
		authService:          authService,
		passwordResetService: resetService,
		userRepo:             userRepo,
		auditLogger:          auditLogger,
	}
}

type Handler struct {
	AdminService              AdminServiceInterface
	AuthService               AuthServiceInterface
	CaseService               CaseServiceInterface
	EvidenceService           EvidenceServiceInterface
	UserService               UserServiceInterface
	CaseHandler               *CaseHandler
	UploadHandler             *UploadHandler
	DownloadHandler           *DownloadHandler
	MetadataHandler           *MetadataHandler
	MessageHandler            *MessageHandler
	AnnotationThreadHandler   *AnnotationThreadHandler
	ChatHandler               *ChatHandler
	ProfileHandler            *ProfileHandler
	GetCollaboratorsHandler   *GetCollaboratorsHandler
	EvidenceViewerHandler     *EvidenceViewerHandler
	EvidenceTagHandler        *EvidenceTagHandler
	PermissionChecker         middleware.PermissionChecker
	CaseTagHandler            *CaseTagHandler
	CaseEvidenceTotalsHandler *CaseEvidenceTotalsHandler
	WebSocketHub              *websocket.Hub
	RecentActivityHandler     *RecentActivityHandler
	TeamRepo                  registration.TeamRepository //
	TenantRepo                registration.TenantRepository
	UserRepo                  registration.UserRepository // Optional, if you have a user repository
	NotificationService       *notification.NotificationService
	IOCHandler                *IOCHandler
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
	MessageHandler *MessageHandler,
	annotationThreadHandler *AnnotationThreadHandler,
	chatHandler *ChatHandler,
	profileHandler *ProfileHandler,
	getCollaboratorsHandler *GetCollaboratorsHandler,
	evidenceViewerHandler *EvidenceViewerHandler,
	evidenceTagHandler *EvidenceTagHandler,
	permissionChecker middleware.PermissionChecker,
	caseTagHandler *CaseTagHandler,
	CaseEvidenceTotalsHandler *CaseEvidenceTotalsHandler,
	webSocketHub *websocket.Hub,
	recentActivityHandler *RecentActivityHandler,
	teamRepo registration.TeamRepository,
	tenantRepo registration.TenantRepository, // Optional, if you have a tenant repository
	userRepo registration.UserRepository, // Optional, if you have a user repository
	notificationService *notification.NotificationService,
	IOCHandler *IOCHandler,
) *Handler {
	return &Handler{
		AdminService:              adminSvc,
		AuthService:               authSvc,
		CaseService:               caseSvc,
		EvidenceService:           evidenceSvc,
		UserService:               userSvc,
		CaseHandler:               caseHandler,
		UploadHandler:             uploadHandler,
		DownloadHandler:           downloadHandler,
		MetadataHandler:           metadataHandler,
		MessageHandler:            MessageHandler,
		AnnotationThreadHandler:   annotationThreadHandler,
		ChatHandler:               chatHandler,
		ProfileHandler:            profileHandler,
		GetCollaboratorsHandler:   getCollaboratorsHandler,
		EvidenceViewerHandler:     evidenceViewerHandler,
		EvidenceTagHandler:        evidenceTagHandler,
		PermissionChecker:         permissionChecker,
		CaseTagHandler:            caseTagHandler,
		CaseEvidenceTotalsHandler: CaseEvidenceTotalsHandler,
		WebSocketHub:              webSocketHub,
		RecentActivityHandler:     recentActivityHandler,
		TeamRepo:                  teamRepo,
		TenantRepo:                tenantRepo, // Optional, if you have a tenant repository
		UserRepo:                  userRepo,   // Optional, if you have a user repository
		NotificationService:       notificationService,
		IOCHandler:                IOCHandler,
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
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LOGIN_ATTEMPT",
			Actor: auditlog.Actor{
				ID:   "", // Unknown until login successful
				Role: "",
			},
			Target: auditlog.Target{
				Type: "user",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      status,
			Description: "Failed login attempt",
		})
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "authentication_failed",
			Message: "Invalid email or password",
		})
		return
	}

	// Log successful attempt
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "LOGIN_ATTEMPT",
		Actor: auditlog.Actor{
			ID:    resp.ID,
			Role:  resp.Role,
			Email: resp.Email,
		},
		Target: auditlog.Target{
			Type: "user",
			ID:   resp.ID,
		},
		Service:     "auth",
		Status:      status,
		Description: "User logged in successfully",
	})

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
		// Invalid request payload
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REQUEST_PASSWORD_RESET",
			Actor:  auditlog.Actor{}, // anonymous actor
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Invalid request payload",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	user, err := h.userRepo.GetUserByEmail(req.Email)
	if err != nil {
		// User not found
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REQUEST_PASSWORD_RESET",
			Actor:  auditlog.Actor{}, // unknown user
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "No user found with provided email",
		})
		c.JSON(http.StatusNotFound, structs.ErrorResponse{
			Error:   "user_not_found",
			Message: "No user found with that email",
		})
		return
	}

	// Defensive: you have an unnecessary check
	// (uid := user.ID will never fail since it's a string)
	uid := user.ID

	// Attempt to send reset
	err = h.passwordResetService.RequestPasswordReset(uid, user.Email)
	if err != nil {
		// Failed to send email
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REQUEST_PASSWORD_RESET",
			Actor: auditlog.Actor{
				ID:    user.ID.String(),
				Role:  user.Role,
				Email: user.Email, // Include email for better tracking
			},
			Target: auditlog.Target{
				Type: "user",
				ID:   user.ID.String(),
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Failed to send password reset email",
		})
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "reset_failed",
			Message: "Failed to send password reset email",
		})
		return
	}

	// Success
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "REQUEST_PASSWORD_RESET",
		Actor: auditlog.Actor{
			ID:    user.ID.String(),
			Role:  user.Role,
			Email: user.Email, // Include email for better tracking
		},
		Target: auditlog.Target{
			Type: "user",
			ID:   user.ID.String(),
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "Password reset email sent successfully",
	})

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
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "RESET_PASSWORD",
			Actor:  auditlog.Actor{}, // unknown actor
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Invalid request payload for password reset",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	err := h.passwordResetService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "RESET_PASSWORD",
			Actor:  auditlog.Actor{}, // anonymous actor
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Password reset failed: " + err.Error(),
		})
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "reset_failed",
			Message: err.Error(),
		})
		return
	}

	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "RESET_PASSWORD",
		Actor:  auditlog.Actor{}, // optionally fill with session user if you have
		Target: auditlog.Target{
			Type: "user_email",
			ID:   req.Email,
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "Password reset successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Password reset successful",
	})
}
