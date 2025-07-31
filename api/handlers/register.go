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

	"github.com/google/uuid"

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
	RegisterTenantUser(c *gin.Context)
	RegisterTeamUser(c *gin.Context)
	VerifyEmail(c *gin.Context)
	AcceptTerms(c *gin.Context)
	ListUsers(c *gin.Context)
	CreateTenant(c *gin.Context)
	CreateTeam(c *gin.Context)
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
type CreateTenantRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateTeamRequest struct {
	TeamName string     `json:"team_name" binding:"required"`
	TenantID *uuid.UUID `json:"tenant_id" binding:"required"`
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
	// Extract tenant ID from context
	tenantIDVal, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID missing from token",
		})
		return
	}
	tenantIDStr, ok := tenantIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID is not a string",
		})
		return
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid tenant ID format",
		})
		return
	}

	// Assign it to the request
	req.TenantID = &tenantID
	teamIDVal, hasTeam := c.Get("teamID")
	if hasTeam {
		if teamIDStr, ok := teamIDVal.(string); ok {
			if teamID, err := uuid.Parse(teamIDStr); err == nil {
				req.TeamID = &teamID
			} else {
				// Optional: log or return error if team ID is invalid
				c.JSON(http.StatusBadRequest, structs.ErrorResponse{
					Error:   "bad_request",
					Message: "Invalid team ID format",
				})
				return
			}
		}
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
				if err != nil {
					return err.Error()
				}
				return "unknown error"
			})(),
		})
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "verification_failed",
			Message: err.Error(),
		})
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

// POST /api/v1/auth/accept-terms
func (s *AdminService) AcceptTerms(c *gin.Context) {
	var req struct {
		Token string `json:"token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ACCEPT_TERMS",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "terms_acceptance",
				ID:   "",
			},
			Service:     "auth",
			Status:      "FAILED",
			Description: "Missing or invalid token: " + err.Error(),
		})

		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: "Token is required",
		})
		return
	}

	err := s.registrationService.AcceptTerms(req.Token)
	if err != nil {
		s.auditLogger.Log(c, auditlog.AuditLog{
			Action: "ACCEPT_TERMS",
			Actor:  auditlog.Actor{},
			Target: auditlog.Target{
				Type: "terms_acceptance",
				ID:   req.Token,
			},
			Service: "auth",
			Status:  "FAILED",
			Description: "Terms acceptance failed: " + (func() string {
				if err != nil {
					return err.Error()
				}
				return "unknown error"
			})(),
		})

		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "accept_terms_failed",
			Message: err.Error(),
		})
		return
	}

	s.auditLogger.Log(c, auditlog.AuditLog{
		Action: "ACCEPT_TERMS",
		Actor:  auditlog.Actor{},
		Target: auditlog.Target{
			Type: "terms_acceptance",
			ID:   req.Token,
		},
		Service:     "auth",
		Status:      "SUCCESS",
		Description: "Terms and conditions accepted successfully",
	})

	c.JSON(http.StatusOK, structs.SuccessResponse{
		Success: true,
		Message: "Terms accepted successfully",
	})
}

func (s *AdminService) CreateTenant(c *gin.Context) {
	var req CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	tenant, err := s.registrationService.CreateTenant(req.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Tenant created successfully",
		Data:    tenant,
	})
}

func (s *AdminService) CreateTeam(c *gin.Context) {
	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	team, err := s.registrationService.CreateTeam(req.TeamName, req.TenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "create_failed",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Team created successfully",
		Data:    team,
	})
}
func (s *AdminService) RegisterTenantUser(c *gin.Context) {
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}
	user, err := s.registrationService.RegisterTenantUser(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "Tenant and user registered successfully.",
		Data:    registration.EntityToResponse(user),
	})
}

func (s *AdminService) RegisterTeamUser(c *gin.Context) {
	var req registration.RegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}
	tenantIDVal, exists := c.Get("tenantID")
	if !exists {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID missing from token",
		})
		return
	}

	tenantIDStr, ok := tenantIDVal.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, structs.ErrorResponse{
			Error:   "unauthorized",
			Message: "Tenant ID is not a string",
		})
		return
	}

	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, structs.ErrorResponse{
			Error:   "bad_request",
			Message: "Invalid tenant ID format",
		})
		return
	}
	req.TenantID = &tenantID
	user, err := s.registrationService.RegisterTeamUser(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, structs.ErrorResponse{
			Error:   "registration_failed",
			Message: err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, structs.SuccessResponse{
		Success: true,
		Message: "User registered successfully",
		Data:    user,
	})
}
