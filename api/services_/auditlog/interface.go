package auditlog

import "github.com/gin-gonic/gin"

type MongoLoggerInterface interface {
	Log(ctx *gin.Context, log AuditLog) error
}

type ZapLoggerInterface interface {
	Log(log AuditLog)
}
