package x3dh

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuditLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Timestamp   time.Time          `bson:"timestamp"`
	Action      string             `bson:"action"`
	Actor       ActorInfo          `bson:"actor"`
	Target      TargetInfo         `bson:"target"`
	Service     string             `bson:"service"`
	Status      string             `bson:"status"`
	Description string             `bson:"description"`
	Metadata    bson.M             `bson:"metadata"` // generic key-value metadata
}

type ActorInfo struct {
	ID        string `bson:"id"`
	Role      string `bson:"role"`
	UserAgent string `bson:"user_agent"`
}

type TargetInfo struct {
	Type           string                 `bson:"type"`
	ID             string                 `bson:"id"`
	AdditionalInfo map[string]interface{} `bson:"additional_info"`
}

type MongoAuditLogger struct {
	collection *mongo.Collection
}

func NewMongoAuditLogger(db *mongo.Database) *MongoAuditLogger {
	return &MongoAuditLogger{
		collection: db.Collection("audit_logs_secure_chat"),
	}
}

func (l *MongoAuditLogger) Log(ctx context.Context, log AuditLog) error {
	log.Timestamp = time.Now().UTC()
	_, err := l.collection.InsertOne(ctx, log)
	return err
}
