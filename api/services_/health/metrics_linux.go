//go:build linux
// +build linux

package health

import (
	"runtime"

	"golang.org/x/sys/unix"
)

func UpdateResourceMetrics() {
	// Memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	MemoryUsage.Set(float64(m.Alloc))

	// Goroutines
	NumGoroutines.Set(float64(runtime.NumGoroutine()))

	// Disk
	var stat unix.Statfs_t
	if err := unix.Statfs("/", &stat); err == nil {
		available := stat.Bavail * uint64(stat.Bsize)
		DiskAvailable.Set(float64(available))
	}
}
