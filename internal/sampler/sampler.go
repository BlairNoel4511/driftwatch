// Package sampler provides probabilistic sampling for drift events,
// allowing high-volume environments to reduce alert noise by forwarding
// only a statistical fraction of detected drift events.
package sampler

import (
	"fmt"
	"math/rand"
	"sync"
)

// Sampler decides whether a given drift event should be forwarded
// based on a per-path sampling rate.
type Sampler struct {
	mu      sync.Mutex
	rates   map[string]float64
	default_ float64
	rng     *rand.Rand
}

// New creates a Sampler with the given default sampling rate.
// rate must be in the range (0, 1]. A rate of 1.0 passes all events.
func New(defaultRate float64, src rand.Source) *Sampler {
	if defaultRate <= 0 || defaultRate > 1.0 {
		panic(fmt.Sprintf("sampler: defaultRate must be in (0,1], got %v", defaultRate))
	}
	if src == nil {
		panic("sampler: rand.Source must not be nil")
	}
	return &Sampler{
		rates:    make(map[string]float64),
		default_: defaultRate,
		rng:      rand.New(src), //nolint:gosec
	}
}

// SetRate overrides the sampling rate for a specific path.
// rate must be in the range (0, 1].
func (s *Sampler) SetRate(path string, rate float64) error {
	if rate <= 0 || rate > 1.0 {
		return fmt.Errorf("sampler: rate must be in (0,1], got %v", rate)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rates[path] = rate
	return nil
}

// Allow returns true if the event for the given path should be forwarded.
func (s *Sampler) Allow(path string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	rate, ok := s.rates[path]
	if !ok {
		rate = s.default_
	}
	return s.rng.Float64() < rate
}

// Rate returns the effective sampling rate for a path.
func (s *Sampler) Rate(path string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.rates[path]; ok {
		return r
	}
	return s.default_
}
