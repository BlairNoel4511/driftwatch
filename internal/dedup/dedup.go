// Package dedup provides event deduplication by tracking recently seen
// fingerprints and suppressing identical events within a configurable window.
package dedup

import (
	"sync"
	"time"
)

// Deduplicator suppresses duplicate events within a sliding time window.
type Deduplicator struct {
	mu     sync.Mutex
	window time.Duration
	seen   map[string]time.Time
	now    func() time.Time
}

// New creates a Deduplicator with the given deduplication window.
// Panics if window is zero.
func New(window time.Duration) *Deduplicator {
	if window == 0 {
		panic("dedup: window must be greater than zero")
	}
	return &Deduplicator{
		window: window,
		seen:   make(map[string]time.Time),
		now:    time.Now,
	}
}

// IsDuplicate reports whether key has been seen within the deduplication
// window. If it has not been seen (or the window has expired), it records
// the key and returns false. Otherwise it returns true.
func (d *Deduplicator) IsDuplicate(key string) bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := d.now()
	d.evict(now)

	if _, ok := d.seen[key]; ok {
		return true
	}

	d.seen[key] = now
	return false
}

// Forget removes key from the seen set, allowing the next occurrence to pass.
func (d *Deduplicator) Forget(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.seen, key)
}

// Len returns the number of keys currently tracked.
func (d *Deduplicator) Len() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.seen)
}

// evict removes entries whose window has expired. Must be called with mu held.
func (d *Deduplicator) evict(now time.Time) {
	for k, t := range d.seen {
		if now.Sub(t) >= d.window {
			delete(d.seen, k)
		}
	}
}
