package messages

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MessageRepository struct {
	DB *gorm.DB
}

func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{DB: db}
}

func (r *MessageRepository) CreateMessage(msg *ThreadMessage) error {
	return r.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(msg).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *MessageRepository) GetMessagesByThread(threadID uuid.UUID) ([]ThreadMessage, error) {
	var messages []ThreadMessage
	err := r.DB.Preload("Mentions").Preload("Reactions").Where("thread_id = ?", threadID).Order("created_at asc").Find(&messages).Error
	return messages, err
}

func (r *MessageRepository) ApproveMessage(messageID, approverID uuid.UUID) error {
	now := time.Now()
	approved := true
	return r.DB.Model(&ThreadMessage{}).
		Where("id = ?", messageID).
		Updates(map[string]interface{}{
			"is_approved": approved,
			"approved_by": approverID,
			"approved_at": now,
		}).Error
}

func (r *MessageRepository) AddReaction(messageID, userID uuid.UUID, reaction string) error {
	react := MessageReaction{
		ID:        uuid.New(),
		MessageID: messageID,
		UserID:    userID,
		Reaction:  reaction,
		CreatedAt: time.Now(),
	}
	return r.DB.Create(&react).Error
}

func (r *MessageRepository) RemoveReaction(messageID, userID uuid.UUID) error {
	return r.DB.Where("message_id = ? AND user_id = ?", messageID, userID).Delete(&MessageReaction{}).Error
}

func (r *MessageRepository) GetReplies(parentMessageID uuid.UUID) ([]ThreadMessage, error) {
	var replies []ThreadMessage
	err := r.DB.Preload("Mentions").Preload("Reactions").Where("parent_message_id = ?", parentMessageID).Order("created_at asc").Find(&replies).Error
	return replies, err
}

func (r *MessageRepository) GetMessageByID(messageID uuid.UUID) (*ThreadMessage, error) {
	var msg ThreadMessage
	err := r.DB.Where("id = ?", messageID).First(&msg).Error
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepository) AddMentions(messageID uuid.UUID, mentions []uuid.UUID) error {
	if len(mentions) == 0 {
		return nil
	}

	type MessageMention struct {
		MessageID       uuid.UUID `gorm:"column:message_id"`
		MentionedUserID uuid.UUID `gorm:"column:mentioned_user_id"`
		CreatedAt       time.Time `gorm:"column:created_at"`
	}

	var mentionRows []MessageMention
	now := time.Now()
	for _, userID := range mentions {
		mentionRows = append(mentionRows, MessageMention{
			MessageID:       messageID,
			MentionedUserID: userID,
			CreatedAt:       now,
		})
	}

	return r.DB.Create(&mentionRows).Error
}
