package reset_password

// PasswordResetService orchestrates password reset logic using repositories and email services.
type PasswordResetService struct {
	repo    ResetTokenRepository
	users   UserRepository
	emailer EmailSender
}
