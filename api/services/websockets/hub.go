
package websockets

import (
	"context"
	"log"
	"time"

	"aegis-api/models/chat"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Hub struct {
	Clients     map[string]*Client             // userEmail -> client
	GroupClients map[string]map[string]*Client // groupID -> userEmail -> client
	Register    chan *Client
	Unregister  chan *Client
	Incoming    chan Message
}

type Message struct {
	SenderEmail string
	Raw         []byte
}

func NewHub() *Hub {
	return &Hub{
		Clients:     make(map[string]*Client),
		GroupClients: make(map[string]map[string]*Client),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Incoming:    make(chan Message),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.UserEmail] = client

			for groupID := range client.GroupIDs {
				if h.GroupClients[groupID] == nil {
					h.GroupClients[groupID] = make(map[string]*Client)
				}
				h.GroupClients[groupID][client.UserEmail] = client
			}

			// Fetch undelivered messages
			undelivered := fetchUndeliveredMessages(client.UserEmail)
			for _, msg := range undelivered {
				client.Send <- marshalToWebSocketEvent(msg)
				updateMessageAsDelivered(msg.ID, client.UserEmail, time.Now())
			}

		case client := <-h.Unregister:
			delete(h.Clients, client.UserEmail)
			for groupID := range client.GroupIDs {
				delete(h.GroupClients[groupID], client.UserEmail)
			}
			close(client.Send)

		case msg := <-h.Incoming:
			// Parse message format
			var chatMsg chat.Message
			if err := json.Unmarshal(msg.Raw, &chatMsg); err != nil {
				log.Println("invalid message format:", err)
				continue
			}

			h.saveMessage(chatMsg)
			h.deliverMessage(chatMsg)
		}
	}
}

func (h *Hub) deliverMessage(msg chat.Message) {
	for email, client := range h.GroupClients[msg.GroupID.Hex()] {
		if email == msg.SenderEmail {
			continue
		}
		if client != nil {
			client.Send <- marshalToWebSocketEvent(msg)
		} // else: offline, already handled by Register logic
	}
}

func (h *Hub) saveMessage(msg chat.Message) {
	msg.ID = primitive.NewObjectID()
	msg.CreatedAt = time.Now()
	msg.Status = chat.MessageStatus{Sent: time.Now()}
	msg.IsDeleted = false
	_, err := chatCollection.InsertOne(context.TODO(), msg)
	if err != nil {
		log.Println("failed to save message:", err)
	}
}

func fetchUndeliveredMessages(email string) []chat.Message {
	filter := bson.M{
		"status.read_by.user_email": bson.M{"$ne": email},
		"status.delivered":          nil,
		"is_deleted":                false,
	}

	cursor, err := chatCollection.Find(context.TODO(), filter)
	if err != nil {
		log.Println("fetchUndeliveredMessages error:", err)
		return nil
	}

	var messages []chat.Message
	if err := cursor.All(context.TODO(), &messages); err != nil {
		log.Println("cursor decode error:", err)
		return nil
	}
	return messages
}

func updateMessageAsDelivered(messageID primitive.ObjectID, email string, deliveredAt time.Time) {
	filter := bson.M{"_id": messageID}
	update := bson.M{
		"$set": bson.M{"status.delivered": deliveredAt},
		"$push": bson.M{"status.read_by": chat.ReadReceipt{
			UserEmail: email,
			ReadAt:    deliveredAt,
		}},
	}
	_, err := chatCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Println("failed to update delivered message:", err)
	}
}

func marshalToWebSocketEvent(msg chat.Message) []byte {
	event := chat.WebSocketEvent{
		Type:      chat.EventNewMessage,
		GroupID:   msg.GroupID.Hex(),
		Data:      msg,
		Timestamp: time.Now().Unix(),
		UserEmail: msg.SenderEmail,
	}
	data, _ := json.Marshal(event)
	return data
}
