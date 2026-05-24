// Package window provides a sliding-window event counter used to track
// how many drift events have occurred within a rolling time period.
package window

import (
	"sync"
	"time"
)

// Window is a thread-safe sliding-window counter keyed by an arbitrary string
// (typically a file path). It records event timestamps and reports how many
// events fall within the configured duration.
type Window struct {
	mu       sync.Mutex
	duration time.Duration
	events   map[string][]time.Time
	now      func() time.Time
}

// New creates a Window with the given sliding duration.
// Panics if duration is zero.
func New(d time.Duration) *Window {
	if d == 0 {
		panic("window: duration must be greater than zero")
	}
	return &Window{
		duration: d,
		events:   make(map[string][]time.Time),
		now:      time.Now,
	}
}

// Record adds a new event for the given key at the current time.
func (w *Window) Record(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := w.now()
	w.events[key] = append(w.prune(key, now), now)
}

// Count returns the number of events for key that fall within the sliding window.
func (w *Window) Count(key string) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	return len(w.prune(key, w.now()))
}

// Reset clears all recorded events for the given key.
func (w *Window) Reset(key string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	delete(w.events, key)
}

// prune removes events older than the window duration and returns the remaining
// slice. Callers must hold w.mu.
func (w *Window) prune(key string, now time.Time) []time.Time {
	cutoff := now.Add(-w.duration)
	old := w.events[key]
	var fresh []time.Time
	for _, t := range old {
		if t.After(cutoff) {
			fresh = append(fresh, t)
		}
	}
	w.events[key] = fresh
	return fresh
}
