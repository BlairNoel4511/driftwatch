// Package history maintains a rolling log of drift events detected
// during a driftwatch session, enabling summary reports and auditing.
package history

import (
	"sync"
	"time"
)

// Event records a single drift detection occurrence.
type Event struct {
	OccurredAt time.Time
	Path       string
	Reason     string
}

// History is a thread-safe, bounded log of drift events.
type History struct {
	mu     sync.Mutex
	events []Event
	max    int
}

// New creates a History that retains at most maxEvents entries.
// If maxEvents is <= 0, a default of 1000 is used.
func New(maxEvents int) *History {
	if maxEvents <= 0 {
		maxEvents = 1000
	}
	return &History{max: maxEvents}
}

// Record appends a new drift event. If the log is full the oldest
// entry is evicted to make room.
func (h *History) Record(path, reason string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if len(h.events) >= h.max {
		h.events = h.events[1:]
	}
	h.events = append(h.events, Event{
		OccurredAt: time.Now(),
		Path:       path,
		Reason:     reason,
	})
}

// All returns a copy of all recorded events, oldest first.
func (h *History) All() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := make([]Event, len(h.events))
	copy(out, h.events)
	return out
}

// Len returns the current number of recorded events.
func (h *History) Len() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.events)
}

// Clear removes all recorded events.
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.events = h.events[:0]
}
