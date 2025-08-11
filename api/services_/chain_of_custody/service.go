// coc/service.go
package coc

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"strings"
	"sync"
	"time"
)

type Authorizer interface {
	CanLogCoC(ctx context.Context, actorID *string, caseID, evidenceID string, action Action) bool
}
type Auditor interface {
	Log(ctx context.Context, typ string, fields map[string]any)
}

type Repo interface {
	Insert(ctx context.Context, p LogParams) (string, error)
	ListByEvidence(ctx context.Context, evidenceID string, f ListFilters) ([]Entry, error)
}

type Service struct {
	Repo      Repo
	Authz     Authorizer
	Audit     Auditor
	DedupeWin time.Duration // e.g., 3s for "view", 0 to disable

	mu       sync.Mutex
	lastSeen map[string]time.Time // key: evidenceID|actorID|action
}

// Log logs a Chain of Custody event (insert into DB and log in Mongo/Zap)
func (s *Service) Log(ctx context.Context, p LogParams) (string, error) {
	// Default to the current time if OccurredAt is zero
	if p.OccurredAt.IsZero() {
		p.OccurredAt = time.Now()
	}

	// RBAC authorization check (if implemented)
	if s.Authz != nil && !s.Authz.CanLogCoC(ctx, p.ActorID, p.CaseID, p.EvidenceID, p.Action) {
		// Log failure event with audit logger
		s.audit(ctx, "CHAIN_OF_CUSTODY_LOG", map[string]any{
			"status": "DENY", "reason": "rbac_forbidden",
			"caseId": p.CaseID, "evidenceId": p.EvidenceID, "actorId": p.ActorID, "action": p.Action,
		})
		return "", errors.New("forbidden")
	}

	// Normalize hash values (lowercase)
	if p.HashMD5 != nil {
		v := strings.ToLower(*p.HashMD5)
		p.HashMD5 = &v
	}
	if p.HashSHA1 != nil {
		v := strings.ToLower(*p.HashSHA1)
		p.HashSHA1 = &v
	}
	if p.HashSHA256 != nil {
		v := strings.ToLower(*p.HashSHA256)
		p.HashSHA256 = &v
	}

	// Deduplicate 'view' actions (optional)
	if s.DedupeWin > 0 && p.Action == ActionView {
		if s.lastSeen == nil {
			s.lastSeen = make(map[string]time.Time)
		}
		key := p.EvidenceID + "|" + deref(p.ActorID) + "|" + string(p.Action)
		s.mu.Lock()
		if t, ok := s.lastSeen[key]; ok && p.OccurredAt.Sub(t) <= s.DedupeWin {
			s.mu.Unlock()
			// Skip and log the duplication event
			s.audit(ctx, "CHAIN_OF_CUSTODY_LOG", map[string]any{
				"status": "SKIP_DUP", "caseId": p.CaseID, "evidenceId": p.EvidenceID,
				"actorId": p.ActorID, "action": p.Action, "occurredAt": p.OccurredAt,
			})
			return "", nil
		}
		s.lastSeen[key] = p.OccurredAt
		s.mu.Unlock()
	}

	// Insert into DB
	id, err := s.Repo.Insert(ctx, p)
	if err != nil {
		// Log failure event
		s.audit(ctx, "CHAIN_OF_CUSTODY_LOG", map[string]any{
			"status": "FAILED", "error": err.Error(),
			"caseId": p.CaseID, "evidenceId": p.EvidenceID, "actorId": p.ActorID, "action": p.Action,
			"hashMD5": p.HashMD5, "hashSHA1": p.HashSHA1, "hashSHA256": p.HashSHA256, "occurredAt": p.OccurredAt,
		})
		return "", err
	}

	// Log success event
	s.audit(ctx, "CHAIN_OF_CUSTODY_LOG", map[string]any{
		"status": "SUCCESS", "id": id,
		"caseId": p.CaseID, "evidenceId": p.EvidenceID, "actorId": p.ActorID, "action": p.Action,
		"hashMD5": p.HashMD5, "hashSHA1": p.HashSHA1, "hashSHA256": p.HashSHA256, "occurredAt": p.OccurredAt,
	})

	return id, nil
}

// audit logs an event to both MongoDB and Zap

// audit logs an event to both MongoDB and Zap
func (s *Service) audit(ctx context.Context, typ string, fields map[string]any) {
	// Convert fields map to a map[string]string for Metadata
	metadata := make(map[string]string)
	for key, value := range fields {
		if strVal, ok := value.(string); ok {
			metadata[key] = strVal
		}
	}

	// Directly log the event using the AuditLogger
	if s.Audit != nil {
		// AuditLogger Log method expects context, action type, and fields as map[string]any
		s.Audit.Log(ctx, typ, fields) // Call the AuditLogger's Log method
	}
}
func deref(p *string) string {
	if p != nil {
		return *p
	}
	return ""
}

func (s *Service) ListByEvidence(ctx context.Context, evidenceID string, f ListFilters) ([]Entry, error) {
	return s.Repo.ListByEvidence(ctx, evidenceID, f)
}

func (s *Service) ToCSV(entries []Entry) ([]byte, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)

	// Header
	_ = writer.Write([]string{
		"ID", "CaseID", "EvidenceID", "ActorID", "Action",
		"Reason", "Location", "MD5", "SHA1", "SHA256", "OccurredAt", "CreatedAt",
	})

	// Rows
	for _, e := range entries {
		_ = writer.Write([]string{
			e.ID, e.CaseID, e.EvidenceID, deref(e.ActorID), string(e.Action),
			deref(e.Reason), deref(e.Location), deref(e.HashMD5), deref(e.HashSHA1), deref(e.HashSHA256),
			e.OccurredAt.Format(time.RFC3339), e.CreatedAt.Format(time.RFC3339),
		})
	}
	writer.Flush()
	return buf.Bytes(), writer.Error()
}
