//go:build linux

package health

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func (r *Repository) CheckMemory() ComponentStatus {
	start := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	total, free, available := readMemInfo()
	latency := time.Since(start)

	status := "ok"
	errMsg := ""
	if total > 0 && float64(available)/float64(total) < 0.1 {
		status = "unhealthy"
		errMsg = "available memory below 10%"
	}

	return ComponentStatus{
		Name:      "memory",
		Status:    status,
		Latency:   latency,
		Error:     errMsg,
		Timestamp: time.Now(),
		Details: map[string]interface{}{
			"go_alloc_bytes": m.Alloc,
			"go_sys_bytes":   m.Sys,
			"total_kb":       total,
			"free_kb":        free,
			"available_kb":   available,
		},
	}
}

func readMemInfo() (total, free, available uint64) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSuffix(fields[0], ":")
		value, _ := strconv.ParseUint(fields[1], 10, 64) // already in KB
		switch key {
		case "MemTotal":
			total = value
		case "MemFree":
			free = value
		case "MemAvailable":
			available = value
		}
	}
	return
}
