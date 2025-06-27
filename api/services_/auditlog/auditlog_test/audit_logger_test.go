package auditlog_test

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"aegis-api/services_/auditlog"
)

// ---- Mock implementations of the auditlog interfaces ----

// MockMongoLogger mocks the MongoLoggerInterface
type MockMongoLogger struct {
	mock.Mock
}

func (m *MockMongoLogger) Log(ctx *gin.Context, log auditlog.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// MockZapLogger mocks the ZapLoggerInterface
type MockZapLogger struct {
	mock.Mock
}

func (z *MockZapLogger) Log(log auditlog.AuditLog) {
	z.Called(log)
}

func TestAuditLogger_Log(t *testing.T) {
	// Setup mocks
	mockMongo := new(MockMongoLogger)
	mockZap := new(MockZapLogger)

	// Create the unified logger using the mocks
	audit := auditlog.NewAuditLogger(mockMongo, mockZap)

	// Create a basic log entry
	log := auditlog.AuditLog{
		Action:  "LOGIN",
		Service: "auth",
	}

	// Create a dummy Gin context
	ctx, _ := gin.CreateTestContext(nil)

	// Set expectations
	mockMongo.On("Log", ctx, log).Return(nil)
	mockZap.On("Log", log).Return()

	// Call the logger
	err := audit.Log(ctx, log)

	// Assertions
	assert.NoError(t, err)
	mockMongo.AssertExpectations(t)
	mockZap.AssertExpectations(t)
}
