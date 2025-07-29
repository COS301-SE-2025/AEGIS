package chat

import (
	"context"
	"mime/multipart"
	"net/http"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ChatRepository defines the interface for chat data access
type ChatRepository interface {
	// Group operations
	CreateGroup(ctx context.Context, group *ChatGroup) error
	GetGroupByID(ctx context.Context, groupID primitive.ObjectID) (*ChatGroup, error)
	GetUserGroups(ctx context.Context, userEmail string) ([]*ChatGroup, error)
	UpdateGroup(ctx context.Context, group *ChatGroup) error
	DeleteGroup(ctx context.Context, groupID primitive.ObjectID) error
	AddMemberToGroup(ctx context.Context, groupID primitive.ObjectID, member *Member) error
	RemoveMemberFromGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) error
	IsUserInGroup(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error)
	UpdateLastMessage(ctx context.Context, groupID primitive.ObjectID, lastMessage *LastMessage) error
	GetGroupsByCaseID(ctx context.Context, caseID primitive.ObjectID) ([]*ChatGroup, error)
	UpdateGroupImage(ctx context.Context, groupID primitive.ObjectID, imageURL string) error

	// Message operations
	CreateMessage(ctx context.Context, message *Message) error
	GetMessageByID(ctx context.Context, messageID primitive.ObjectID) (*Message, error)
	GetMessages(ctx context.Context, groupID primitive.ObjectID, limit int, before *primitive.ObjectID) ([]*Message, error)
	SearchMessages(ctx context.Context, groupID primitive.ObjectID, query string, limit int, skip int) ([]*Message, error)
	UpdateMessage(ctx context.Context, message *Message) error
	DeleteMessage(ctx context.Context, messageID primitive.ObjectID) error
	MarkMessagesAsRead(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error
	GetUnreadCount(ctx context.Context, groupID primitive.ObjectID, userEmail string) (int, error)

	// Utility operations
	GetGroupMembers(ctx context.Context, groupID primitive.ObjectID) ([]*Member, error)
	IsGroupAdmin(ctx context.Context, groupID primitive.ObjectID, userEmail string) (bool, error)

	GetUndeliveredMessages(ctx context.Context, userEmail string, limit int, before *primitive.ObjectID) ([]*Message, error)
	MarkMessagesAsDelivered(ctx context.Context, groupID primitive.ObjectID, messageIDs []primitive.ObjectID, userEmail string) error
}

// IPFSUploader defines the interface for IPFS file operations
type IPFSUploader interface {
	UploadFile(ctx context.Context, file multipart.File, fileName string) (*IPFSUploadResult, error)
	UploadBytes(ctx context.Context, data []byte, fileName string) (*IPFSUploadResult, error)
	GetFileURL(hash string) string
	DeleteFile(ctx context.Context, hash string) error
}

// UserService defines the interface for user operations (assuming it exists)
type UserService interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUsersByEmails(ctx context.Context, emails []string) ([]*User, error)
	ValidateUserExists(ctx context.Context, email string) (bool, error)
}

// WebSocketManager defines the interface for real-time communication
type WebSocketManager interface {
	BroadcastToGroup(groupID string, message WebSocketMessage) error
	SendToUser(userEmail string, message interface{}) error
	AddUserToGroup(userEmail string, groupID, caseID string, conn *websocket.Conn) error
	RemoveUserFromGroup(userEmail, groupID string) error
	GetActiveUsers(groupID string) []string
	HandleConnection(wr http.ResponseWriter, r *http.Request) error
	BroadcastToCase(caseID string, message WebSocketMessage) error
	AddConnection(userID, caseID string, conn *websocket.Conn)
}
