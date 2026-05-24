package window

import (
	"testing"
	"time"
)

func TestNew_PanicsOnZeroDuration(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero duration")
		}
	}()
	New(0)
}

func TestCount_EmptyWindow(t *testing.T) {
	w := New(time.Minute)
	if got := w.Count("x"); got != 0 {
		t.Fatalf("expected 0, got %d", got)
	}
}

func TestRecord_And_Count(t *testing.T) {
	w := New(time.Minute)
	w.Record("/etc/hosts")
	w.Record("/etc/hosts")
	if got := w.Count("/etc/hosts"); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}
}

func TestCount_ExpiresOldEvents(t *testing.T) {
	base := time.Now()
	w := New(time.Second)

	// inject a controlled clock
	w.now = func() time.Time { return base }
	w.Record("f")
	w.Record("f")

	// advance clock past the window
	w.now = func() time.Time { return base.Add(2 * time.Second) }
	if got := w.Count("f"); got != 0 {
		t.Fatalf("expected 0 after expiry, got %d", got)
	}
}

func TestCount_PartialExpiry(t *testing.T) {
	base := time.Now()
	w := New(time.Minute)

	w.now = func() time.Time { return base }
	w.Record("p")

	w.now = func() time.Time { return base.Add(30 * time.Second) }
	w.Record("p")

	// advance 70 s from base: first event is outside window, second is inside
	w.now = func() time.Time { return base.Add(70 * time.Second) }
	if got := w.Count("p"); got != 1 {
		t.Fatalf("expected 1, got %d", got)
	}
}

func TestReset_ClearsEvents(t *testing.T) {
	w := New(time.Minute)
	w.Record("r")
	w.Record("r")
	w.Reset("r")
	if got := w.Count("r"); got != 0 {
		t.Fatalf("expected 0 after reset, got %d", got)
	}
}

func TestCount_IndependentKeys(t *testing.T) {
	w := New(time.Minute)
	w.Record("a")
	w.Record("a")
	w.Record("b")

	if got := w.Count("a"); got != 2 {
		t.Fatalf("a: expected 2, got %d", got)
	}
	if got := w.Count("b"); got != 1 {
		t.Fatalf("b: expected 1, got %d", got)
	}
}
