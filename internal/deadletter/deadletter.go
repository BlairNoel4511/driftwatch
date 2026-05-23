// Package deadletter provides a dead-letter queue for drift events that
// could not be processed or delivered after exhausting retry attempts.
package deadletter

import (
	"sync"
	"time"

	"github.com/driftwatch/driftwatch/internal/watcher"
)

// Entry holds a failed drift event along with metadata about the failure.
type Entry struct {
	Event     watcher.DriftEvent
	Reason    string
	FailedAt  time.Time
	Attempts  int
}

// Queue is a bounded dead-letter queue for undeliverable drift events.
type Queue struct {
	mu      sync.Mutex
	entries []Entry
	cap     int
}

// New creates a Queue with the given maximum capacity.
// It panics if cap is zero.
func New(cap int) *Queue {
	if cap == 0 {
		panic("deadletter: capacity must be greater than zero")
	}
	return &Queue{
		entries: make([]Entry, 0, cap),
		cap:     cap,
	}
}

// Push adds a failed event to the queue. If the queue is full the oldest
// entry is evicted to make room (FIFO eviction).
func (q *Queue) Push(event watcher.DriftEvent, reason string, attempts int) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.entries) >= q.cap {
		q.entries = q.entries[1:]
	}
	q.entries = append(q.entries, Entry{
		Event:    event,
		Reason:   reason,
		FailedAt: time.Now(),
		Attempts: attempts,
	})
}

// All returns a snapshot of all current entries.
func (q *Queue) All() []Entry {
	q.mu.Lock()
	defer q.mu.Unlock()

	out := make([]Entry, len(q.entries))
	copy(out, q.entries)
	return out
}

// Len returns the number of entries currently in the queue.
func (q *Queue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.entries)
}

// Clear removes all entries from the queue.
func (q *Queue) Clear() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.entries = q.entries[:0]
}
