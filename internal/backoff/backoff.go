// Package backoff provides an exponential back-off strategy for retrying
// failed operations. The delay doubles on each consecutive failure up to a
// configurable maximum, then resets when a success is recorded.
package backoff

import (
	"fmt"
	"sync"
	"time"
)

// Backoff tracks per-key retry state and computes the next wait duration.
type Backoff struct {
	mu      sync.Mutex
	base    time.Duration
	max     time.Duration
	state   map[string]*entry
}

type entry struct {
	attempts int
	next     time.Duration
}

// New creates a Backoff with the given base and max durations.
// Panics if base or max is zero, or base exceeds max.
func New(base, max time.Duration) *Backoff {
	if base <= 0 {
		panic("backoff: base must be greater than zero")
	}
	if max <= 0 {
		panic("backoff: max must be greater than zero")
	}
	if base > max {
		panic(fmt.Sprintf("backoff: base %v exceeds max %v", base, max))
	}
	return &Backoff{
		base:  base,
		max:   max,
		state: make(map[string]*entry),
	}
}

// Next returns the duration to wait before the next retry for the given key
// and increments the internal failure counter.
func (b *Backoff) Next(key string) time.Duration {
	b.mu.Lock()
	defer b.mu.Unlock()

	e, ok := b.state[key]
	if !ok {
		e = &entry{next: b.base}
		b.state[key] = e
	}

	wait := e.next
	e.attempts++
	nextDelay := e.next * 2
	if nextDelay > b.max {
		nextDelay = b.max
	}
	e.next = nextDelay
	return wait
}

// Reset clears the back-off state for the given key, signalling a success.
func (b *Backoff) Reset(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.state, key)
}

// Attempts returns the number of consecutive failures recorded for key.
func (b *Backoff) Attempts(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	if e, ok := b.state[key]; ok {
		return e.attempts
	}
	return 0
}
