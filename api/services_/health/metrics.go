package health

import (
	"github.com/prometheus/client_golang/prometheus"
	
)

// Prometheus metrics
var (
	DependencyErrors = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dependency_errors_total",
			Help: "Total number of dependency check errors",
		},
		[]string{"dependency"},
	)

	DependencyLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dependency_check_latency_seconds",
			Help:    "Latency of dependency health checks",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"dependency"},
	)

	MemoryUsage = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_memory_bytes",
			Help: "Current memory usage of the application",
		},
	)

	NumGoroutines = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_goroutines",
			Help: "Number of goroutines",
		},
	)

	DiskAvailable = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_disk_available_bytes",
			Help: "Available disk space in bytes",
		},
	)
)

// RegisterMetrics must be called in main.go
func RegisterMetrics() {
	prometheus.MustRegister(DependencyErrors)
	prometheus.MustRegister(DependencyLatency)
	prometheus.MustRegister(MemoryUsage)
	prometheus.MustRegister(NumGoroutines)
	prometheus.MustRegister(DiskAvailable)
}

