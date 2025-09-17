package sharecase

import (
	"time"

	"github.com/google/uuid"
)

// Repository interface
type CaseShareRepository interface {
	CreateToken(senderID, caseID uuid.UUID, expiresIn time.Duration) (string, error)
	GetValidToken(rawToken string) (*Token, error)
	RedeemToken(rawToken string, recipientID, roleID uuid.UUID) error
	ListSharedCases(userID uuid.UUID) ([]Case, error)
	ExpireOldShares() error
}

// Service interface
type CaseShareService interface {
	ShareCase(senderID, caseID uuid.UUID, recipientEmail string) error
	AcceptCaseShare(rawToken string, recipientID uuid.UUID, roleID uuid.UUID) error
	GetSharedCases(userID uuid.UUID) ([]Case, error)
}
