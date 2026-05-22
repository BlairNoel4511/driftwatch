package scheduler_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/scheduler"
)

func TestNew_PanicsOnZeroInterval(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero interval")
		}
	}()
	scheduler.New(0, func(_ context.Context, _ time.Time) {})
}

func TestNew_PanicsOnNilFunc(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil fn")
		}
	}()
	scheduler.New(time.Second, nil)
}

func TestScheduler_Interval(t *testing.T) {
	s := scheduler.New(42*time.Millisecond, func(_ context.Context, _ time.Time) {})
	if s.Interval() != 42*time.Millisecond {
		t.Fatalf("expected 42ms, got %v", s.Interval())
	}
}

func TestScheduler_TicksMultipleTimes(t *testing.T) {
	var count atomic.Int64

	s := scheduler.New(20*time.Millisecond, func(_ context.Context, _ time.Time) {
		count.Add(1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Millisecond)
	defer cancel()

	err := s.Run(ctx)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}

	got := count.Load()
	if got < 2 {
		t.Fatalf("expected at least 2 ticks, got %d", got)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	s := scheduler.New(50*time.Millisecond, func(_ context.Context, _ time.Time) {})

	done := make(chan error, 1)
	go func() {
		done <- s.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != context.Canceled {
			t.Fatalf("expected Canceled, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("scheduler did not stop after context cancel")
	}
}
