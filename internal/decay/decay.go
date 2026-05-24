// Package decay provides a score tracker that reduces a per-path score
// toward zero over time, useful for gradually clearing drift severity
// after repeated clean checks.
package decay

import (
	"sync"
	"time"
)

// Decayer holds per-path scores that decay linearly toward zero based on
// elapsed time since the last update.
type Decayer struct {
	mu       sync.Mutex
	rate     float64 // points per second to subtract
	entries  map[string]*entry
	now      func() time.Time
}

type entry struct {
	score     float64
	updatedAt time.Time
}

// New returns a Decayer that reduces scores by rate points per second.
// Panics if rate is zero or negative.
func New(rate float64) *Decayer {
	if rate <= 0 {
		panic("decay: rate must be positive")
	}
	return &Decayer{
		rate:    rate,
		entries: make(map[string]*entry),
		now:     time.Now,
	}
}

// Add increases the score for path by delta (delta must be > 0).
func (d *Decayer) Add(path string, delta float64) {
	if delta <= 0 {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	e := d.current(path)
	e.score += delta
	e.updatedAt = d.now()
}

// Score returns the current decayed score for path.
func (d *Decayer) Score(path string) float64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	e := d.current(path)
	return e.score
}

// Reset clears the score for path.
func (d *Decayer) Reset(path string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.entries, path)
}

// current applies decay and returns the live entry for path.
// Must be called with d.mu held.
func (d *Decayer) current(path string) *entry {
	now := d.now()
	e, ok := d.entries[path]
	if !ok {
		e = &entry{updatedAt: now}
		d.entries[path] = e
		return e
	}
	elapsed := now.Sub(e.updatedAt).Seconds()
	if elapsed > 0 {
		e.score -= elapsed * d.rate
		if e.score < 0 {
			e.score = 0
		}
		e.updatedAt = now
	}
	return e
}
