package case_creation

import (
	"aegis-api/pkg/websocket"
	"aegis-api/services_/notification"
	"errors"
	"time"

	"context"

	"github.com/google/uuid"
)

// Service handles business logic for case creation.
// Service handles business logic for case creation.
type Service struct {
	repo                CaseRepository
	notificationService *notification.NotificationService
	hub                 *websocket.Hub
}

// NewCaseService constructs a new CaseService.
func NewCaseService(repo CaseRepository, notifService *notification.NotificationService, hub *websocket.Hub) *Service {
	return &Service{
		repo:                repo,
		notificationService: notifService,
		hub:                 hub,
	}
}

// CaseRepository defines persistence operations for cases

// CreateCase validates and creates a new case.
func (s *Service) CreateCase(req *CreateCaseRequest) (*Case, error) {
	// Validate title and team name
	if req.Title == "" {
		return nil, errors.New("title is required")
	}
	if req.TeamName == "" {
		return nil, errors.New("team name is required")
	}

	creatorUUID := req.CreatedBy
	// Construct new Case
	newCase := &Case{
		ID:                 uuid.New(),
		Title:              req.Title,
		Description:        req.Description,
		Status:             req.Status,
		Priority:           req.Priority,
		InvestigationStage: req.InvestigationStage,
		CreatedBy:          creatorUUID, // Use the resolved user ID as uuid.UUID
		TeamName:           req.TeamName,
		CreatedAt:          time.Now(),
		TenantID:           req.TenantID, // Assuming this is a tenant ID for multi-tenancy
		TeamID:             req.TeamID,   // Optional team ID
		UpdatedAt:          time.Now(),   // Set initial update time
		Progress:           GetProgressForStage(req.InvestigationStage),
	}

	// Persist the case in the repository
	if err := s.repo.CreateCase(newCase); err != nil {
		return nil, err
	}

	// ✅ Trigger a notification for the case creator
	// after s.repo.CreateCase(newCase) succeeds

	// Only notify if deps are wired
	if s.hub != nil && s.notificationService != nil {
		go websocket.NotifyUser(
			s.hub,
			s.notificationService,
			creatorUUID.String(),
			req.TenantID.String(),
			req.TeamID.String(),
			"Case Created",
			`Your case "`+req.Title+`" has been created successfully.`,
		)
	}

	return newCase, nil
}

// GetCaseByID fetches a case by its ID (and optionally tenant/team if needed)
// Implements: GetCaseByID(ctx context.Context, caseID string) (any, error)
func (s *Service) GetCaseByID(ctx context.Context, caseID string) (any, error) {
	id, err := uuid.Parse(caseID)
	if err != nil {
		return nil, err
	}
	caseObj, err := s.repo.GetCaseByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return caseObj, nil
}
