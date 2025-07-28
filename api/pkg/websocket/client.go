package websocket

import (
	"aegis-api/pkg/sharedws"
	"aegis-api/services_/chat"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait  = 10 * time.Second
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

type Client struct {
	*sharedws.Client // embeds shared fields
	Hub              *Hub
	Send             chan []byte
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})
	c.Conn.SetCloseHandler(func(code int, text string) error {
		log.Printf("🔌 Client %s closed connection: %d - %s", c.UserID, code, text)
		return nil
	})

	for {
		_, rawMsg, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("⚠️ Client %s disconnected unexpectedly: %v", c.UserID, err)
			}
			break
		}

		var msg chat.WebSocketEvent
		if err := json.Unmarshal(rawMsg, &msg); err != nil {
			log.Printf("❌ Invalid WebSocket message: %v", err)
			continue
		}

		switch msg.Type {
		case chat.EventNewMessage:
			data, err := json.Marshal(msg.Payload)
			if err != nil {
				log.Printf("❌ Failed to re-marshal payload: %v", err)
				continue
			}

			var payload chat.NewMessagePayload
			if err := json.Unmarshal(data, &payload); err != nil {
				log.Printf("❌ Failed to unmarshal NEW_MESSAGE payload: %v", err)
				continue
			}

			if err := chat.SaveMessageToDB(payload); err != nil {
				log.Printf("❌ Failed to save message to DB: %v", err)
				continue
			}

			c.Hub.broadcast <- MessageEnvelope{
				CaseID: c.CaseID,
				Data:   rawMsg, // original message
			}

		default:
			log.Printf("⚠️ Unsupported WebSocket event type: %s", msg.Type)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		//origin := r.Header.Get("Origin")
		// Example: only allow your frontend
		return true //origin == "http://localhost:5173" || origin == "https://yourdomain.com" || origin == "http://127.0.1:5173"
	},
}

func ServeWS(hub *Hub, upgrader websocket.Upgrader, w http.ResponseWriter, r *http.Request, userID, caseID string) {
	var conn *websocket.Conn
	var err error

	// Attempt connection upgrade
	conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Failed to upgrade WebSocket connection: %v", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}

	// Create client
	client := &Client{
		Client: &sharedws.Client{
			UserID: userID,
			CaseID: caseID,
			Conn:   conn,
		},
		Hub:  hub,
		Send: make(chan []byte, 256),
	}

	// Register client
	client.Hub.register <- client

	// Launch read/write pumps
	go client.WritePump()
	go client.ReadPump()
}
