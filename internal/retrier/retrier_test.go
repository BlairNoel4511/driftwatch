package retrier_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"driftwatch/internal/retrier"
)

var errTransient = errors.New("transient failure")

func TestNew_PanicsOnZeroAttempts(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for MaxAttempts=0")
		}
	}()
	retrier.New(retrier.Config{MaxAttempts: 0, BaseDelay: time.Millisecond, MaxDelay: time.Second})
}

func TestNew_PanicsOnZeroBaseDelay(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for BaseDelay=0")
		}
	}()
	retrier.New(retrier.Config{MaxAttempts: 3, BaseDelay: 0, MaxDelay: time.Second})
}

func TestNew_PanicsWhenMaxDelayLessThanBase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when MaxDelay < BaseDelay")
		}
	}()
	retrier.New(retrier.Config{MaxAttempts: 3, BaseDelay: time.Second, MaxDelay: time.Millisecond})
}

func TestDo_SucceedsOnFirstAttempt(t *testing.T) {
	r := retrier.New(retrier.Config{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: time.Second})
	calls := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		calls++
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected 1 call, got %d", calls)
	}
}

func TestDo_RetriesAndSucceeds(t *testing.T) {
	r := retrier.New(retrier.Config{MaxAttempts: 5, BaseDelay: time.Millisecond, MaxDelay: 10 * time.Millisecond})
	calls := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		calls++
		if calls < 3 {
			return errTransient
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_ExhaustsAttemptsReturnsWrappedError(t *testing.T) {
	r := retrier.New(retrier.Config{MaxAttempts: 3, BaseDelay: time.Millisecond, MaxDelay: 5 * time.Millisecond})
	calls := 0
	err := r.Do(context.Background(), func(_ context.Context) error {
		calls++
		return errTransient
	})
	if !errors.Is(err, retrier.ErrMaxAttemptsReached) {
		t.Fatalf("expected ErrMaxAttemptsReached, got %v", err)
	}
	if !errors.Is(err, errTransient) {
		t.Fatalf("expected wrapped errTransient, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 calls, got %d", calls)
	}
}

func TestDo_StopsOnContextCancel(t *testing.T) {
	r := retrier.New(retrier.Config{MaxAttempts: 10, BaseDelay: 50 * time.Millisecond, MaxDelay: time.Second})
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	err := r.Do(ctx, func(_ context.Context) error {
		calls++
		if calls == 1 {
			cancel()
		}
		return errTransient
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
