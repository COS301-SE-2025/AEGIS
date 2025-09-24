//go:build linux

package health

import (
	"time"

	"golang.org/x/sys/unix"
)

// CheckDisk checks disk usage on the given path (e.g. "/" or "/data")
func (r *Repository) CheckDisk() ComponentStatus {
	start := time.Now()

	var stat unix.Statfs_t
	err := unix.Statfs("/", &stat)
	latency := time.Since(start)

	status := "ok"
	errMsg := ""
	var total, available, used uint64

	if err == nil {
		total = stat.Blocks * uint64(stat.Bsize)
		available = stat.Bavail * uint64(stat.Bsize)
		used = total - available

		if total > 0 && float64(available)/float64(total) < 0.1 {
			status = "unhealthy"
			errMsg = "disk space below 10% free"
		}
	} else {
		status = "unhealthy"
		errMsg = err.Error()
		DependencyErrors.WithLabelValues("disk").Inc()
	}

	return ComponentStatus{
		Name:      "disk",
		Status:    status,
		Latency:   latency,
		Error:     errMsg,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"total_bytes":     total,
			"used_bytes":      used,
			"available_bytes": available,
		},
	}
}
