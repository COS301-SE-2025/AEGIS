package reportshared

import "time"

// Report struct for cross-service use
// Add fields as needed for AI and context autofill

type Report struct {
	ID        string
	Title     string
	CreatedAt time.Time
	// Add other fields as needed
}
