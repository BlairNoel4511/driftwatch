// Package healthcheck provides a simple HTTP health endpoint for driftwatch.
package healthcheck

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
	"time"
)

// Status holds the current health state of the daemon.
type Status struct {
	Healthy    bool      `json:"healthy"`
	LastCheck  time.Time `json:"last_check,omitempty"`
	DriftCount int64     `json:"drift_count"`
	Uptime     string    `json:"uptime"`
}

// Handler is an HTTP handler that exposes daemon health.
type Handler struct {
	startTime  time.Time
	driftCount atomic.Int64
	lastCheck  atomic.Value // stores time.Time
}

// New creates a new Handler.
func New() *Handler {
	return &Handler{
		startTime: time.Now(),
	}
}

// RecordCheck updates the last-check timestamp.
func (h *Handler) RecordCheck() {
	h.lastCheck.Store(time.Now())
}

// RecordDrift increments the drift counter.
func (h *Handler) RecordDrift() {
	h.driftCount.Add(1)
}

// ServeHTTP writes a JSON health response.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var lastCheck time.Time
	if v := h.lastCheck.Load(); v != nil {
		lastCheck = v.(time.Time)
	}

	s := Status{
		Healthy:    true,
		LastCheck:  lastCheck,
		DriftCount: h.driftCount.Load(),
		Uptime:     time.Since(h.startTime).Round(time.Second).String(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(s)
}
