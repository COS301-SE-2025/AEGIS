package auditlog

import "time"

// AuditLog represents a structured audit trail entry
// that captures key user/system actions in the AEGIS system.

type Actor struct {
	ID        string `bson:"id"`
	Role      string `bson:"role"`
	UserAgent string `bson:"user_agent"`
	IPAddress string `bson:"ip_address"`
	Email     string `bson:"email,omitempty" json:"email,omitempty"`
}

type Target struct {
	Type           string            `bson:"type"`
	ID             string            `bson:"id"`
	AdditionalInfo map[string]string `bson:"additional_info"`
}

type AuditLog struct {
	ID          string // UUID
	Timestamp   time.Time
	Action      string // e.g., "UPLOAD_EVIDENCE"
	Actor       Actor  // contains user ID, role, IP, user agent
	Target      Target // resource affected: type, ID, extra info
	Service     string // service that triggered log, e.g., "chat"
	Status      string // e.g., "SUCCESS"
	Description string
	Metadata    map[string]string // route, method, etc.
}

// Inside auditlog/user_repository.go or similar

type User struct {
	ID    string
	Email string
}
