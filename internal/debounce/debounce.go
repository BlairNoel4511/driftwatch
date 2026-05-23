// Package debounce provides a mechanism to suppress rapid repeated drift
// events for the same path, emitting only after a quiet period has elapsed.
package debounce

import (
	"sync"
	"time"
)

// Debouncer delays forwarding of events until no new event for the same key
// has arrived within the configured wait duration.
type Debouncer struct {
	wait    time.Duration
	mu      sync.Mutex
	timers  map[string]*time.Timer
	callback func(key string)
}

// New creates a Debouncer that invokes callback(key) after wait has elapsed
// since the last call to Trigger(key).
// Panics if wait is zero or callback is nil.
func New(wait time.Duration, callback func(key string)) *Debouncer {
	if wait <= 0 {
		panic("debounce: wait duration must be greater than zero")
	}
	if callback == nil {
		panic("debounce: callback must not be nil")
	}
	return &Debouncer{
		wait:     wait,
		timers:   make(map[string]*time.Timer),
		callback: callback,
	}
}

// Trigger schedules callback(key) to fire after the wait duration.
// If Trigger is called again for the same key before the timer fires,
// the timer is reset, effectively debouncing rapid events.
func (d *Debouncer) Trigger(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.timers[key]; ok {
		t.Reset(d.wait)
		return
	}

	d.timers[key] = time.AfterFunc(d.wait, func() {
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
		d.callback(key)
	})
}

// Cancel cancels any pending timer for key without invoking the callback.
func (d *Debouncer) Cancel(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if t, ok := d.timers[key]; ok {
		t.Stop()
		delete(d.timers, key)
	}
}

// Pending returns the number of keys with active pending timers.
func (d *Debouncer) Pending() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.timers)
}
