// Package cooldown provides per-path exponential backoff for drift alerts,
// preventing alert storms when a file oscillates between states.
package cooldown

import (
	"sync"
	"time"
)

// Cooldown tracks per-path backoff durations and enforces a minimum quiet
// period that doubles on each consecutive suppressed event.
type Cooldown struct {
	mu      sync.Mutex
	base    time.Duration
	max     time.Duration
	state   map[string]*entry
}

type entry struct {
	current  time.Duration
	readyAt  time.Time
}

// New returns a Cooldown with the given base and maximum backoff durations.
// Panics if base or max are zero, or if base exceeds max.
func New(base, max time.Duration) *Cooldown {
	if base <= 0 {
		panic("cooldown: base duration must be positive")
	}
	if max <= 0 {
		panic("cooldown: max duration must be positive")
	}
	if base > max {
		panic("cooldown: base must not exceed max")
	}
	return &Cooldown{
		base:  base,
		max:   max,
		state: make(map[string]*entry),
	}
}

// Allow reports whether an alert for path is permitted at now.
// If allowed, the cooldown window for path is doubled (up to max).
// If denied, the window is unchanged and the caller should suppress the alert.
func (c *Cooldown) Allow(path string, now time.Time) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.state[path]
	if !ok {
		c.state[path] = &entry{
			current: c.base,
			readyAt: now.Add(c.base),
		}
		return true
	}

	if now.Before(e.readyAt) {
		return false
	}

	next := e.current * 2
	if next > c.max {
		next = c.max
	}
	e.current = next
	e.readyAt = now.Add(next)
	return true
}

// Reset clears the backoff state for path, allowing the next alert immediately.
func (c *Cooldown) Reset(path string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.state, path)
}

// ResetAll clears backoff state for all paths.
func (c *Cooldown) ResetAll() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = make(map[string]*entry)
}
