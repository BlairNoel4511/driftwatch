// Package retrier provides retry logic with configurable backoff for
// transient failures encountered during drift checks or alert delivery.
package retrier

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// RetryableFunc is a function that can be retried on failure.
type RetryableFunc func(ctx context.Context) error

// ErrMaxAttemptsReached is returned when all retry attempts are exhausted.
var ErrMaxAttemptsReached = errors.New("retrier: max attempts reached")

// Config holds configuration for a Retrier.
type Config struct {
	// MaxAttempts is the total number of attempts (including the first).
	MaxAttempts int
	// BaseDelay is the initial delay between retries.
	BaseDelay time.Duration
	// MaxDelay caps the exponential backoff delay.
	MaxDelay time.Duration
	// Multiplier scales the delay after each failure (default 2.0).
	Multiplier float64
}

// Retrier executes a RetryableFunc with exponential backoff.
type Retrier struct {
	cfg Config
}

// New creates a new Retrier. Panics if MaxAttempts < 1, BaseDelay <= 0, or
// MaxDelay < BaseDelay.
func New(cfg Config) *Retrier {
	if cfg.MaxAttempts < 1 {
		panic("retrier: MaxAttempts must be >= 1")
	}
	if cfg.BaseDelay <= 0 {
		panic("retrier: BaseDelay must be > 0")
	}
	if cfg.MaxDelay < cfg.BaseDelay {
		panic("retrier: MaxDelay must be >= BaseDelay")
	}
	if cfg.Multiplier <= 0 {
		cfg.Multiplier = 2.0
	}
	return &Retrier{cfg: cfg}
}

// Do executes fn up to MaxAttempts times. It stops early if ctx is cancelled
// or fn returns nil. On exhaustion it returns ErrMaxAttemptsReached wrapping
// the last error.
func (r *Retrier) Do(ctx context.Context, fn RetryableFunc) error {
	delay := r.cfg.BaseDelay
	var lastErr error

	for attempt := 1; attempt <= r.cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		lastErr = fn(ctx)
		if lastErr == nil {
			return nil
		}
		if attempt == r.cfg.MaxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}
		delay = time.Duration(float64(delay) * r.cfg.Multiplier)
		if delay > r.cfg.MaxDelay {
			delay = r.cfg.MaxDelay
		}
	}
	return fmt.Errorf("%w: %w", ErrMaxAttemptsReached, lastErr)
}
