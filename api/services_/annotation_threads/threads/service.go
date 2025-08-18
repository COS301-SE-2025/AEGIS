package annotationthreads

import (
	"aegis-api/pkg/websocket"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Annotationthreadservice implements the AnnotationThreadService interface
type Annotationthreadservice struct {
	repo AnnotationThreadRepository
	hub  *websocket.Hub
}

func NewAnnotationThreadService(repo AnnotationThreadRepository, hub *websocket.Hub) AnnotationThreadService {
	return &Annotationthreadservice{
		repo: repo,
		hub:  hub}
}

func (s *Annotationthreadservice) CreateThread(
	caseID, fileID, userID uuid.UUID,
	title string, tags []string,
	priority ThreadPriority,
) (*AnnotationThread, error) {

	thread := &AnnotationThread{
		ID:        uuid.New(),
		Title:     title,
		CaseID:    caseID,
		FileID:    fileID,
		CreatedBy: userID,
		Status:    StatusOpen,
		Priority:  priority,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Thread + initial participant + tags inserted in one transaction
	err := s.repo.CreateThreadWithParticipant(thread, tags, userID)
	if err != nil {
		return nil, err
	}

	// Notify via WebSocket
	payload := websocket.ThreadCreatedPayload{
		ThreadID:  thread.ID.String(),
		Title:     thread.Title,
		CaseID:    thread.CaseID.String(),
		FileID:    thread.FileID.String(),
		CreatedBy: userID.String(),
		CreatedAt: thread.CreatedAt.Format(time.RFC3339),
		Priority:  string(thread.Priority),
	}
	if err := websocket.SendThreadCreated(s.hub, payload); err != nil {
		return nil, errors.New("failed to send thread created event: " + err.Error())
	}

	return thread, nil
}

func (s *Annotationthreadservice) GetThreadsByFile(fileID uuid.UUID) ([]AnnotationThread, error) {
	return s.repo.GetThreadsByFile(fileID)
}

func (s *Annotationthreadservice) GetThreadsByCase(caseID uuid.UUID) ([]AnnotationThread, error) {
	return s.repo.GetThreadsByCase(caseID)
}

func (s *Annotationthreadservice) AddParticipant(threadID, userID uuid.UUID) error {
	err := s.repo.AddParticipant(threadID, userID)
	if err != nil {
		return err
	}

	// Fetch user info for broadcast
	user, _ := s.repo.GetUserByID(userID)
	thread, _ := s.repo.GetThreadByID(threadID)

	_ = websocket.SendThreadParticipantAdded(s.hub, websocket.ThreadParticipantPayload{
		ThreadID: threadID.String(),
		UserID:   userID.String(),
		UserName: user.FullName,
		//Avatar:   user.Avatar, // optional
		JoinedAt: time.Now().Format(time.RFC3339),
		CaseID:   thread.CaseID.String(),
	})

	return nil
}

func (s *Annotationthreadservice) GetThreadParticipants(threadID uuid.UUID) ([]ThreadParticipant, error) {
	return s.repo.GetThreadParticipants(threadID)
}

func (s *Annotationthreadservice) UpdateThreadStatus(threadID uuid.UUID, status ThreadStatus, updatedBy uuid.UUID) error {
	if !isLeadInvestigator(updatedBy) {
		return errors.New("only lead investigators can update thread status")
	}
	err := s.repo.UpdateThreadStatus(threadID, status)
	if err != nil {
		return err
	}

	// Broadcast update
	thread, _ := s.repo.GetThreadByID(threadID)
	_ = websocket.SendThreadEvent(s.hub, websocket.ThreadEventPayload{
		ThreadID:  threadID.String(),
		CaseID:    thread.CaseID.String(),
		UpdatedBy: updatedBy.String(),
		NewStatus: string(status),
	}, "thread_status_updated")

	return nil

}

func (s *Annotationthreadservice) UpdateThreadPriority(threadID uuid.UUID, priority ThreadPriority, updatedBy uuid.UUID) error {
	if !isLeadInvestigator(updatedBy) {
		return errors.New("only lead investigators can update thread priority")
	}
	err := s.repo.DB.Model(&AnnotationThread{}).Where("id = ?", threadID).Update("priority", priority).Error
	if err != nil {
		return err
	}

	thread, _ := s.repo.GetThreadByID(threadID)
	_ = websocket.SendThreadEvent(s.hub, websocket.ThreadEventPayload{
		ThreadID:    threadID.String(),
		CaseID:      thread.CaseID.String(),
		UpdatedBy:   updatedBy.String(),
		NewPriority: string(priority),
	}, "thread_priority_updated")

	return nil
}

// GetThreadByID retrieves a thread by its ID.
func (s *Annotationthreadservice) GetThreadByID(threadID uuid.UUID) (*AnnotationThread, error) {
	return s.repo.GetThreadByID(threadID)
}

// GetUserByID retrieves a user by ID.
func (s *Annotationthreadservice) GetUserByID(userID uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(userID)
}

// CreateThreadWithParticipant adds a participant to an existing thread with tags.
func (s *Annotationthreadservice) CreateThreadWithParticipant(thread *AnnotationThread, tags []string, participantID uuid.UUID) error {
	// Optionally update tags if needed
	if len(tags) > 0 {
		if err := s.repo.UpdateThreadTags(thread.ID, tags); err != nil {
			return err
		}
	}
	// Add participant
	if err := s.AddParticipant(thread.ID, participantID); err != nil {
		return err
	}
	return nil
}

// UpdateThreadTags updates the tags for a given thread.
func (s *Annotationthreadservice) UpdateThreadTags(threadID uuid.UUID, tags []string) error {
	return s.repo.UpdateThreadTags(threadID, tags)
}

// Placeholder for actual role verification
func isLeadInvestigator(userID uuid.UUID) bool {
	// TODO: Replace with real RBAC/role-check logic
	return true
}
