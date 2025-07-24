package auditlog

import (
	"context"
	"sort"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// AuditLogService provides methods to retrieve logs for users.
type AuditLogService struct {
	db *mongo.Database
}

// NewAuditLogService creates a new instance of AuditLogService
func NewAuditLogService(db *mongo.Database) *AuditLogService {
	return &AuditLogService{db: db}
}

// Compile-time check
var _ AuditLogReader = (*AuditLogService)(nil)

// GetRecentUserActivities returns the 20 most recent successful logs for a user across relevant collections.
func (s *AuditLogService) GetRecentUserActivities(ctx context.Context, userID string) ([]AuditLog, error) {
	collections := []string{
		"audit_logs_evidence",
		"audit_logs_case",
		"audit_logs_user",
		"audit_logs_chat",
		"audit_logs_annotation_threads",
		"audit_logs_annotation_messages",
	}

	var logs []AuditLog
	for _, collName := range collections {
		coll := s.db.Collection(collName)

		// Build the query filter
		filter := bson.M{
			"actor.id": userID,
			"status":   "SUCCESS",
		}

		// Sort by timestamp descending and limit to 20 (per collection, will deduplicate/limit after merge)
		opts := options.Find().
			SetSort(bson.D{{Key: "timestamp", Value: -1}}).
			SetLimit(20)

		cursor, err := coll.Find(ctx, filter, opts)
		if err != nil {
			continue // skip collection on error
		}

		var temp []AuditLog
		if err := cursor.All(ctx, &temp); err != nil {
			continue
		}

		logs = append(logs, temp...)
	}

	// Optional: sort and deduplicate merged logs
	logs = sortAndLimitLogs(logs, 20)

	return logs, nil
}

// sortAndLimitLogs sorts logs by timestamp descending and returns top N
func sortAndLimitLogs(logs []AuditLog, limit int) []AuditLog {
	// Sort
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	// Limit
	if len(logs) > limit {
		return logs[:limit]
	}
	return logs
}
