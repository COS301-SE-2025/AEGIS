package registration

// UserModel represents the domain model used in the business logic layer.
// It excludes database-specific concerns like ID and timestamps.
type UserModel struct {
	Name         string
	Surname      string
	Email        string
	PasswordHash string
}

// FullName returns the combined full name of the user.
// This is a domain-level utility that does not belong in DTO or entity layers.
func (u UserModel) FullName() string {
	return u.Name + " " + u.Surname
}

// NewUserModel constructs a new UserModel from a RegistrationRequest DTO and hashed password.
// This allows clean separation between web-layer input and domain-level logic.
func NewUserModel(req RegistrationRequest, hash string) UserModel {
	return UserModel{
		Name:         req.Name,
		Surname:      req.Surname,
		Email:        req.Email,
		PasswordHash: hash,
	}	
}
