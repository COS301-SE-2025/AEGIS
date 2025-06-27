package websocket

import "encoding/json"

const (
	ThreadCreatedType          = "THREAD_CREATED"
	ThreadResolvedType         = "THREAD_RESOLVED"
	ThreadParticipantAddedType = "THREAD_PARTICIPANT_ADDED"
)

type ThreadEventPayload struct {
	ThreadID    string `json:"thread_id"`
	Title       string `json:"title,omitempty"`
	UpdatedBy   string `json:"updated_by,omitempty"`
	NewStatus   string `json:"new_status,omitempty"`
	NewPriority string `json:"new_priority,omitempty"`
	CaseID      string `json:"case_id"`
	CreatedAt   string `json:"created_at,omitempty"`
}

type ThreadParticipantPayload struct {
	ThreadID string `json:"thread_id"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Avatar   string `json:"avatar,omitempty"`
	JoinedAt string `json:"joined_at"`
	CaseID   string `json:"case_id"`
}

type ThreadCreatedPayload struct {
	ThreadID  string `json:"thread_id"`
	Title     string `json:"title"`
	CaseID    string `json:"case_id"`
	FileID    string `json:"file_id"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	Priority  string `json:"priority"`
}

// Use this to broadcast thread updates
func SendThreadEvent(hub *Hub, payload ThreadEventPayload, eventType string) error {
	msg := WebSocketMessage{
		Type:    eventType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	hub.broadcast <- MessageEnvelope{
		CaseID: payload.CaseID,
		Data:   data,
	}
	return nil
}

func SendThreadParticipantAdded(hub *Hub, payload ThreadParticipantPayload) error {
	msg := WebSocketMessage{
		Type:    ThreadParticipantAddedType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	hub.broadcast <- MessageEnvelope{
		CaseID: payload.CaseID,
		Data:   data,
	}
	return nil
}

func SendThreadCreated(hub *Hub, payload ThreadCreatedPayload) error {
	msg := WebSocketMessage{
		Type:    ThreadCreatedType,
		Payload: payload,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	hub.broadcast <- MessageEnvelope{
		CaseID: payload.CaseID,
		Data:   data,
	}
	return nil
}
