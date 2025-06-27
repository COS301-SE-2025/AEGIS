package messages

import "github.com/google/uuid"

type MessageService interface {
	SendMessage(threadID, userID uuid.UUID, message string, parentMessageID *uuid.UUID, mentions []uuid.UUID) (*ThreadMessage, error)
	GetMessagesByThread(threadID uuid.UUID) ([]ThreadMessage, error)
	ApproveMessage(messageID, approverID uuid.UUID) error
	AddReaction(messageID, userID uuid.UUID, reaction string) error
	RemoveReaction(messageID, userID uuid.UUID) error
	GetReplies(parentMessageID uuid.UUID) ([]ThreadMessage, error)
	AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error
	GetMessageByID(messageID uuid.UUID) (*ThreadMessage, error)
}
