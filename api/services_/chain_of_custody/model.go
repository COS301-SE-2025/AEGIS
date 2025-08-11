// coc/dto.go
package coc

import "time"

// Action enumerates Chain of Custody event types supported by AEGIS.
type Action string

const (
	ActionUpload   Action = "upload"
	ActionDownload Action = "download"
	ActionArchive  Action = "archive"
	ActionView     Action = "view"
)

func (a Action) String() string { return string(a) }

// ParseAction converts a raw string to Action (case-insensitive).
// Returns ok=false if the value is not one of the supported actions.
func ParseAction(s string) (a Action, ok bool) {
	switch Action(s) {
	case ActionUpload, ActionDownload, ActionArchive, ActionView:
		return Action(s), true
	default:
		return "", false
	}
}

// Entry is a read model returned by the service/handlers.
type Entry struct {
	ID         string    `json:"id"`
	CaseID     string    `json:"caseId"`
	EvidenceID string    `json:"evidenceId"`
	ActorID    *string   `json:"actorId,omitempty"`
	Action     Action    `json:"action"`
	Reason     *string   `json:"reason,omitempty"`
	Location   *string   `json:"location,omitempty"`
	HashMD5    *string   `json:"hashMd5,omitempty"`
	HashSHA1   *string   `json:"hashSha1,omitempty"`
	HashSHA256 *string   `json:"hashSha256,omitempty"`
	OccurredAt time.Time `json:"occurredAt"`
	CreatedAt  time.Time `json:"createdAt"`
}

// LogParams is the write model used to insert a new CoC entry.
type LogParams struct {
	CaseID     string    `json:"caseId"`
	EvidenceID string    `json:"evidenceId"`
	ActorID    *string   `json:"actorId,omitempty"`
	Action     Action    `json:"action"`
	Reason     *string   `json:"reason,omitempty"`
	Location   *string   `json:"location,omitempty"`
	HashMD5    *string   `json:"hashMd5,omitempty"`
	HashSHA1   *string   `json:"hashSha1,omitempty"`
	HashSHA256 *string   `json:"hashSha256,omitempty"`
	OccurredAt time.Time `json:"occurredAt"` // service will default to time.Now() if zero
}

// ListFilters controls ListByEvidence queries.
type ListFilters struct {
	Action  *Action    `json:"action,omitempty"`
	ActorID *string    `json:"actorId,omitempty"`
	Since   *time.Time `json:"since,omitempty"`
	Until   *time.Time `json:"until,omitempty"`
	Limit   int        `json:"limit,omitempty"`
	Offset  int        `json:"offset,omitempty"`
}
