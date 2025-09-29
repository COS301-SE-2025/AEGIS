package websocket

import (
	"aegis-api/pkg/chatModels"
	"context"
	"encoding/json"
	"fmt"
	"log"
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
	// Guard
	if MessageCollection == nil {
		log.Println("[WS] MessageCollection is nil; skipping DB persist")
		return nil
	}

	// Timestamp
	created, err := time.Parse(time.RFC3339, payload.Timestamp)
	if err != nil {
		log.Printf("[WS] invalid timestamp %q: %v; using time.Now()", payload.Timestamp, err)
		created = time.Now().UTC()
	}

	// GroupID
	var groupID primitive.ObjectID
	if payload.GroupID != "" {
		if gid, err := primitive.ObjectIDFromHex(payload.GroupID); err == nil {
			groupID = gid
		} else {
			log.Printf("[WS] invalid group ID %q: %v; generating new ObjectID()", payload.GroupID, err)
			groupID = primitive.NewObjectID()
		}
	} else {
		groupID = primitive.NewObjectID()
	}

	// Derive message type
	msgType := "text"
	if len(payload.Attachments) > 0 {
		msgType = "file"
	}

	// Build doc — ciphertext path vs plaintext path
	doc := chatModels.Message{
		ID:          payload.MessageID, // string _id is OK in Mongo
		GroupID:     groupID,
		SenderEmail: payload.SenderEmail, // ✅ use SenderEmail (not SenderID/SenderName)
		SenderName:  payload.SenderName,
		MessageType: msgType,
		IsEncrypted: payload.IsEncrypted,
		Envelope:    payload.Envelope, // may be nil if plaintext

		// plaintext content only when not encrypted
		Content:     "",
		Status:      chatModels.MessageStatus{Sent: created},
		CreatedAt:   created,
		UpdatedAt:   created,
		IsDeleted:   false,
		Attachments: payload.Attachments, // if you extended with envelope, it passes through
	}

	if !payload.IsEncrypted {
		doc.Content = payload.Text
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if _, err := MessageCollection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	return nil
}
