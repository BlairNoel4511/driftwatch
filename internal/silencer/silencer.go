// Package silencer provides a mechanism to temporarily silence alerts
// for specific paths, preventing repeated notifications during known
// maintenance windows or acknowledged incidents.
package silencer

import (
	"sync"
	"time"
)

// Silencer tracks silenced paths and their expiry times.
type Silencer struct {
	mu      sync.Mutex
	silenced map[string]time.Time
	now     func() time.Time
}

// New creates a new Silencer.
func New() *Silencer {
	return &Silencer{
		silenced: make(map[string]time.Time),
		now:      time.Now,
	}
}

// Silence suppresses alerts for the given path for the specified duration.
// Calling Silence on an already-silenced path extends the silence window.
func (s *Silencer) Silence(path string, duration time.Duration) {
	if duration <= 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.silenced[path] = s.now().Add(duration)
}

// IsSilenced reports whether the given path is currently silenced.
func (s *Silencer) IsSilenced(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	expiry, ok := s.silenced[path]
	if !ok {
		return false
	}
	if s.now().After(expiry) {
		delete(s.silenced, path)
		return false
	}
	return true
}

// Lift removes the silence for the given path immediately.
func (s *Silencer) Lift(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.silenced, path)
}

// Active returns a snapshot of all currently silenced paths and their
// remaining durations.
func (s *Silencer) Active() map[string]time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now()
	out := make(map[string]time.Duration)
	for path, expiry := range s.silenced {
		if remaining := expiry.Sub(now); remaining > 0 {
			out[path] = remaining
		} else {
			delete(s.silenced, path)
		}
	}
	return out
}
