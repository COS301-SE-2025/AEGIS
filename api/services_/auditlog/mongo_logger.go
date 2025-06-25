package auditlog

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

// MongoLogger is responsible for storing audit logs in a MongoDB database.
type MongoLogger struct {
	db *mongo.Database
}

// NewMongoLogger initializes and returns a new instance of MongoLogger.
func NewMongoLogger(db *mongo.Database) *MongoLogger {
	return &MongoLogger{db: db}
}

// Log inserts a structured audit log entry into the appropriate MongoDB collection.
// It auto-fills missing fields such as UUID, timestamp, IP address, user agent, and request metadata.
func (l *MongoLogger) Log(ctx *gin.Context, log AuditLog) error {
	// Generate a unique ID and capture the current timestamp
	log.ID = uuid.NewString()
	log.Timestamp = time.Now().UTC()

	// Extract actor metadata from the request context
	log.Actor.IPAddress = ctx.ClientIP()
	log.Actor.UserAgent = ctx.Request.UserAgent()

	// Initialize metadata map if nil
	if log.Metadata == nil {
		log.Metadata = map[string]string{}
	}

	// Capture HTTP route and method for traceability
	log.Metadata["route"] = ctx.FullPath()
	log.Metadata["method"] = ctx.Request.Method

	// Select MongoDB collection based on the service that generated the log
	var collection *mongo.Collection
	switch log.Service {
	case "evidence":
		collection = l.db.Collection("audit_logs_evidence")
	case "case":
		collection = l.db.Collection("audit_logs_case")
	case "auth", "user":
		collection = l.db.Collection("audit_logs_user")
	case "admin":
		collection = l.db.Collection("audit_logs_admin")
	case "chat":
		collection = l.db.Collection("audit_logs_chat")
	default:
		// Fallback to general collection if service type is unrecognized
		collection = l.db.Collection("audit_logs_general")
	}

	// Insert the log entry into the selected collection
	_, err := collection.InsertOne(context.Background(), log)
	return err
}
