package handlers

import (
	"encoding/json"
	"net/http"

	"aegis-api/services_/health"
)

type HealthHandler struct {
	Service *health.Service
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	resp := h.Service.GetHealth()
	statusCode := http.StatusOK
	if resp.Status != "ok" {
		statusCode = http.StatusServiceUnavailable
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(resp)
}

func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	if h.Service.GetReadiness() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ready"))
	} else {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
	}
}

func (h *HealthHandler) Liveness(w http.ResponseWriter, r *http.Request) {
	if h.Service.GetLiveness() {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("alive"))
	} else {
		http.Error(w, "dead", http.StatusServiceUnavailable)
	}
}
