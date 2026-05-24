// Package jitter adds randomised offsets to durations to prevent
// thundering-herd problems when many goroutines fire at the same time.
package jitter

import (
	"math/rand"
	"sync"
	"time"
)

// Jitter applies a random fraction of a base duration as an additive offset.
// It is safe for concurrent use.
type Jitter struct {
	mu      sync.Mutex
	rng     *rand.Rand
	factor  float64 // fraction of base to use as max jitter, e.g. 0.25
}

// New creates a Jitter with the given factor.
// factor must be in the range (0, 1]; it controls how much randomness is
// added — e.g. 0.25 means up to 25 % of the base duration is added.
// Panics if factor is not in (0, 1].
func New(factor float64) *Jitter {
	if factor <= 0 || factor > 1 {
		panic("jitter: factor must be in (0, 1]")
	}
	return &Jitter{
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
		factor: factor,
	}
}

// Apply returns base plus a random offset in [0, base*factor).
func (j *Jitter) Apply(base time.Duration) time.Duration {
	j.mu.Lock()
	offset := time.Duration(float64(base) * j.factor * j.rng.Float64())
	j.mu.Unlock()
	return base + offset
}

// ApplyRange returns a duration uniformly distributed in [min, max).
// Panics if min >= max.
func (j *Jitter) ApplyRange(min, max time.Duration) time.Duration {
	if min >= max {
		panic("jitter: min must be less than max")
	}
	j.mu.Lock()
	d := min + time.Duration(j.rng.Int63n(int64(max-min)))
	j.mu.Unlock()
	return d
}
