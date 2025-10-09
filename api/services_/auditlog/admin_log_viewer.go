package auditlog

import (
	"context"
	"sort"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Add to auditlog/interfaces.go or similar
type AdminAuditLogReader interface {
	GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, error)
}

// Add to auditlog/types.go or define inline
type AuditLogFilter struct {
	Status  string // "SUCCESS", "FAILED", "ALL" (default "ALL")
	Action  string // e.g., "EXTRACT_IOCS" for IOC retrievals
	Service string // e.g., "timelineai"
	Limit   int    // default 100, max 1000
	// Add more: DateFrom, DateTo, ActorID, etc.
}

// Update auditlog_service.go
// Add to AuditLogService struct (no change needed)

// Compile-time check
var _ AdminAuditLogReader = (*AuditLogService)(nil)

// GetAuditLogs retrieves audit logs across all collections with optional filters.
// For admin use; aggregates from all service-specific collections.
func (s *AuditLogService) GetAuditLogs(ctx context.Context, filter AuditLogFilter) ([]AuditLog, error) {
	if filter.Limit == 0 {
		filter.Limit = 100 // default
	}
	if filter.Limit > 1000 {
		filter.Limit = 1000 // sane max
	}

	collections := []string{
		"audit_logs_evidence",
		"audit_logs_case",
		"audit_logs_user",
		"audit_logs_report",
		"audit_logs_timelineai",
	}

	var logs []AuditLog

	// Get more logs from each collection to ensure we have enough to sort and limit properly
	tempLimit := filter.Limit * 2 // Get 2x the requested amount initially

	for _, collName := range collections {
		coll := s.db.Collection(collName)

		// Build filter
		query := bson.M{}
		if filter.Status != "ALL" {
			query["status"] = filter.Status
		}
		if filter.Action != "" {
			query["action"] = filter.Action
		}
		if filter.Service != "" {
			query["service"] = filter.Service
		}

		// Sort by timestamp descending - get more logs per collection
		opts := options.Find().
			SetSort(bson.D{{Key: "timestamp", Value: -1}}).
			SetLimit(int64(tempLimit)) // Get more logs initially

		cursor, err := coll.Find(ctx, query, opts)
		if err != nil {
			continue // Skip on error
		}

		var temp []AuditLog
		if err := cursor.All(ctx, &temp); err != nil {
			continue
		}

		logs = append(logs, temp...)
	}

	// Sort ALL logs by timestamp descending
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	// Now apply the actual limit to get the most recent logs across all collections
	if len(logs) > filter.Limit {
		logs = logs[:filter.Limit]
	}

	// Optional: Enrich emails like in GetRecentUserActivities
	for i, log := range logs {
		// Validate UUID format before querying
		if log.Actor.ID != "" {
			// Check if it's a valid UUID format
			if _, err := uuid.Parse(log.Actor.ID); err == nil {
				user, err := s.userRepo.GetByID(ctx, log.Actor.ID)
				if err == nil && user != nil {
					logs[i].Actor.Email = user.Email
				} else {
					logs[i].Actor.Email = "(unknown)"
				}
			} else {
				// Invalid UUID format
				logs[i].Actor.Email = "(invalid-id)"
			}
		} else {
			// Empty actor ID
			logs[i].Actor.Email = "(system)"
		}
	}

	return logs, nil
}
