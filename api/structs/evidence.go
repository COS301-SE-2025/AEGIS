package structs

import "mime/multipart"

// UploadEvidenceRequest defines the structure for the evidence upload request at the API layer
type UploadEvidenceRequest struct {
	Name        string                `form:"name" binding:"required"`
	Type        string                `form:"type" binding:"required"`
	Description string                `form:"description"`
	File        *multipart.FileHeader `form:"file" binding:"required"`
	// CaseID comes from URL parameter
}
