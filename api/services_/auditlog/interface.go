package auditlog

import (
	"context"

	"github.com/gin-gonic/gin"
)

type MongoLoggerInterface interface {
	Log(ctx *gin.Context, log AuditLog) error
}

type ZapLoggerInterface interface {
	Log(log AuditLog)
}
type AuditLogReader interface {
	GetRecentUserActivities(ctx context.Context, userID string) ([]AuditLog, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
}
