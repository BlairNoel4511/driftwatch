package backoff

import (
	"testing"
	"time"
)

const (
	base = 100 * time.Millisecond
	max  = 800 * time.Millisecond
)

func TestNew_PanicsOnZeroBase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero base")
		}
	}()
	New(0, max)
}

func TestNew_PanicsOnZeroMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero max")
		}
	}()
	New(base, 0)
}

func TestNew_PanicsWhenBaseExceedsMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when base > max")
		}
	}()
	New(max, base)
}

func TestNext_FirstCallReturnsBase(t *testing.T) {
	b := New(base, max)
	got := b.Next("key")
	if got != base {
		t.Fatalf("expected %v, got %v", base, got)
	}
}

func TestNext_DoublesOnSubsequentCalls(t *testing.T) {
	b := New(base, max)
	wants := []time.Duration{base, 2 * base, 4 * base, 8 * base}
	for i, want := range wants {
		got := b.Next("key")
		if got != want {
			t.Fatalf("call %d: expected %v, got %v", i+1, want, got)
		}
	}
}

func TestNext_CapsAtMax(t *testing.T) {
	b := New(base, max)
	var last time.Duration
	for i := 0; i < 20; i++ {
		last = b.Next("key")
	}
	if last > max {
		t.Fatalf("delay %v exceeds max %v", last, max)
	}
}

func TestReset_ClearsState(t *testing.T) {
	b := New(base, max)
	b.Next("key")
	b.Next("key")
	b.Reset("key")
	got := b.Next("key")
	if got != base {
		t.Fatalf("after reset expected %v, got %v", base, got)
	}
}

func TestReset_UnknownKeyIsNoop(t *testing.T) {
	b := New(base, max)
	// Resetting a key that has never been used should not panic.
	b.Reset("nonexistent")
}

func TestAttempts_TracksFailures(t *testing.T) {
	b := New(base, max)
	if b.Attempts("key") != 0 {
		t.Fatal("expected 0 attempts before any call")
	}
	b.Next("key")
	b.Next("key")
	if got := b.Attempts("key"); got != 2 {
		t.Fatalf("expected 2 attempts, got %d", got)
	}
}

func TestAttempts_ResetsToCero(t *testing.T) {
	b := New(base, max)
	b.Next("key")
	b.Next("key")
	b.Reset("key")
	if got := b.Attempts("key"); got != 0 {
		t.Fatalf("expected 0 attempts after reset, got %d", got)
	}
}

func TestNext_IndependentKeys(t *testing.T) {
	b := New(base, max)
	b.Next("a")
	b.Next("a")
	gotB := b.Next("b")
	if gotB != base {
		t.Fatalf("key b should start at base %v, got %v", base, gotB)
	}
}
