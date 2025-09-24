//go:build windows
// +build windows

package health

import (
	"runtime"
)

func UpdateResourceMetrics() {
	// Memory
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	MemoryUsage.Set(float64(m.Alloc))

	// Goroutines
	NumGoroutines.Set(float64(runtime.NumGoroutine()))

	// Disk (not supported on Windows for now)
	DiskAvailable.Set(0)
}
