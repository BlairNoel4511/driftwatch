// Package observer tracks how long a path has been in a drifted state
// and emits severity levels based on configurable age thresholds.
package observer

import (
	"sync"
	"time"
)

// Severity represents how long a drift has been active.
type Severity int

const (
	SeverityNone     Severity = iota
	SeverityWarning           // drift age >= warning threshold
	SeverityCritical          // drift age >= critical threshold
)

func (s Severity) String() string {
	switch s {
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	default:
		return "none"
	}
}

type entry struct {
	firstSeen time.Time
	last      time.Time
}

// Observer records when drift was first observed per path and computes
// a severity based on how long the drift has persisted.
type Observer struct {
	mu       sync.Mutex
	entries  map[string]entry
	warning  time.Duration
	critical time.Duration
	now      func() time.Time
}

// New creates an Observer. warning and critical must be > 0 and
// warning must be < critical.
func New(warning, critical time.Duration) *Observer {
	if warning <= 0 {
		panic("observer: warning threshold must be > 0")
	}
	if critical <= 0 {
		panic("observer: critical threshold must be > 0")
	}
	if warning >= critical {
		panic("observer: warning threshold must be less than critical threshold")
	}
	return &Observer{
		entries:  make(map[string]entry),
		warning:  warning,
		critical: critical,
		now:      time.Now,
	}
}

// Observe records an active drift event for path and returns the current
// severity. Call Resolve to clear the tracked state when drift is resolved.
func (o *Observer) Observe(path string) Severity {
	o.mu.Lock()
	defer o.mu.Unlock()

	now := o.now()
	e, ok := o.entries[path]
	if !ok {
		e = entry{firstSeen: now}
	}
	e.last = now
	o.entries[path] = e

	age := now.Sub(e.firstSeen)
	switch {
	case age >= o.critical:
		return SeverityCritical
	case age >= o.warning:
		return SeverityWarning
	default:
		return SeverityNone
	}
}

// Resolve removes drift tracking for path.
func (o *Observer) Resolve(path string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	delete(o.entries, path)
}

// Age returns how long path has been in a drifted state.
// Returns 0 if the path is not currently tracked.
func (o *Observer) Age(path string) time.Duration {
	o.mu.Lock()
	defer o.mu.Unlock()
	e, ok := o.entries[path]
	if !ok {
		return 0
	}
	return o.now().Sub(e.firstSeen)
}
