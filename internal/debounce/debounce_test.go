package debounce_test

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"driftwatch/internal/debounce"
)

func TestNew_PanicsOnZeroWait(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero wait duration")
		}
	}()
	debounce.New(0, func(string) {})
}

func TestNew_PanicsOnNilCallback(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil callback")
		}
	}()
	debounce.New(10*time.Millisecond, nil)
}

func TestTrigger_CallbackFiredAfterWait(t *testing.T) {
	var fired atomic.Int32
	d := debounce.New(30*time.Millisecond, func(key string) {
		fired.Add(1)
	})

	d.Trigger("path/a")
	time.Sleep(60 * time.Millisecond)

	if fired.Load() != 1 {
		t.Fatalf("expected callback fired once, got %d", fired.Load())
	}
}

func TestTrigger_DebouncesRapidEvents(t *testing.T) {
	var fired atomic.Int32
	d := debounce.New(50*time.Millisecond, func(string) {
		fired.Add(1)
	})

	for i := 0; i < 5; i++ {
		d.Trigger("path/b")
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)

	if fired.Load() != 1 {
		t.Fatalf("expected exactly 1 callback, got %d", fired.Load())
	}
}

func TestTrigger_IndependentKeys(t *testing.T) {
	var mu sync.Mutex
	fired := map[string]int{}
	d := debounce.New(30*time.Millisecond, func(key string) {
		mu.Lock()
		fired[key]++
		mu.Unlock()
	})

	d.Trigger("key1")
	d.Trigger("key2")
	time.Sleep(80 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	if fired["key1"] != 1 || fired["key2"] != 1 {
		t.Fatalf("expected one fire per key, got %v", fired)
	}
}

func TestCancel_PreventsCallback(t *testing.T) {
	var fired atomic.Int32
	d := debounce.New(50*time.Millisecond, func(string) {
		fired.Add(1)
	})

	d.Trigger("path/c")
	d.Cancel("path/c")
	time.Sleep(100 * time.Millisecond)

	if fired.Load() != 0 {
		t.Fatalf("expected no callback after cancel, got %d", fired.Load())
	}
}

func TestPending_ReflectsActiveTimers(t *testing.T) {
	d := debounce.New(100*time.Millisecond, func(string) {})

	if d.Pending() != 0 {
		t.Fatal("expected 0 pending timers initially")
	}

	d.Trigger("x")
	d.Trigger("y")
	if d.Pending() != 2 {
		t.Fatalf("expected 2 pending, got %d", d.Pending())
	}

	time.Sleep(200 * time.Millisecond)
	if d.Pending() != 0 {
		t.Fatalf("expected 0 pending after timers fire, got %d", d.Pending())
	}
}
