package backoff_test

import (
	"testing"
	"time"

	"driftwatch/internal/backoff"
)

// TestBackoff_ResetResumesFromBase verifies that after a reset the key behaves
// as if it has never been seen before, even mid-sequence.
func TestBackoff_ResetResumesFromBase(t *testing.T) {
	const (
		base = 50 * time.Millisecond
		max  = 400 * time.Millisecond
	)
	b := backoff.New(base, max)

	for i := 0; i < 5; i++ {
		b.Next("path")
	}
	if b.Attempts("path") != 5 {
		t.Fatalf("expected 5 attempts, got %d", b.Attempts("path"))
	}

	b.Reset("path")

	if b.Attempts("path") != 0 {
		t.Fatalf("expected 0 attempts after reset, got %d", b.Attempts("path"))
	}
	got := b.Next("path")
	if got != base {
		t.Fatalf("expected base delay %v after reset, got %v", base, got)
	}
}

// TestBackoff_ConcurrentKeys confirms independent keys do not interfere under
// concurrent access.
func TestBackoff_ConcurrentKeys(t *testing.T) {
	const (
		base    = 10 * time.Millisecond
		max     = 160 * time.Millisecond
		workers = 20
		calls   = 10
	)
	b := backoff.New(base, max)

	done := make(chan struct{}, workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			key := fmt.Sprintf("key-%d", id)
			for j := 0; j < calls; j++ {
				b.Next(key)
			}
			done <- struct{}{}
		}(i)
	}
	for i := 0; i < workers; i++ {
		<-done
	}
}
