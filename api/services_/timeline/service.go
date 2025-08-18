package timeline

import (
	"gorm.io/datatypes"
)

// Service exposes business methods used by handlers.
type Service interface {
	AddEvent(event *TimelineEvent) (*TimelineEvent, error)
	GetEvent(id string) (*TimelineEvent, error)
	ListEvents(caseID string) ([]*TimelineEventResponse, error)
	UpdateEvent(event *TimelineEvent) (*TimelineEvent, error)
	DeleteEvent(id string) error
	ReorderEvents(caseID string, orderedIDs []string) error
	GetEventByID(eventID string) (*TimelineEvent, error)
}

type timelineService struct {
	repo Repository
}
type TimelineEventResponse struct {
	ID          string         `json:"id"`
	Description string         `json:"description"`
	Severity    string         `json:"severity"`
	AnalystName string         `json:"analystName"`
	Date        string         `json:"date"`
	Time        string         `json:"time"`
	Evidence    datatypes.JSON `json:"evidence"`
	Tags        datatypes.JSON `json:"tags"`
}

func (s *timelineService) ListEvents(caseID string) ([]*TimelineEventResponse, error) {
	events, err := s.repo.ListByCase(caseID)
	if err != nil {
		return nil, err
	}

	var resp []*TimelineEventResponse
	for _, ev := range events {
		resp = append(resp, &TimelineEventResponse{
			ID:          ev.ID,
			Description: ev.Description,
			Severity:    ev.Severity,
			AnalystName: ev.AnalystName,
			Date:        ev.CreatedAt.Format("2006-01-02"),
			Time:        ev.CreatedAt.Format("15:04"),
			Evidence:    ev.Evidence,
			Tags:        ev.Tags,
		})
	}
	return resp, nil
}

func NewService(repo Repository) Service {
	return &timelineService{repo: repo}
}

func (s *timelineService) AddEvent(event *TimelineEvent) (*TimelineEvent, error) {
	if event == nil {
		return nil, ErrInvalidEvent
	}
	// normalize empty arrays for evidence/tags
	if len(event.Evidence) == 0 {
		event.Evidence = datatypes.JSON([]byte("[]"))

	}
	if len(event.Tags) == 0 {
		event.Tags = datatypes.JSON([]byte("[]"))

	}

	if err := s.repo.Create(event); err != nil {
		return nil, err
	}
	return event, nil
}

func (s *timelineService) GetEvent(id string) (*TimelineEvent, error) {
	return s.repo.GetByID(id)
}

func (s *timelineService) UpdateEvent(event *TimelineEvent) (*TimelineEvent, error) {
	if event == nil || event.ID == "" {
		return nil, ErrInvalidEvent
	}
	if err := s.repo.Update(event); err != nil {
		return nil, err
	}
	return s.repo.GetByID(event.ID)
}

func (s *timelineService) DeleteEvent(id string) error {
	return s.repo.Delete(id)
}

func (s *timelineService) ReorderEvents(caseID string, orderedIDs []string) error {
	return s.repo.UpdateOrder(caseID, orderedIDs)
}

// Domain errors
var (
	ErrInvalidEvent = &TimelineError{"invalid timeline event"}
)

type TimelineError struct {
	Msg string
}

func (e *TimelineError) Error() string { return e.Msg }
func (s *timelineService) GetEventByID(eventID string) (*TimelineEvent, error) {
	return s.repo.FindByID(eventID)
}
