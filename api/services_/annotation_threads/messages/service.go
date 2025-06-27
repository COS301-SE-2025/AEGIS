package messages

import (
	"aegis-api/pkg/websocket"
	"errors"
	"time"

	"github.com/google/uuid"
)

type MessageServiceImpl struct {
	repo MessageRepository
	hub  *websocket.Hub
}

func NewMessageService(repo MessageRepository, hub *websocket.Hub) *MessageServiceImpl {
	return &MessageServiceImpl{repo: repo, hub: hub}

}

func (s *MessageServiceImpl) SendMessage(threadID, userID uuid.UUID, message string, parentMessageID *uuid.UUID, mentions []uuid.UUID) (*ThreadMessage, error) {
	// Validate inputs
	if message == "" {
		return nil, errors.New("message cannot be empty")
	}

	// Create the message model
	msg := &ThreadMessage{
		ID:              uuid.New(),
		ThreadID:        threadID,
		ParentMessageID: parentMessageID,
		UserID:          userID,
		Message:         message,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save the message to the database
	err := s.repo.CreateMessage(msg)
	if err != nil {
		return nil, err
	}

	// Add mentions separately
	if len(mentions) > 0 {
		err = s.repo.AddMentions(msg.ID, mentions)
		if err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (s *MessageServiceImpl) GetMessagesByThread(threadID uuid.UUID) ([]ThreadMessage, error) {
	return s.repo.GetMessagesByThread(threadID)
}

func (s *MessageServiceImpl) ApproveMessage(messageID, approverID uuid.UUID) error {
	err := s.repo.ApproveMessage(messageID, approverID)
	if err != nil {
		return err
	}

	// Fetch updated message for broadcasting
	msg, err := s.repo.GetMessageByID(messageID)
	if err != nil {
		return nil // or log error but do not block approval flow
	}

	// Broadcast approval event
	_ = websocket.SendJSONMessage(s.hub, msg.ThreadID.String(), websocket.MessageApprovedType, msg)

	return nil
}

func (s *MessageServiceImpl) AddReaction(messageID, userID uuid.UUID, reaction string) error {
	err := s.repo.AddReaction(messageID, userID, reaction)
	if err != nil {
		return err
	}

	// Broadcast reaction update
	msg, err := s.repo.GetMessageByID(messageID)
	if err == nil {
		_ = websocket.SendJSONMessage(s.hub, msg.ThreadID.String(), websocket.ReactionUpdatedType, msg)
	}

	return nil
}

func (s *MessageServiceImpl) RemoveReaction(messageID, userID uuid.UUID) error {
	err := s.repo.RemoveReaction(messageID, userID)
	if err != nil {
		return err
	}

	// Broadcast reaction update
	msg, err := s.repo.GetMessageByID(messageID)
	if err == nil {
		_ = websocket.SendJSONMessage(s.hub, msg.ThreadID.String(), websocket.ReactionUpdatedType, msg)
	}

	return nil
}

func (s *MessageServiceImpl) GetReplies(parentMessageID uuid.UUID) ([]ThreadMessage, error) {
	return s.repo.GetReplies(parentMessageID)
}

func (s *MessageServiceImpl) AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error {
	if len(mentions) == 0 {
		return nil // No mentions to add
	}

	return s.repo.AddMentions(messageID, mentions)
}

func (s *MessageServiceImpl) GetMessageByID(messageID uuid.UUID) (*ThreadMessage, error) {
	return s.repo.GetMessageByID(messageID)
}
