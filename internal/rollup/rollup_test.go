package rollup_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/rollup"
	"github.com/yourorg/driftwatch/internal/watcher"
)

func makeEvent(path string) watcher.DriftEvent {
	return watcher.DriftEvent{Path: path}
}

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero window")
		}
	}()
	rollup.New(0, func(rollup.Batch) {})
}

func TestNew_PanicsOnNilHandler(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil handler")
		}
	}()
	rollup.New(time.Second, nil)
}

func TestRollup_BatchesEvents(t *testing.T) {
	var mu sync.Mutex
	var batches []rollup.Batch

	r := rollup.New(50*time.Millisecond, func(b rollup.Batch) {
		mu.Lock()
		batches = append(batches, b)
		mu.Unlock()
	})

	r.Add(makeEvent("/etc/hosts"))
	r.Add(makeEvent("/etc/passwd"))
	r.Add(makeEvent("/etc/shadow"))

	time.Sleep(120 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if len(batches) != 1 {
		t.Fatalf("expected 1 batch, got %d", len(batches))
	}
	if len(batches[0].Events) != 3 {
		t.Fatalf("expected 3 events in batch, got %d", len(batches[0].Events))
	}
}

func TestRollup_MultipleWindows(t *testing.T) {
	var mu sync.Mutex
	var count int

	r := rollup.New(40*time.Millisecond, func(b rollup.Batch) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	r.Add(makeEvent("/a"))
	time.Sleep(80 * time.Millisecond)
	r.Add(makeEvent("/b"))
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if count != 2 {
		t.Fatalf("expected 2 flushes, got %d", count)
	}
}

func TestRollup_Run_StopsOnContextCancel(t *testing.T) {
	ch := make(chan watcher.DriftEvent, 1)
	r := rollup.New(time.Second, func(rollup.Batch) {})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.Run(ctx, ch)
		close(done)
	}()

	cancel()
	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Run did not stop after context cancel")
	}
}
