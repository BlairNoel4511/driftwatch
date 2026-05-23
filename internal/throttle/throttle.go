// Package throttle provides a token-bucket style throttle that limits
// how many events can pass through per unit of time across named keys.
package throttle

import (
	"fmt"
	"sync"
	"time"
)

// Throttle limits events to at most Burst occurrences per Window for each key.
type Throttle struct {
	mu     sync.Mutex
	window time.Duration
	burst  int
	buckets map[string]*bucket
}

type bucket struct {
	count     int
	windowEnd time.Time
}

// New creates a Throttle that allows up to burst events per window per key.
// Panics if window is zero or burst is less than 1.
func New(window time.Duration, burst int) *Throttle {
	if window <= 0 {
		panic("throttle: window must be positive")
	}
	if burst < 1 {
		panic("throttle: burst must be at least 1")
	}
	return &Throttle{
		window:  window,
		burst:   burst,
		buckets: make(map[string]*bucket),
	}
}

// Allow returns true if the event for key is within the burst limit for the
// current window, false if the limit has been exceeded.
func (t *Throttle) Allow(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	b, ok := t.buckets[key]
	if !ok || now.After(b.windowEnd) {
		t.buckets[key] = &bucket{count: 1, windowEnd: now.Add(t.window)}
		return true
	}
	if b.count < t.burst {
		b.count++
		return true
	}
	return false
}

// Remaining returns the number of events still allowed in the current window
// for key. Returns burst if no events have been recorded yet.
func (t *Throttle) Remaining(key string) int {
	t.mu.Lock()
	defer t.mu.Unlock()

	b, ok := t.buckets[key]
	if !ok || time.Now().After(b.windowEnd) {
		return t.burst
	}
	remaining := t.burst - b.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Reset clears the throttle state for key.
func (t *Throttle) Reset(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.buckets, key)
}

// String returns a human-readable description of the throttle configuration.
func (t *Throttle) String() string {
	return fmt.Sprintf("Throttle(burst=%d, window=%s)", t.burst, t.window)
}
