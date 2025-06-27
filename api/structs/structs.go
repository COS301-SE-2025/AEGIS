package structs

// import (
//
//	"github.com/google/uuid"
//	"time"
//
// )
//
//	type LoginRequest struct {
//		Email    string `json:"email" binding:"required,email"`
//		Password string `json:"password" binding:"required"`
//	}
//
//	type LoginResponse struct {
//		Token     string    `json:"token"`
//		ExpiresAt time.Time `json:"expires_at"`
//		User      User      `json:"user"`
//	}
//
//	type User struct {
//		ID         string    `json:"id"`
//		Email      string    `json:"email"`
//		FullName   string    `json:"full_name"`
//		Role       UserRole  `json:"role"`
//		CreatedAt  time.Time `json:"created_at"`
//		IsVerified bool      `json:"is_verified"`
//	}
//
//	type Case struct {
//		ID            string             `json:"id"`
//		Title         string             `json:"title"`
//		Description   string             `json:"description"`
//		Status        string             `json:"status"`
//		CreatedBy     string             `json:"created_by"`
//		CreatedAt     time.Time          `json:"created_at"`
//		Collaborators []CollaboratorInfo `json:"collaborators"`
//	}
//
//	type CollaboratorInfo struct {
//		ID       string `json:"id"`
//		FullName string `json:"full_name"`
//		Role     string `json:"role"` // optional
//	}
//
// // type CreateCaseRequest struct {
// // 	Title       string `json:"title" binding:"required"`
// // 	Description string `json:"description"`
// // }
//
// // CreateCaseRequest is what your handler will bind from JSON.
//
//	type CreateCaseRequest struct {
//		Title              string `json:"title" binding:"required"`
//		Description        string `json:"description"`
//		Status             string `json:"status"`
//		Priority           string `json:"priority"`
//		InvestigationStage string `json:"investigationStage"`
//		CreatedBy          string `json:"createdBy" binding:"required,uuid"`
//		TeamName           string `json:"teamName" binding:"required"`
//	}
//
//	type UpdateCaseRequest struct {
//		Title       string `json:"title"`
//		Description string `json:"description"`
//		Status      string `json:"status"`
//	}
//
//	type UpdateCaseStatusRequest struct {
//		CaseID string `json:"case_id" validate:"required,uuid"`
//		Status string `json:"status" validate:"required"`
//	}
//
//	type AssignCaseRequest struct {
//		UserID string `json:"user_id" binding:"required"`
//		Role   string `json:"role" binding:"required"` //might need to remove
//	}
//
//	type EvidenceItem struct {
//		ID          string    `json:"id"`
//		CaseID      string    `json:"case_id"`
//		Name        string    `json:"name"`
//		Type        string    `json:"type"`
//		Hash        string    `json:"hash"`
//		UploadedBy  string    `json:"uploaded_by"`
//		UploadedAt  time.Time `json:"uploaded_at"`
//		StoragePath string    `json:"storage_path"`
//	}
//
//	type EvidenceFilter struct {
//		Type       string `form:"type"`
//		UploadedBy string `form:"uploaded_by"`
//		StartDate  string `form:"start_date"`
//		EndDate    string `form:"end_date"`
//	}
//
//	type UploadEvidenceRequest struct {
//		Name        string `json:"name" binding:"required"`
//		Type        string `json:"type" binding:"required"`
//		File        []byte `json:"file" binding:"required"`
//		Description string `json:"description"`
//	}
//
//	type EvidencePreview struct {
//		ID           string `json:"id"`
//		Name         string `json:"name"`
//		Type         string `json:"type"`
//		PreviewURL   string `json:"preview_url"`
//		ThumbnailURL string `json:"thumbnail_url,omitempty"`
//	}
//
//	type UserRole struct {
//		ID          string   `json:"id"`
//		Name        string   `json:"name"`
//		Permissions []string `json:"permissions"`
//	}
//
//	type UserFilter struct {
//		Role      string `form:"role"`
//		Status    string `form:"status"`
//		StartDate string `form:"start_date"`
//		EndDate   string `form:"end_date"`
//	}
//
//	type UserActivity struct {
//		UserID    string    `json:"user_id"`
//		Action    string    `json:"action"`
//		Resource  string    `json:"resource"`
//		Timestamp time.Time `json:"timestamp"`
//	}
//
//	type RegisterUserRequest struct {
//		Email    string `json:"email" binding:"required,email"`
//		Password string `json:"password" binding:"required"`
//		FullName string `json:"full_name" binding:"required"`
//		Role     string `json:"role" binding:"required,oneof='Incident Responder' 'Forensic Analyst' 'Malware Analyst' 'Threat Intelligent Analyst' 'DFIR Manager' 'Legal/Compliance Liaison' 'Detection Engineer' 'Generic user'"`
//	}
//
// //type PasswordResetRequestBody struct {
// //	Email string `json:"email" binding:"required,email"`
// //}
// //

//	type UpdateProfileRequest struct {
//		FullName string `json:"full_name,omitempty" binding:"omitempty,min=1"`
//		Email    string `json:"email,omitempty" binding:"omitempty,email"`
//	}
//
//	type UpdateUserRoleRequest struct {
//		Role string `json:"role" binding:"required"`
//	}
//
//	type UpdateUserInfoRequest struct {
//		Name  string `json:"name"`
//		Email string `json:"email" binding:"omitempty,email"`
//	}
//
//	type CaseFilter struct {
//		Status    string `form:"status"`
//		StartDate string `form:"start_date"`
//		EndDate   string `form:"end_date"`
//		Page      string `form:"page"`
//		PageSize  string `form:"page_size"`
//	}
//
//	type ResetTokenRepository interface {
//		CreateToken(userID uuid.UUID, token string, expiresAt time.Time) error
//		GetUserIDByToken(token string) (uuid.UUID, time.Time, error)
//		MarkTokenUsed(token string) error
//	}
//
// Error response structure
type ErrorResponse struct {
	Error   string      `json:"error"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Success response structure
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

//
//type DeleteUserRequest struct {
//	UserID string `json:"user_id"`
//}
