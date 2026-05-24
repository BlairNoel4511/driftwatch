package retrier_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"driftwatch/internal/retrier"
)

// TestRetrier_ConcurrentDo verifies that multiple goroutines can safely use
// the same Retrier instance concurrently.
func TestRetrier_ConcurrentDo(t *testing.T) {
	r := retrier.New(retrier.Config{
		MaxAttempts: 4,
		BaseDelay:   time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
	})

	const goroutines = 20
	var wg sync.WaitGroup
	errCh := make(chan error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			attempts := 0
			err := r.Do(context.Background(), func(_ context.Context) error {
				attempts++
				if attempts < 2 {
					return errTransient
				}
				return nil
			})
			if err != nil {
				errCh <- err
			}
		}(i)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Errorf("goroutine returned unexpected error: %v", err)
	}
}

// TestRetrier_BackoffCapIsRespected ensures the delay never exceeds MaxDelay.
func TestRetrier_BackoffCapIsRespected(t *testing.T) {
	const maxDelay = 10 * time.Millisecond
	r := retrier.New(retrier.Config{
		MaxAttempts: 6,
		BaseDelay:   2 * time.Millisecond,
		MaxDelay:    maxDelay,
		Multiplier:  10.0, // aggressive multiplier to hit cap quickly
	})

	start := time.Now()
	_ = r.Do(context.Background(), func(_ context.Context) error {
		return errTransient
	})
	elapsed := time.Since(start)

	// With 6 attempts and delays capped at 10ms, total wait <= 5*10ms = 50ms.
	// Allow generous headroom for slow CI environments.
	if elapsed > 500*time.Millisecond {
		t.Fatalf("retrier took too long (%v), backoff cap may not be working", elapsed)
	}
}
