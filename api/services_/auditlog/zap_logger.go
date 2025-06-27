package auditlog

import (
	"go.uber.org/zap"
)

// ZapLogger is a structured logger that logs audit events to the console
// or file using Uber's high-performance zap logging library.
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger creates and returns a new ZapLogger instance.
// zap.NewProduction() is used for structured, production-ready logs.
// Use zap.NewDevelopment() if more human-readable output is needed during development.
func NewZapLogger() *ZapLogger {
	logger, _ := zap.NewProduction() // swap with NewDevelopment() for dev logs
	return &ZapLogger{logger: logger}
}
func NewZapLoggerWithInstance(l *zap.Logger) *ZapLogger {
	return &ZapLogger{logger: l}
}

// Log writes a structured audit log to the configured zap output (stdout, file, etc.).
// This is useful for observability, debugging, or compliance tracking.
func (z *ZapLogger) Log(log AuditLog) {
	z.logger.Info("Audit Event",
		// General info
		zap.String("id", log.ID),
		zap.String("timestamp", log.Timestamp.Format("2006-01-02T15:04:05Z07:00")),
		zap.String("action", log.Action),

		// Actor metadata
		zap.String("actor.id", log.Actor.ID),
		zap.String("actor.role", log.Actor.Role),
		zap.String("actor.ip", log.Actor.IPAddress),
		zap.String("actor.ua", log.Actor.UserAgent),

		// Target details
		zap.String("target.type", log.Target.Type),
		zap.String("target.id", log.Target.ID),
		zap.Any("target.extra", log.Target.AdditionalInfo),

		// Contextual details
		zap.String("service", log.Service),
		zap.String("status", log.Status),
		zap.String("description", log.Description),
		zap.Any("metadata", log.Metadata),
	)
}
