package auditlog

import (
	"aegis-api/services_/auditlog"

	"github.com/gin-gonic/gin"
)

type MongoLoggerInterface interface {
	Log(ctx *gin.Context, log auditlog.AuditLog) error
}

type ZapLoggerInterface interface {
	Log(log auditlog.AuditLog)
}

type AuditLogger struct {
	mongo MongoLoggerInterface
	zap   ZapLoggerInterface
}

func NewAuditLogger(mongo MongoLoggerInterface, zap ZapLoggerInterface) *AuditLogger {
	return &AuditLogger{mongo: mongo, zap: zap}
}

func (a *AuditLogger) Log(ctx *gin.Context, log auditlog.AuditLog) error {
	if err := a.mongo.Log(ctx, log); err != nil {
		return err
	}
	a.zap.Log(log)
	return nil
}
