package registration

// UserModel represents the domain model used in the business logic layer.
// It excludes database-specific concerns like ID and timestamps.
type UserModel struct {
	FullName     string
	Email        string
	PasswordHash string
	Role         string // ENUM: Incident Responder, Forensic Analyst, etc.
}



// NewUserModel constructs a new UserModel from a RegistrationRequest DTO and hashed password.
// This allows clean separation between web-layer input and domain-level logic.
func NewUserModel(req RegistrationRequest, hash string) UserModel {
	return UserModel{
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
	}	
}
