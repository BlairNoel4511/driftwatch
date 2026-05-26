// Package expiry tracks per-path TTLs and reports whether an entry has expired.
// It is useful for invalidating cached state, suppression windows, and
// acknowledgement lifetimes without coupling those subsystems together.
package expiry

import (
	"fmt"
	"sync"
	"time"
)

// Clock is a function that returns the current time. Replaceable in tests.
type Clock func() time.Time

// Expiry tracks expiration deadlines keyed by an arbitrary string path.
type Expiry struct {
	mu      sync.RWMutex
	entries map[string]time.Time
	clock   Clock
}

// New returns a new Expiry using the provided clock.
// If clock is nil, time.Now is used.
func New(clock Clock) *Expiry {
	if clock == nil {
		clock = time.Now
	}
	return &Expiry{
		entries: make(map[string]time.Time),
		clock:   clock,
	}
}

// Set registers path with an expiration deadline of now+ttl.
// Returns an error if path is empty or ttl is non-positive.
func (e *Expiry) Set(path string, ttl time.Duration) error {
	if path == "" {
		return fmt.Errorf("expiry: path must not be empty")
	}
	if ttl <= 0 {
		return fmt.Errorf("expiry: ttl must be positive, got %s", ttl)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.entries[path] = e.clock().Add(ttl)
	return nil
}

// IsExpired reports whether the entry for path has passed its deadline.
// Returns true when path is unknown (never set or already deleted).
func (e *Expiry) IsExpired(path string) bool {
	e.mu.RLock()
	deadline, ok := e.entries[path]
	e.mu.RUnlock()
	if !ok {
		return true
	}
	return e.clock().After(deadline)
}

// Delete removes the expiry record for path.
func (e *Expiry) Delete(path string) {
	e.mu.Lock()
	delete(e.entries, path)
	e.mu.Unlock()
}

// Purge removes all entries whose deadlines have already passed.
func (e *Expiry) Purge() {
	now := e.clock()
	e.mu.Lock()
	defer e.mu.Unlock()
	for p, dl := range e.entries {
		if now.After(dl) {
			delete(e.entries, p)
		}
	}
}

// Len returns the number of tracked entries (including expired ones).
func (e *Expiry) Len() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.entries)
}
