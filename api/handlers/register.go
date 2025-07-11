package handlers

import (
	"aegis-api/services_/admin/delete_user"
	"aegis-api/services_/auditlog"
	"aegis-api/services_/auth/registration"
	"aegis-api/services_/case/ListUsers"
	"aegis-api/services_/user/update_user_role"
	"aegis-api/structs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// //"aegis-core/services"
// "aegis-api/services/delete_user"
// evidence "aegis-api/services/evidence/upload"
// "aegis-api/services/registration"
// "aegis-api/services/reset_password"
// auth "aegis-api/services_/auth/login"
// "aegis-api/services_/case/ListUsers"
// "aegis-api/services_/case/case_assign"
// "aegis-api/services_/case/case_creation"
// "aegis-api/services_/user/update_user_role"
type AdminServiceInterface interface {
	RegisterUser(c *gin.Context)
	VerifyEmail(c *gin.Context)
	ListUsers(c *gin.Context)
}
type AuthServiceInterface interface {
	LoginHandler(c *gin.Context)
	RequestPasswordReset(c *gin.Context)
	ResetPasswordHandler(c *gin.Context)
}

// type CaseServiceInterface interface {
// 	CreateCase(c *gin.Context)
// 	AssignCase(c *gin.Context)
// 	GetCaseByID(c *gin.Context)
// }

type EvidenceServiceInterface interface {
	UploadEvidence(c *gin.Context)
	GetEvidenceByID(c *gin.Context)
	DeleteEvidence(c *gin.Context)
}

type UserServiceInterface interface {
	GetUserInfo(c *gin.Context)
	UpdateUserInfo(c *gin.Context)
	GetUserCases(c *gin.Context)
}

type AdminService struct {
	registrationService *registration.RegistrationService
	listUserService     ListUsers.ListUserService
	userService         *update_user_role.UserService
	userDeleteService   *delete_user.UserDeleteService
	auditLogger         *auditlog.AuditLogger
}

func NewAdminService(
	regService *registration.RegistrationService,
	listUserSvc ListUsers.ListUserService,
	userService *update_user_role.UserService,
	userDeleteService *delete_user.UserDeleteService,
	auditLogger *auditlog.AuditLogger,
) *AdminService {
	return &AdminService{
		registrationService: regService,
		listUserService:     listUserSvc,
		userService:         userService,
		userDeleteService:   userDeleteService,
		auditLogger:         auditLogger,
	}
}

// POST /api/v1/register
func (s *AdminService) RegisterUser(c *gin.Context) {
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REGISTER_USER",
			Actor:  auditlog.Actor{}, // anonymous
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Invalid registration payload",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	user, err := s.registrationService.Register(req)
	if err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "REGISTER_USER",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "user_email",
				ID:   req.Email,
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Registration failed: " + err.Error(),
		})
		log.Printf("Registration error: %v", err)
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}

	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "REGISTER_USER",
		Actor:  auditlog.Actor{}, // or you could pass admin ID if session
		Target: auditlog.Target{
			Type: "user",
			ID:   user.ID.String(),
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "User registered successfully",
	})

	userResp := registration.EntityToResponse(user)

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User registered successfully. Please check your email for a verification link.",
		Data:    userResp,
	})
}

// GET /api/v1/auth/verify?token=xyz
// GET /api/v1/auth/verify?token=xyz
func (s *AdminService) VerifyEmail(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "VERIFY_EMAIL",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "email_verification",
				ID:   "",
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Missing verification token",
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "missing_token",
			Message: "Verification token is required",
		})
		return
	}

	err := s.registrationService.VerifyUser(token)
	if err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "VERIFY_EMAIL",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "email_verification",
				ID:   token,
			},
			Service: "auth",
			Status:  "FAILED",
			Description: "Verification failed: " + (func() string {
				if e, ok := err.(error); ok {
					return e.Error()
				}
				return "unknown error"
			})(),
		})
		if e, ok := err.(error); ok {
			c.JSON(http.StatusBadRequest, structs.ErrorResponse{
				Error:   "verification_failed",
				Message: e.Error(),
			})
		} else {
			c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
				Error:   "unexpected_error",
				Message: "An unexpected error occurred while verifying the email.",
			})
		}
		return
	}

	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "VERIFY_EMAIL",
		Actor:  auditlog.Actor{},
		Target: auditlog.Target{
			Type: "email_verification",
			ID:   token,
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "Email verified successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Email verified successfully",
	})
}
