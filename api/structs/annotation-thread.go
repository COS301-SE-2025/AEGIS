package structs

import "github.com/google/uuid"

// Thread Creation Request Structs
type CreateThreadRequest struct {
	FileID   uuid.UUID `json:"file_id" binding:"required"`
	Title    string    `json:"title" binding:"required,min=1,max=255"`
	Tags     []string  `json:"tags,omitempty"`
	Priority string    `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH CRITICAL"`
}

type UpdateThreadStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=OPEN IN_PROGRESS RESOLVED CLOSED"`
}

type UpdateThreadPriorityRequest struct {
	Priority string `json:"priority" binding:"required,oneof=LOW MEDIUM HIGH CRITICAL"` //come back
}

type AddParticipantRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// Message Request Structs
type SendMessageRequest struct {
	Message         string      `json:"message" binding:"required,min=1,max=5000"`
	ParentMessageID *uuid.UUID  `json:"parent_message_id,omitempty"`
	Mentions        []uuid.UUID `json:"mentions,omitempty"`
}

type AddReactionRequest struct {
	Reaction string `json:"reaction" binding:"required,min=1,max=10"`
}

type RemoveReactionRequest struct {
	Reaction string `json:"reaction" binding:"required,min=1,max=10"`
}

type AddMentionsRequest struct {
	Mentions []uuid.UUID `json:"mentions" binding:"required,min=1"`
}
