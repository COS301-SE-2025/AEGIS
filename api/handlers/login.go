package handlers

import (
	x3dh "aegis-api/internal/x3dh"
	"aegis-api/middleware"
	"aegis-api/pkg/websocket"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/auth/login"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/auth/reset_password"
	"aegis-api/services_/notification"
	"aegis-api/structs"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthHandler struct {
	authService          *login.AuthService
	passwordResetService *reset_password.PasswordResetService
	userRepo             registration.UserRepository
	auditLogger          *auditlog.AuditLogger
	validator            *validator.Validate
}
type ChangePasswordRequest struct {
	OldPassword     string `json:"oldPassword" validate:"required"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
	ConfirmPassword string `json:"confirmPassword" validate:"required"`
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
		validator:            validator.New(),
	}
}

type Handler struct {
	AdminService              AdminServiceInterface
	AuthService               AuthServiceInterface
	CaseService               CaseServiceInterface
	EvidenceService           EvidenceServiceInterface
	UserService               UserServiceInterface
	CaseHandler               *CaseHandler
	CaseDeletionHandler       *CaseDeletionHandler
	CaseListHandler           interface{} // TODO: Replace 'interface{}' with the actual CaseListHandler type when defined/imported
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
	HealthHandler             *HealthHandler

	ReportHandler       *ReportHandler       // Optional: Report generation handler
	ReportStatusHandler *ReportStatusHandler // Optional: Report status update handler

	ReportAIHandler *ReportAIHandler

	IOCHandler            *IOCHandler
	TimelineHandler       *TimelineHandler
	TimelineAIHandler     *TimelineAIHandler
	EvidenceHandler       *EvidenceHandler
	ChainOfCustodyHandler *ChainOfCustodyHandler
	X3DHService           *x3dh.BundleService // Add this
	VerificationHandler   *VerificationHandler
}

func NewHandler(
	adminSvc AdminServiceInterface,
	authSvc AuthServiceInterface,
	caseSvc CaseServiceInterface,
	evidenceSvc EvidenceServiceInterface,
	userSvc UserServiceInterface,
	caseHandler *CaseHandler,
	caseDeletionHandler *CaseDeletionHandler,
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

	reportHandler *ReportHandler, // Optional: Report generation handler
	reportStatusHandler *ReportStatusHandler, // Optional: Report status update handler
	ReportAIHandler *ReportAIHandler,
	IOCHandler *IOCHandler,
	TimelineHandler *TimelineHandler,
	TimelineAIHandler *TimelineAIHandler,

	EvidenceHandler *EvidenceHandler,
	ChainOfCustodyHandler *ChainOfCustodyHandler,

	healthHandler *HealthHandler,

	x3dhService *x3dh.BundleService,
	verificationHandler *VerificationHandler,

) *Handler {
	return &Handler{
		AdminService:              adminSvc,
		AuthService:               authSvc,
		CaseService:               caseSvc,
		EvidenceService:           evidenceSvc,
		UserService:               userSvc,
		CaseHandler:               caseHandler,
		CaseDeletionHandler:       caseDeletionHandler,
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

		ReportHandler:       reportHandler,       // Optional: Report generation handler
		ReportStatusHandler: reportStatusHandler, // Optional: Report status update handler
		ReportAIHandler:     ReportAIHandler,

		IOCHandler:            IOCHandler,
		TimelineHandler:       TimelineHandler,
		TimelineAIHandler:     TimelineAIHandler,
		EvidenceHandler:       EvidenceHandler,
		ChainOfCustodyHandler: ChainOfCustodyHandler,
		HealthHandler:         healthHandler,

		X3DHService:         x3dhService,
		VerificationHandler: verificationHandler,
	}
}

func (h *AuthHandler) LoginHandler(c *gin.Context) {
	// Debug: Print all request headers
	for k, v := range c.Request.Header {
		fmt.Printf("[DEBUG] Header: %s = %v\n", k, v)
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("[DEBUG] Body bind error: %v\n", err)
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	fmt.Printf("[DEBUG] Body: email=%s, password=%s\n", req.Email, req.Password)

	// Add detailed error handling around the service call
	resp, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		// Log the actual error details
		fmt.Printf("[ERROR] Login service error: %v\n", err)
		fmt.Printf("[ERROR] Error type: %T\n", err)

		status := "FAILED"
		h.auditLogger.Log(c, auditlog.AuditLog{
			Action: "LOGIN_ATTEMPT",
			Actor: auditlog.Actor{
				ID:   "",
				Role: "",
			},
			Target: auditlog.Target{
				Type: "user",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      status,
			Description: fmt.Sprintf("Failed login attempt: %v", err),
		})

		// Return 500 if it's an unexpected error, 401 for auth failures
		if strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
				Error:   "authentication_failed",
				Message: "Invalid email or password",
			})
		} else {
			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
				Error:   "internal_error",
				Message: "An internal error occurred",
			})
		}
		return
	}

	// Log successful attempt
	status := "SUCCESS"
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

// Add this method to your existing AuthHandler struct
func (h *AuthHandler) LogoutHandler(c *gin.Context) {
	// Get user ID from context (set by AuthMiddleware)
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusOK, gin.H{"message": "Already logged out"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
		return
	}

	// Simple audit log (optional)
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "USER_LOGOUT",
		Actor: auditlog.Actor{
			ID: userID,
		},
		Target: auditlog.Target{
			Type: "user",
			ID:   userID,
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "User logged out",
		Metadata: map[string]string{
			"timestamp":  time.Now().Format(time.RFC3339),
			"ip_address": c.ClientIP(),
			"user_agent": c.GetHeader("User-Agent"),
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Logged out successfully",
		"timestamp": time.Now().UTC(),
	})
}

// Add this method to your existing AuthHandler
func (h *AuthHandler) ChangePasswordHandler(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate request
	if err := h.validator.Struct(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed"})
		return
	}

	// Check if new passwords match
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "New passwords do not match"})
		return
	}

	// Get user ID from context
	userIDInterface, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, ok := userIDInterface.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Get user from database
	var user struct {
		ID           string `db:"id"`
		PasswordHash string `db:"password_hash"`
		Email        string `db:"email"`
		FullName     string `db:"full_name"`
	}
	// Replace lines 353-358:

	query := `SELECT id, password_hash, email, full_name FROM users WHERE id = ?`
	db := h.userRepo.GetDB()
	err := db.Raw(query, userID).Scan(&user).Error
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Current password is incorrect"})
		return
	}

	// Hash new password
	hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Update password in database
	updateQuery := `
        UPDATE users 
        SET password_hash = $1, updated_at = NOW() 
        WHERE id = $2
    `
	result := h.userRepo.GetDB().Exec(updateQuery, string(hashedNewPassword), userID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}

	// Audit log
	h.auditLogger.Log(c, auditlog.AuditLog{
		Action: "PASSWORD_CHANGE",
		Actor: auditlog.Actor{
			ID: userID,
		},
		Target: auditlog.Target{
			Type: "user",
			ID:   userID,
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "User changed password",
		Metadata: map[string]string{
			"timestamp":  time.Now().Format(time.RFC3339),
			"ip_address": c.ClientIP(),
			"user_agent": c.GetHeader("User-Agent"),
		},
	})

	c.JSON(http.StatusOK, gin.H{
		"message":   "Password changed successfully",
		"timestamp": time.Now().UTC(),
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
