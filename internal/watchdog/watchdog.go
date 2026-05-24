// Package watchdog provides a liveness watchdog that detects stalled
// check loops and emits a synthetic error event when a deadline is missed.
package watchdog

import (
	"context"
	"sync"
	"time"
)

// Handler is called when the watchdog fires because the deadline was missed.
type Handler func(missed time.Duration)

// Watchdog monitors that Kick is called at least once per deadline interval.
// If no kick arrives in time, Handler is invoked on a background goroutine.
type Watchdog struct {
	deadline time.Duration
	handler  Handler

	mu      sync.Mutex
	lastKick time.Time
}

// New creates a Watchdog with the given deadline and miss handler.
// Panics if deadline is zero or handler is nil.
func New(deadline time.Duration, handler Handler) *Watchdog {
	if deadline <= 0 {
		panic("watchdog: deadline must be positive")
	}
	if handler == nil {
		panic("watchdog: handler must not be nil")
	}
	return &Watchdog{
		deadline:  deadline,
		handler:   handler,
		lastKick:  time.Now(),
	}
}

// Kick resets the watchdog timer. Call this after each successful check cycle.
func (w *Watchdog) Kick() {
	w.mu.Lock()
	w.lastKick = time.Now()
	w.mu.Unlock()
}

// LastKick returns the time of the most recent kick.
func (w *Watchdog) LastKick() time.Time {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.lastKick
}

// Run starts the watchdog loop. It blocks until ctx is cancelled.
// The loop ticks every deadline/2 to detect stalls promptly.
func (w *Watchdog) Run(ctx context.Context) {
	ticker := time.NewTicker(w.deadline / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			w.mu.Lock()
			elapsed := now.Sub(w.lastKick)
			w.mu.Unlock()
			if elapsed > w.deadline {
				w.handler(elapsed - w.deadline)
			}
		}
	}
}
