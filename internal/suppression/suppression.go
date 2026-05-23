// Package suppression provides a mechanism to suppress duplicate drift alerts
// for a given path within a configurable time window.
package suppression

import (
	"sync"
	"time"
)

// Suppressor tracks recently alerted paths and suppresses duplicates
// within the configured window.
type Suppressor struct {
	mu     sync.Mutex
	window time.Duration
	last   map[string]time.Time
	now    func() time.Time
}

// New creates a Suppressor with the given suppression window.
// Panics if window is zero or negative.
func New(window time.Duration) *Suppressor {
	if window <= 0 {
		panic("suppression: window must be positive")
	}
	return &Suppressor{
		window: window,
		last:   make(map[string]time.Time),
		now:    time.Now,
	}
}

// Allow returns true if an alert for path should be emitted, i.e. either
// it has never been alerted or the suppression window has expired.
// When Allow returns true it records the current time for path.
func (s *Suppressor) Allow(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	if t, ok := s.last[path]; ok && now.Sub(t) < s.window {
		return false
	}
	s.last[path] = now
	return true
}

// Reset clears the suppression record for a specific path,
// allowing the next alert to pass through immediately.
func (s *Suppressor) Reset(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.last, path)
}

// ResetAll clears all suppression records.
func (s *Suppressor) ResetAll() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.last = make(map[string]time.Time)
}

// Len returns the number of paths currently tracked.
func (s *Suppressor) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.last)
}
