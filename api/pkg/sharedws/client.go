package sharedws

import "github.com/gorilla/websocket"

// Client holds shared client fields that are accessed by both websocket and chat packages
type Client struct {
	UserID string
	CaseID string
	Conn   *websocket.Conn
	Send   chan []byte
}
