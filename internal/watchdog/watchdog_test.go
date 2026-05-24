package watchdog_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"driftwatch/internal/watchdog"
)

func TestNew_PanicsOnZeroDeadline(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero deadline")
		}
	}()
	watchdog.New(0, func(time.Duration) {})
}

func TestNew_PanicsOnNilHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil handler")
		}
	}()
	watchdog.New(time.Second, nil)
}

func TestKick_ResetsTimer(t *testing.T) {
	wd := watchdog.New(time.Second, func(time.Duration) {
		t.Error("handler should not fire after kick")
	})
	before := wd.LastKick()
	time.Sleep(5 * time.Millisecond)
	wd.Kick()
	if !wd.LastKick().After(before) {
		t.Fatal("LastKick should advance after Kick")
	}
}

func TestRun_FiresHandlerWhenStalled(t *testing.T) {
	var fired atomic.Int32
	deadline := 40 * time.Millisecond
	wd := watchdog.New(deadline, func(missed time.Duration) {
		fired.Add(1)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	go wd.Run(ctx)

	// Do not kick — watchdog should fire.
	<-ctx.Done()
	if fired.Load() == 0 {
		t.Fatal("expected handler to fire at least once")
	}
}

func TestRun_DoesNotFireWhenKickedRegularly(t *testing.T) {
	var fired atomic.Int32
	deadline := 60 * time.Millisecond
	wd := watchdog.New(deadline, func(missed time.Duration) {
		fired.Add(1)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go wd.Run(ctx)

	// Kick every 15 ms for 120 ms — well within the 60 ms deadline.
	ticker := time.NewTicker(15 * time.Millisecond)
	defer ticker.Stop()
	timer := time.NewTimer(120 * time.Millisecond)
	defer timer.Stop()
loop:
	for {
		select {
		case <-ticker.C:
			wd.Kick()
		case <-timer.C:
			break loop
		}
	}
	cancel()
	if fired.Load() != 0 {
		t.Fatalf("handler fired %d times unexpectedly", fired.Load())
	}
}

func TestRun_StopsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	wd := watchdog.New(50*time.Millisecond, func(time.Duration) {})
	done := make(chan struct{})
	go func() {
		wd.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Run did not stop after context cancel")
	}
}
