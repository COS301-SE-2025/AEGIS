//go:build windows

package health

import (
	"time"
)

func (r *Repository) CheckDisk() ComponentStatus {
	return ComponentStatus{
		Name:      "disk",
		Status:    "ok", // or "skipped"
		Latency:   0,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"note": "disk check not supported on Windows",
		},
	}
}

