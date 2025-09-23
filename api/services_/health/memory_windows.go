//go:build windows

package health

import (
	"runtime"
	"time"
)

func (r *Repository) CheckMemory() ComponentStatus {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return ComponentStatus{
		Name:      "memory",
		Status:    "ok",
		Latency:   0,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"go_alloc_bytes": m.Alloc,
			"go_sys_bytes":   m.Sys,
		},
	}
}
