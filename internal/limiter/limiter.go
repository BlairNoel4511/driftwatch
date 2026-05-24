// Package limiter provides a concurrency limiter that caps the number of
// simultaneous in-flight operations across monitored paths.
package limiter

import (
	"context"
	"fmt"
	"sync"
)

// Limiter gates concurrent access using a semaphore keyed by an optional
// namespace. A zero namespace uses the shared global slot pool.
type Limiter struct {
	mu      sync.Mutex
	slots   chan struct{}
	maxCon  int
}

// New creates a Limiter that allows at most maxConcurrent simultaneous
// Acquire calls to proceed. It panics if maxConcurrent is zero.
func New(maxConcurrent int) *Limiter {
	if maxConcurrent <= 0 {
		panic(fmt.Sprintf("limiter: maxConcurrent must be > 0, got %d", maxConcurrent))
	}
	return &Limiter{
		slots:  make(chan struct{}, maxConcurrent),
		maxCon: maxConcurrent,
	}
}

// Acquire blocks until a slot is available or ctx is cancelled.
// Returns an error if the context is done before a slot is obtained.
func (l *Limiter) Acquire(ctx context.Context) error {
	select {
	case l.slots <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Release frees a previously acquired slot. It is a no-op if called
// more times than Acquire has succeeded.
func (l *Limiter) Release() {
	select {
	case <-l.slots:
	default:
	}
}

// InFlight returns the number of currently acquired slots.
func (l *Limiter) InFlight() int {
	return len(l.slots)
}

// Cap returns the maximum concurrency configured at construction.
func (l *Limiter) Cap() int {
	return l.maxCon
}
