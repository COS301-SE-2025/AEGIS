package update_case

import (
	"aegis-api/pkg/websocket"
	"aegis-api/services_/admin/get_collaborators"
	"aegis-api/services_/notification"
	"context"
	"fmt"

	"github.com/google/uuid"
)

type Service struct {
	repo                 UpdateCaseRepository
	collaboratorsService *get_collaborators.Service
	notificationService  *notification.NotificationService
	hub                  *websocket.Hub
}

func NewService(
	repo UpdateCaseRepository,
	collabSvc *get_collaborators.Service,
	notifService *notification.NotificationService,
	hub *websocket.Hub,
) *Service {
	return &Service{
		repo:                 repo,
		collaboratorsService: collabSvc,
		notificationService:  notifService,
		hub:                  hub,
	}
}

func (s *Service) UpdateCaseDetails(ctx context.Context, req *UpdateCaseRequest) (*UpdateCaseResponse, error) {
	// 1. Update case in DB
	if err := s.repo.UpdateCase(ctx, req); err != nil {
		return nil, err
	}

	// 2. Get collaborators for this case
	collaborators, err := s.collaboratorsService.GetCollaborators(uuid.MustParse(req.CaseID))
	if err != nil {
		return nil, err
	}

	// 3. Send notifications to all collaborators
	for _, user := range collaborators {
		_ = websocket.NotifyUser(
			s.hub,
			s.notificationService,
			user.ID.String(),
			user.TenantID,
			user.TeamID,
			"Case Updated",
			fmt.Sprintf("Case %s has been updated.", req.CaseID),
		)
	}

	return &UpdateCaseResponse{Success: true}, nil
}
