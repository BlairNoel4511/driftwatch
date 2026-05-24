// Package circuitbreaker implements a per-path circuit breaker that opens
// after a configurable number of consecutive failures and resets after a
// cooldown window, preventing alert storms on persistently broken targets.
package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
)

// State represents the current state of a circuit.
type State int

const (
	StateClosed   State = iota // normal operation
	StateOpen                  // failures exceeded threshold; calls rejected
	StateHalfOpen              // cooldown elapsed; one probe allowed
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// ErrOpen is returned when a call is rejected because the circuit is open.
var ErrOpen = fmt.Errorf("circuit open")

type circuit struct {
	failures  int
	state     State
	openedAt  time.Time
}

// Breaker is a thread-safe circuit breaker keyed by path.
type Breaker struct {
	mu        sync.Mutex
	circuits  map[string]*circuit
	threshold int
	cooldown  time.Duration
	now       func() time.Time
}

// New creates a Breaker that opens after threshold consecutive failures and
// attempts recovery after cooldown. Panics on invalid arguments.
func New(threshold int, cooldown time.Duration) *Breaker {
	if threshold <= 0 {
		panic("circuitbreaker: threshold must be > 0")
	}
	if cooldown <= 0 {
		panic("circuitbreaker: cooldown must be > 0")
	}
	return &Breaker{
		circuits:  make(map[string]*circuit),
		threshold: threshold,
		cooldown:  cooldown,
		now:       time.Now,
	}
}

// Allow returns nil if the call for path is permitted, or ErrOpen if the
// circuit is open and the cooldown has not yet elapsed.
func (b *Breaker) Allow(path string) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.get(path)
	switch c.state {
	case StateClosed:
		return nil
	case StateOpen:
		if b.now().Sub(c.openedAt) >= b.cooldown {
			c.state = StateHalfOpen
			return nil
		}
		return ErrOpen
	case StateHalfOpen:
		return nil
	}
	return nil
}

// RecordSuccess resets the failure count for path and closes the circuit.
func (b *Breaker) RecordSuccess(path string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.get(path)
	c.failures = 0
	c.state = StateClosed
}

// RecordFailure increments the failure count for path and opens the circuit
// when the threshold is reached.
func (b *Breaker) RecordFailure(path string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	c := b.get(path)
	c.failures++
	if c.failures >= b.threshold && c.state != StateOpen {
		c.state = StateOpen
		c.openedAt = b.now()
	}
}

// StateOf returns the current State for path.
func (b *Breaker) StateOf(path string) State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.get(path).state
}

func (b *Breaker) get(path string) *circuit {
	if c, ok := b.circuits[path]; ok {
		return c
	}
	c := &circuit{state: StateClosed}
	b.circuits[path] = c
	return c
}
