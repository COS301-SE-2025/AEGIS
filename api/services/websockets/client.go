package websockets

import (
    "log"
    "github.com/gorilla/websocket"
)

type Client struct {
    UserID   string
    CaseIDs  map[string]bool // e.g., { "case-1": true }
    Conn     *websocket.Conn
    Send     chan []byte
    Hub      *Hub
}

func (c *Client) ReadPump() {
    defer func() {
        c.Hub.Unregister <- c
        c.Conn.Close()
    }()

    for {
        _, message, err := c.Conn.ReadMessage()
        if err != nil {
            log.Println("read error:", err)
            break
        }

        c.Hub.Incoming <- Message{
            SenderID: c.UserID,
            Raw:      message,
        }
    }
}

func (c *Client) WritePump() {
    defer c.Conn.Close()
    for msg := range c.Send {
        if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
            break
        }
    }
}
