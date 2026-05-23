package dedup

import (
	"testing"
	"time"
)

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero window")
		}
	}()
	New(0)
}

func TestIsDuplicate_FirstCallReturnsFalse(t *testing.T) {
	d := New(time.Minute)
	if d.IsDuplicate("key1") {
		t.Error("first call should not be a duplicate")
	}
}

func TestIsDuplicate_SecondCallWithinWindowReturnsTrue(t *testing.T) {
	d := New(time.Minute)
	d.IsDuplicate("key1")
	if !d.IsDuplicate("key1") {
		t.Error("second call within window should be a duplicate")
	}
}

func TestIsDuplicate_PassesAfterWindowExpires(t *testing.T) {
	now := time.Now()
	d := New(time.Second)
	d.now = func() time.Time { return now }

	d.IsDuplicate("key1")

	// Advance time beyond the window.
	d.now = func() time.Time { return now.Add(2 * time.Second) }

	if d.IsDuplicate("key1") {
		t.Error("call after window expiry should not be a duplicate")
	}
}

func TestIsDuplicate_IndependentKeys(t *testing.T) {
	d := New(time.Minute)
	d.IsDuplicate("a")

	if d.IsDuplicate("b") {
		t.Error("different key should not be a duplicate")
	}
}

func TestForget_AllowsKeyToPassAgain(t *testing.T) {
	d := New(time.Minute)
	d.IsDuplicate("key1")
	d.Forget("key1")

	if d.IsDuplicate("key1") {
		t.Error("key should not be duplicate after Forget")
	}
}

func TestLen_TracksCount(t *testing.T) {
	d := New(time.Minute)
	if d.Len() != 0 {
		t.Fatalf("expected 0, got %d", d.Len())
	}
	d.IsDuplicate("a")
	d.IsDuplicate("b")
	if d.Len() != 2 {
		t.Fatalf("expected 2, got %d", d.Len())
	}
}

func TestLen_EvictsExpiredOnNextCall(t *testing.T) {
	now := time.Now()
	d := New(time.Second)
	d.now = func() time.Time { return now }
	d.IsDuplicate("x")

	d.now = func() time.Time { return now.Add(2 * time.Second) }
	d.IsDuplicate("y") // triggers eviction of "x"

	if d.Len() != 1 {
		t.Fatalf("expected 1 after eviction, got %d", d.Len())
	}
}
