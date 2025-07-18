// models.go
package chat

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatGroup represents a chat group
type ChatGroup struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name        string             `bson:"name" json:"name"`
	Description string             `bson:"description" json:"description"`
	Type        string             `bson:"type" json:"type"` // "private", "group", "channel"
	Members     []*Member          `bson:"members" json:"members"`
	CreatedBy   string             `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at" json:"updated_at"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	LastMessage *LastMessage       `bson:"last_message,omitempty" json:"last_message,omitempty"`
	Settings    *GroupSettings     `bson:"settings,omitempty" json:"settings,omitempty"`
	CaseID      string             `bson:"case_id" json:"case_id"`
	//Avatar      string             `bson:"-" json:"avatar,omitempty"`
}

// Member represents a group member
type Member struct {
	UserEmail   string    `bson:"user_email" json:"user_email"`
	Role        string    `bson:"role" json:"role"` // "admin", "member"
	JoinedAt    time.Time `bson:"joined_at" json:"joined_at"`
	IsActive    bool      `bson:"is_active" json:"is_active"`
	Permissions []string  `bson:"permissions,omitempty" json:"permissions,omitempty"`
}

// GroupSettings represents group configuration
type GroupSettings struct {
	IsPublic          bool     `bson:"is_public" json:"is_public"`
	AllowInvites      bool     `bson:"allow_invites" json:"allow_invites"`
	MuteNotifications bool     `bson:"mute_notifications" json:"mute_notifications"`
	AllowedFileTypes  []string `bson:"allowed_file_types,omitempty" json:"allowed_file_types,omitempty"`
	MaxFileSize       int64    `bson:"max_file_size,omitempty" json:"max_file_size,omitempty"`
}

// LastMessage represents the last message in a group
type LastMessage struct {
	MessageID   primitive.ObjectID `bson:"message_id" json:"message_id"`
	Content     string             `bson:"content" json:"content"`
	SenderEmail string             `bson:"sender_email" json:"sender_email"`
	SenderName  string             `bson:"sender_name" json:"sender_name"`
	Timestamp   time.Time          `bson:"timestamp" json:"timestamp"`
	MessageType string             `bson:"message_type" json:"message_type"`
}

// Message represents a chat message
type Message struct {
	ID            string                 `bson:"_id,omitempty" json:"id"`
	GroupID       primitive.ObjectID     `bson:"group_id" json:"group_id"`
	SenderEmail   string                 `bson:"sender_email" json:"sender_email"`
	SenderName    string                 `bson:"sender_name" json:"sender_name"`
	Content       string                 `bson:"content" json:"content"`
	MessageType   string                 `bson:"message_type" json:"message_type"` // "text", "image", "file", "system"
	Attachments   []*Attachment          `bson:"attachments,omitempty" json:"attachments,omitempty"`
	ReplyTo       *primitive.ObjectID    `bson:"reply_to,omitempty" json:"reply_to,omitempty"`
	Mentions      []string               `bson:"mentions,omitempty" json:"mentions,omitempty"`
	Status        MessageStatus          `bson:"status" json:"status"`
	CreatedAt     time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt     time.Time              `bson:"updated_at" json:"updated_at"`
	IsDeleted     bool                   `bson:"is_deleted" json:"is_deleted"`
	Metadata      map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
	AttachmentURL string                 `bson:"attachment_url,omitempty" json:"attachment_url,omitempty"`
}

// MessageStatus represents message delivery and read status
type MessageStatus struct {
	Sent      time.Time      `bson:"sent" json:"sent"`
	Delivered *time.Time     `bson:"delivered,omitempty" json:"delivered,omitempty"`
	Edited    *time.Time     `bson:"edited,omitempty" json:"edited,omitempty"`
	ReadBy    []*ReadReceipt `bson:"read_by,omitempty" json:"read_by,omitempty"`
}

// ReadReceipt represents when a user read a message
type ReadReceipt struct {
	UserEmail string    `bson:"user_email" json:"user_email"`
	ReadAt    time.Time `bson:"read_at" json:"read_at"`
}

// Attachment represents a file attachment
type Attachment struct {
	ID       string                 `bson:"id" json:"id"`
	FileName string                 `bson:"file_name" json:"file_name"`
	FileType string                 `bson:"file_type" json:"file_type"`
	FileSize int64                  `bson:"file_size" json:"file_size"`
	URL      string                 `bson:"url" json:"url"`
	Hash     string                 `bson:"hash,omitempty" json:"hash,omitempty"` // IPFS hash
	Metadata map[string]interface{} `bson:"metadata,omitempty" json:"metadata,omitempty"`
}

// IPFSUploadResult represents the result of an IPFS upload
type IPFSUploadResult struct {
	Hash     string `json:"hash"`
	URL      string `json:"url"`
	Size     int64  `json:"size"`
	FileName string `json:"file_name"`
}

// User represents a user (simplified structure)
type User struct {
	ID       string     `bson:"_id,omitempty" json:"id"`
	Email    string     `bson:"email" json:"email"`
	FullName string     `bson:"full_name" json:"full_name"`
	Role     string     `bson:"role" json:"role"`
	Avatar   string     `bson:"avatar,omitempty" json:"avatar,omitempty"`
	Status   string     `bson:"status,omitempty" json:"status,omitempty"` // "online", "offline", "away"
	LastSeen *time.Time `bson:"last_seen,omitempty" json:"last_seen,omitempty"`
}

// EventType represents different types of real-time events
type EventType string

const (
	EventNewMessage   EventType = "new_message"
	EventMessageRead  EventType = "message_read"
	EventUserJoined   EventType = "user_joined"
	EventUserLeft     EventType = "user_left"
	EventGroupUpdated EventType = "group_updated"
	EventGroupDeleted EventType = "group_deleted"
	EventTypingStart  EventType = "typing_start"
	EventTypingStop   EventType = "typing_stop"
	EventUserOnline   EventType = "user_online"
	EventUserOffline  EventType = "user_offline"
)

// WebSocketEvent represents a real-time event
type WebSocketEvent struct {
	Type      EventType   `json:"type"`
	GroupID   string      `json:"group_id,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp int64       `json:"timestamp"`
	UserEmail string      `json:"user_email,omitempty"`
}

// TypingEvent represents a typing indicator event
type TypingEvent struct {
	UserEmail string `json:"user_email"`
	UserName  string `json:"user_name"`
	GroupID   string `json:"group_id"`
	IsTyping  bool   `json:"is_typing"`
}

// MessageReadEvent represents a message read event
type MessageReadEvent struct {
	MessageIDs []string `json:"message_ids"`
	GroupID    string   `json:"group_id"`
	UserEmail  string   `json:"user_email"`
	ReadAt     int64    `json:"read_at"`
}

// UserPresenceEvent represents user online/offline status
type UserPresenceEvent struct {
	UserEmail string `json:"user_email"`
	Status    string `json:"status"` // "online", "offline", "away"
	LastSeen  *int64 `json:"last_seen,omitempty"`
}
