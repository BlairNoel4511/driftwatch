// Package scheduler provides a configurable tick-based scheduler that
// triggers periodic checks at a defined interval.
package scheduler

import (
	"context"
	"time"
)

// TickFunc is a function invoked on each scheduled tick.
type TickFunc func(ctx context.Context, at time.Time)

// Scheduler drives periodic execution of a TickFunc.
type Scheduler struct {
	interval time.Duration
	fn       TickFunc
}

// New creates a Scheduler that will call fn every interval.
// interval must be positive; if it is not, New panics.
func New(interval time.Duration, fn TickFunc) *Scheduler {
	if interval <= 0 {
		panic("scheduler: interval must be positive")
	}
	if fn == nil {
		panic("scheduler: fn must not be nil")
	}
	return &Scheduler{interval: interval, fn: fn}
}

// Run blocks, calling the TickFunc on every tick until ctx is cancelled.
// The first tick fires after one full interval. Run returns the context
// error when the context is done.
func (s *Scheduler) Run(ctx context.Context) error {
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			s.fn(ctx, t)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Interval returns the configured tick interval.
func (s *Scheduler) Interval() time.Duration {
	return s.interval
}
