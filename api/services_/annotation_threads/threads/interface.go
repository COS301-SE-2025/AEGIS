package annotationthreads

import (
	"github.com/google/uuid"
)

type AnnotationThreadService interface {
	// CreateThread creates a new annotation thread.
	CreateThread(caseID, fileID, userID uuid.UUID, title string, tags []string, priority ThreadPriority) (*AnnotationThread, error)
	// GetThread retrieves an annotation thread by its FileID.
	GetThreadsByFile(fileID uuid.UUID) ([]AnnotationThread, error)
	// GetThread retrieves an annotation threads in a case by its CaseID.
	GetThreadsByCase(caseID uuid.UUID) ([]AnnotationThread, error)
	//UpdateThreadStatus updates the status of an annotation thread.
	UpdateThreadStatus(threadID uuid.UUID, status ThreadStatus, updatedBy uuid.UUID) error
	// UpdateThreadPriority updates the priority of an annotation thread.
	UpdateThreadPriority(threadID uuid.UUID, priority ThreadPriority, updatedBy uuid.UUID) error
	//AddParticipant adds a participant to an annotation thread.
	AddParticipant(threadID, userID uuid.UUID) error
	//GetThreadParticipants retrieves all participants of an annotation thread.
	GetThreadParticipants(threadID uuid.UUID) ([]ThreadParticipant, error)
	// GetThreadByID retrieves an annotation thread by its ID.
	GetThreadByID(threadID uuid.UUID) (*AnnotationThread, error)
	// GetUserByID retrieves threads by user ID (participant or creator).
	GetUserByID(userID uuid.UUID) (*User, error)
	CreateThreadWithParticipant(thread *AnnotationThread, tags []string, userID uuid.UUID) error
	UpdateThreadTags(threadID uuid.UUID, tags []string) error
}
