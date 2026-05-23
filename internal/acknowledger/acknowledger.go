// Package acknowledger tracks which drift events have been acknowledged
// by an operator, suppressing repeat alerts for known divergences.
package acknowledger

import (
	"sync"
	"time"
)

// Entry records when a path was acknowledged and for how long.
type Entry struct {
	Path       string
	AckedAt    time.Time
	Expiry     time.Time
	AckedBy    string
}

// Acknowledger stores acknowledgement state for monitored paths.
type Acknowledger struct {
	mu      sync.RWMutex
	entries map[string]Entry
	now     func() time.Time
}

// New returns a new Acknowledger.
func New() *Acknowledger {
	return &Acknowledger{
		entries: make(map[string]Entry),
		now:     time.Now,
	}
}

// Acknowledge marks path as acknowledged by ackedBy for the given duration.
// Subsequent calls to IsAcknowledged will return true until the duration elapses.
func (a *Acknowledger) Acknowledge(path, ackedBy string, duration time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	a.entries[path] = Entry{
		Path:    path,
		AckedAt: now,
		Expiry:  now.Add(duration),
		AckedBy: ackedBy,
	}
}

// IsAcknowledged reports whether path currently has a valid (non-expired) acknowledgement.
func (a *Acknowledger) IsAcknowledged(path string) bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	e, ok := a.entries[path]
	if !ok {
		return false
	}
	return a.now().Before(e.Expiry)
}

// Revoke removes an acknowledgement for path, re-enabling alerts immediately.
func (a *Acknowledger) Revoke(path string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.entries, path)
}

// Get returns the Entry for path and whether it exists (expired or not).
func (a *Acknowledger) Get(path string) (Entry, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()
	e, ok := a.entries[path]
	return e, ok
}

// Purge removes all expired acknowledgements.
func (a *Acknowledger) Purge() {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	for path, e := range a.entries {
		if !now.Before(e.Expiry) {
			delete(a.entries, path)
		}
	}
}
