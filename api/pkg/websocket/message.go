package websocket

import "encoding/json"

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
