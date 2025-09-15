package websocket

import (
	"aegis-api/pkg/chatModels"
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	NewMessageType      = "NEW_MESSAGE"
	ThreadUpdatedType   = "THREAD_UPDATED"
	ReactionUpdatedType = "REACTION_UPDATED"
	MessageApprovedType = "MESSAGE_APPROVED"
)

type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type MessageEnvelope struct {
	CaseID string
	Data   []byte
}

var MessageCollection *mongo.Collection // Inject this from main.go or init

func SendJSONMessage(hub *Hub, caseID string, messageType string, payload interface{}) error {
	message := WebSocketMessage{
		Type:    messageType,
		Payload: payload,
	}
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}
	hub.broadcast <- MessageEnvelope{
		CaseID: caseID,
		Data:   data,
	}
	return nil
}

func SaveMessageToDB(payload chatModels.NewMessagePayload) error {
	parsedTime, err := time.Parse(time.RFC3339, payload.Timestamp)
	if err != nil {
		return err
	}

	groupID := primitive.NewObjectID()
	if payload.GroupID != "" {
		var err error
		// Convert the GroupID string to a primitive ObjectID
		groupID, err = primitive.ObjectIDFromHex(payload.GroupID)
		if err != nil {
			return fmt.Errorf("invalid group ID: %w", err)
		}
	}

	message := chatModels.Message{
		ID:          payload.MessageID,
		GroupID:     groupID,
		SenderEmail: payload.SenderName,
		SenderName:  payload.SenderName,
		Content:     payload.Text,
		CreatedAt:   parsedTime,
		UpdatedAt:   parsedTime,
		IsDeleted:   false,
		Status: chatModels.MessageStatus{
			Sent: parsedTime,
		},
		MessageType: "text",
	}

	if len(payload.Attachments) > 0 {
		message.Attachments = payload.Attachments
		message.MessageType = "file"
	}

	_, err = MessageCollection.InsertOne(context.Background(), message)
	return err
}
