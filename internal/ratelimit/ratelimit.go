// Package ratelimit provides a token-bucket rate limiter for suppressing
// repeated drift alerts for the same file path within a cooldown window.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter tracks per-path alert timestamps and suppresses duplicates
// that arrive within the configured cooldown duration.
type Limiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	last     map[string]time.Time
	now      func() time.Time // injectable for testing
}

// New creates a Limiter with the given cooldown window.
// Panics if cooldown is zero or negative.
func New(cooldown time.Duration) *Limiter {
	if cooldown <= 0 {
		panic("ratelimit: cooldown must be positive")
	}
	return &Limiter{
		cooldown: cooldown,
		last:     make(map[string]time.Time),
		now:      time.Now,
	}
}

// Allow reports whether an alert for the given path should be allowed
// through. It returns true the first time a path is seen, and again only
// after the cooldown window has elapsed since the last allowed event.
func (l *Limiter) Allow(path string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	if t, ok := l.last[path]; ok && now.Sub(t) < l.cooldown {
		return false
	}
	l.last[path] = now
	return true
}

// Reset clears the recorded timestamp for path, allowing the next
// call to Allow to pass immediately regardless of cooldown state.
func (l *Limiter) Reset(path string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.last, path)
}

// Len returns the number of paths currently tracked.
func (l *Limiter) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.last)
}
