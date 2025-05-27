package services

//stub service layer functions
import "aegis-api/structs"

// AuthService handles authentication operations
type AuthService interface {
	Login(credentials structs.LoginRequest) (structs.LoginResponse, error)
	Logout(userID string) error
	ResetPassword(req structs.ResetPasswordRequest) error
	ValidateToken(token string) (structs.User, error)
}

// CaseService handles case management
type CaseService interface {
	CreateCase(userID string, req structs.CreateCaseRequest) (structs.Case, error)
	GetCases(filter structs.CaseFilter) ([]structs.Case, error)
	GetCase(caseID string) (structs.Case, error)
	UpdateCase(caseID string, req structs.UpdateCaseRequest) error
	AssignCase(caseID string, req structs.AssignCaseRequest) error
	GetCollaborators(caseID string) ([]structs.User, error)
	AddCollaborator(caseID string, req structs.User) error
	RemoveCollaborator(caseID string, userID string) error
}

// EvidenceService handles evidence management
type EvidenceService interface {
	GetEvidence(caseID string, filter structs.EvidenceFilter) ([]structs.EvidenceItem, error)
	UploadEvidence(caseID string, req structs.UploadEvidenceRequest) (structs.EvidenceItem, error)
	GetEvidenceItem(caseID string, evidenceID string) (structs.EvidenceItem, error)
	PreviewEvidence(caseID string, evidenceID string) (structs.EvidencePreview, error)
}

// UserService handles user-specific operations
type UserService interface {
	GetUserInfo(userID string) (structs.User, error)
	UpdateUserInfo(userID string, req structs.UpdateUserInfoRequest) error
	GetUserCases(userID string) ([]structs.Case, error)
}

// AdminService handles administrative operations
type AdminService interface {
	Register(req structs.RegisterUserRequest) (structs.User, error)
	ListUsers(filter structs.UserFilter) ([]structs.User, error)
	GetUserActivity(userID string) ([]structs.UserActivity, error)
	UpdateUserRole(userID string, roleName string) error
	DeleteUser(userID string) error
	GetRoles() ([]structs.UserRole, error)
}
