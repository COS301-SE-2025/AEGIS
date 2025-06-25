package auditlog_test

import (
	"testing"

	"aegis-api/services_/auditlog"

	"go.uber.org/zap/zaptest"
)

func TestZapLogger_Log(t *testing.T) {
	// Create a test-friendly zap logger
	// Use a helper constructor to inject it
	zapLogger := auditlog.NewZapLoggerWithInstance(zaptest.NewLogger(t))

	log := auditlog.AuditLog{
		ID:      "test-id",
		Action:  "LOGIN",
		Service: "auth",
		Actor: auditlog.Actor{
			ID:   "user123",
			Role: "admin",
		},
	}

	zapLogger.Log(log) // should log to test output without panic
}
