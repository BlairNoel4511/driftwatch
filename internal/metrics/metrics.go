// Package metrics tracks runtime counters for drift detection activity.
package metrics

import (
	"sync"
	"sync/atomic"
)

// Counters holds all runtime metrics for driftwatch.
type Counters struct {
	mu sync.RWMutex

	ChecksTotal   atomic.Int64
	DriftsTotal   atomic.Int64
	AlertsTotal   atomic.Int64
	ErrorsTotal   atomic.Int64

	labels map[string]string
}

// New returns a new Counters instance with optional static labels.
func New(labels map[string]string) *Counters {
	c := &Counters{
		labels: make(map[string]string),
	}
	for k, v := range labels {
		c.labels[k] = v
	}
	return c
}

// RecordCheck increments the total number of checks performed.
func (c *Counters) RecordCheck() {
	c.ChecksTotal.Add(1)
}

// RecordDrift increments the drift counter.
func (c *Counters) RecordDrift() {
	c.DriftsTotal.Add(1)
}

// RecordAlert increments the alert counter.
func (c *Counters) RecordAlert() {
	c.AlertsTotal.Add(1)
}

// RecordError increments the error counter.
func (c *Counters) RecordError() {
	c.ErrorsTotal.Add(1)
}

// Snapshot returns a point-in-time copy of all counter values.
func (c *Counters) Snapshot() map[string]int64 {
	return map[string]int64{
		"checks_total": c.ChecksTotal.Load(),
		"drifts_total": c.DriftsTotal.Load(),
		"alerts_total": c.AlertsTotal.Load(),
		"errors_total": c.ErrorsTotal.Load(),
	}
}

// Labels returns a copy of the static labels attached to this counter set.
func (c *Counters) Labels() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]string, len(c.labels))
	for k, v := range c.labels {
		out[k] = v
	}
	return out
}

// Reset zeroes all counters. Primarily useful in tests.
func (c *Counters) Reset() {
	c.ChecksTotal.Store(0)
	c.DriftsTotal.Store(0)
	c.AlertsTotal.Store(0)
	c.ErrorsTotal.Store(0)
}
