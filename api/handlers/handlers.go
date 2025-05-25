package handlers

import (
	//"aegis-core/services"
	"github.com/gin-gonic/gin"
	//"net/http"
)

// mock services
type AdminServiceInterface interface {
	RegisterUser(c *gin.Context)
	ListUsers(c *gin.Context)
	GetUserActivity(c *gin.Context)
	UpdateUserRole(c *gin.Context)
	DeleteUser(c *gin.Context)
	GetRoles(c *gin.Context)
}

type AuthServiceInterface interface {
	Login(c *gin.Context)
	Logout(c *gin.Context)
	ResetPassword(c *gin.Context)
}

type CaseServiceInterface interface {
	GetCases(c *gin.Context)
	CreateCase(c *gin.Context)
	GetCase(c *gin.Context)
	UpdateCase(c *gin.Context)
	AssignCase(c *gin.Context)
	GetCollaborators(c *gin.Context)
	CreateCollaborator(c *gin.Context)
	RemoveCollaborator(c *gin.Context)
}

type EvidenceServiceInterface interface {
	GetEvidence(c *gin.Context)
	UploadEvidence(c *gin.Context)
	GetEvidenceItem(c *gin.Context)
	PreviewEvidence(c *gin.Context)
}

type UserServiceInterface interface {
	GetUserInfo(c *gin.Context)
	UpdateUserInfo(c *gin.Context)
	GetUserCases(c *gin.Context)
}

type Handler struct {
	AdminService    AdminServiceInterface
	AuthService     AuthServiceInterface
	CaseService     CaseServiceInterface
	EvidenceService EvidenceServiceInterface
	UserService     UserServiceInterface
}

type MockAdminService struct{}

func (m MockAdminService) RegisterUser(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockAdminService) ListUsers(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockAdminService) GetUserActivity(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockAdminService) UpdateUserRole(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockAdminService) DeleteUser(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockAdminService) GetRoles(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

type MockAuthService struct{}

func (m MockAuthService) Login(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockAuthService) Logout(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockAuthService) ResetPassword(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

type MockCaseService struct{}

func (m MockCaseService) GetCases(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockCaseService) CreateCase(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockCaseService) GetCase(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockCaseService) UpdateCase(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockCaseService) AssignCase(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockCaseService) GetCollaborators(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockCaseService) CreateCollaborator(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockCaseService) RemoveCollaborator(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

type MockEvidenceService struct{}

func (m MockEvidenceService) GetEvidence(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockEvidenceService) UploadEvidence(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockEvidenceService) GetEvidenceItem(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockEvidenceService) PreviewEvidence(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

type MockUserService struct{}

func (m MockUserService) GetUserInfo(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockUserService) UpdateUserInfo(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func (m MockUserService) GetUserCases(c *gin.Context) {
	//TODO implement me
	panic("implement me")
}

func NewHandler() *Handler {
	return &Handler{
		AdminService:    &MockAdminService{},
		AuthService:     &MockAuthService{},
		CaseService:     &MockCaseService{},
		EvidenceService: &MockEvidenceService{},
		UserService:     &MockUserService{},
	}
}
