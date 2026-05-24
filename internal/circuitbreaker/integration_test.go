package circuitbreaker_test

import (
	"sync"
	"testing"
	"time"

	"driftwatch/internal/circuitbreaker"
)

// TestCircuitBreaker_ConcurrentAccess verifies the breaker is safe under
// concurrent reads and writes across multiple goroutines.
func TestCircuitBreaker_ConcurrentAccess(t *testing.T) {
	b := circuitbreaker.New(5, 50*time.Millisecond)
	paths := []string{"/a", "/b", "/c"}

	var wg sync.WaitGroup
	for i := 0; i < 30; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			p := paths[i%len(paths)]
			if i%3 == 0 {
				b.RecordFailure(p)
			} else if i%3 == 1 {
				b.RecordSuccess(p)
			} else {
				_ = b.Allow(p)
			}
		}(i)
	}
	wg.Wait() // must not race
}

// TestCircuitBreaker_FullCycle exercises the closed → open → half-open →
// closed lifecycle end-to-end.
func TestCircuitBreaker_FullCycle(t *testing.T) {
	now := time.Now()
	b := circuitbreaker.New(2, 100*time.Millisecond)
	// Expose clock via the unexported field is not possible from outside the
	// package, so we rely on real time with a short cooldown.
	path := "/etc/passwd"

	// Closed → failures accumulate.
	b.RecordFailure(path)
	if b.StateOf(path) != circuitbreaker.StateClosed {
		t.Fatal("should still be closed after one failure")
	}

	b.RecordFailure(path)
	if b.StateOf(path) != circuitbreaker.StateOpen {
		t.Fatal("should be open after threshold")
	}

	if err := b.Allow(path); err != circuitbreaker.ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}

	_ = now // keep import tidy

	// Wait for cooldown.
	time.Sleep(150 * time.Millisecond)

	if err := b.Allow(path); err != nil {
		t.Fatalf("expected nil in half-open probe, got %v", err)
	}
	if b.StateOf(path) != circuitbreaker.StateHalfOpen {
		t.Fatal("expected half-open after cooldown")
	}

	// Successful probe closes the circuit.
	b.RecordSuccess(path)
	if b.StateOf(path) != circuitbreaker.StateClosed {
		t.Fatal("expected closed after success")
	}
}
