// auditlog/adapter.go
package auditlog

import (
	"context"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AuditLogAdapter adapts auditlog.AuditLogger to the coc.Auditor interface
type AuditLogAdapter struct {
	AuditLogger *AuditLogger
}

// Log adapts the Log method signature to fit the coc.Auditor interface
func (a *AuditLogAdapter) Log(ctx context.Context, typ string, fields map[string]any) {

	// Convert fields map to a map[string]string for Metadata
	metadata := make(map[string]string)
	for key, value := range fields {
		// Ensure we only store string values in the metadata map
		if strVal, ok := value.(string); ok {
			metadata[key] = strVal
		} else {
			metadata[key] = fmt.Sprintf("%v", value) // Convert non-strings to string
		}
	}
	// Map the fields to an AuditLog structure
	log := AuditLog{
		ID:          generateUUID(),            // Generate a new unique ID
		Timestamp:   time.Now(),                // Use current timestamp
		Action:      typ,                       // Log type (e.g., "UPLOAD_EVIDENCE")
		Actor:       fields["actor"].(Actor),   // Extract actor info
		Target:      fields["target"].(Target), // Extract target info (evidence)
		Service:     fields["service"].(string),
		Status:      fields["status"].(string),
		Description: fields["description"].(string),
		Metadata:    metadata, // Use the mapped metadata
	}

	// Convert context into a gin.Context if needed, or use nil as fallback
	var ginCtx *gin.Context
	if c, ok := ctx.(*gin.Context); ok {
		ginCtx = c
	}

	// Call the original AuditLogger's Log method (Gin context and AuditLog)
	if err := a.AuditLogger.Log(ginCtx, log); err != nil {
		// Handle logging error (Optional: log to another place or return error)
	}
}

func generateUUID() string {
	return uuid.New().String()
}
