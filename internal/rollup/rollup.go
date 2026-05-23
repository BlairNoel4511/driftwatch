// Package rollup aggregates multiple drift events within a time window
// into a single batched notification, reducing alert noise.
package rollup

import (
	"context"
	"sync"
	"time"

	"github.com/yourorg/driftwatch/internal/watcher"
)

// Batch holds a collection of drift events accumulated during a window.
type Batch struct {
	Events    []watcher.DriftEvent
	WindowEnd time.Time
}

// Handler is called with a batch of drift events at the end of each window.
type Handler func(batch Batch)

// Rollup buffers DriftEvents and flushes them as a Batch after a window.
type Rollup struct {
	mu      sync.Mutex
	window  time.Duration
	handler Handler
	buf     []watcher.DriftEvent
	timer   *time.Timer
}

// New creates a Rollup that flushes accumulated events after window duration.
// Panics if window is zero or handler is nil.
func New(window time.Duration, handler Handler) *Rollup {
	if window <= 0 {
		panic("rollup: window must be greater than zero")
	}
	if handler == nil {
		panic("rollup: handler must not be nil")
	}
	return &Rollup{
		window:  window,
		handler: handler,
	}
}

// Add enqueues a DriftEvent. If no flush is pending, it starts the window timer.
func (r *Rollup) Add(event watcher.DriftEvent) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buf = append(r.buf, event)
	if r.timer == nil {
		r.timer = time.AfterFunc(r.window, r.flush)
	}
}

// flush drains the buffer and invokes the handler. Called by the timer.
func (r *Rollup) flush() {
	r.mu.Lock()
	events := make([]watcher.DriftEvent, len(r.buf))
	copy(events, r.buf)
	r.buf = r.buf[:0]
	r.timer = nil
	r.mu.Unlock()

	if len(events) == 0 {
		return
	}
	r.handler(Batch{
		Events:    events,
		WindowEnd: time.Now(),
	})
}

// Run listens on eventCh and forwards events to Add until ctx is cancelled.
func (r *Rollup) Run(ctx context.Context, eventCh <-chan watcher.DriftEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-eventCh:
			if !ok {
				return
			}
			r.Add(ev)
		}
	}
}
