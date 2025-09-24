package health

import "time"

type ComponentStatus struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

type HealthResponse struct {
	Status     string            `json:"status"`
	Timestamp  time.Time         `json:"timestamp"`
	Components []ComponentStatus `json:"components"`
}
