package auditlog

import (
	"github.com/gin-gonic/gin"
)

type AuditLogger struct {
	mongo MongoLoggerInterface
	zap   ZapLoggerInterface
}

func NewAuditLogger(mongo MongoLoggerInterface, zap ZapLoggerInterface) *AuditLogger {
	return &AuditLogger{mongo: mongo, zap: zap}
}
func (a *AuditLogger) Log(ctx *gin.Context, log AuditLog) error {
	// Mongo log (persistent)
	if err := a.mongo.Log(ctx, log); err != nil {
		return err
	}

	// Zap log (console)
	a.zap.Log(log)

	return nil
}
