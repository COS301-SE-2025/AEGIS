package accept_terms

// Repository interface
type Repository interface {
	IsUserVerified(userID string) (bool, error)
	SetTermsAccepted(userID string) error
}

// Service interface
type Service interface {
	AcceptTerms(userID string) error
}
