package suppression

import (
	"testing"
	"time"
)

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero window")
		}
	}()
	New(0)
}

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	s := New(time.Minute)
	if !s.Allow("/etc/hosts") {
		t.Fatal("first call should always pass")
	}
}

func TestAllow_SecondCallWithinWindowBlocked(t *testing.T) {
	s := New(time.Minute)
	s.Allow("/etc/hosts")
	if s.Allow("/etc/hosts") {
		t.Fatal("second call within window should be blocked")
	}
}

func TestAllow_PassesAfterWindowExpires(t *testing.T) {
	s := New(50 * time.Millisecond)
	base := time.Now()
	s.now = func() time.Time { return base }
	s.Allow("/etc/hosts")

	s.now = func() time.Time { return base.Add(100 * time.Millisecond) }
	if !s.Allow("/etc/hosts") {
		t.Fatal("call after window expiry should pass")
	}
}

func TestAllow_IndependentPaths(t *testing.T) {
	s := New(time.Minute)
	s.Allow("/etc/hosts")
	if !s.Allow("/etc/resolv.conf") {
		t.Fatal("different path should not be suppressed")
	}
}

func TestReset_ClearsPath(t *testing.T) {
	s := New(time.Minute)
	s.Allow("/etc/hosts")
	s.Reset("/etc/hosts")
	if !s.Allow("/etc/hosts") {
		t.Fatal("after Reset, path should be allowed again")
	}
}

func TestResetAll_ClearsAllPaths(t *testing.T) {
	s := New(time.Minute)
	s.Allow("/etc/hosts")
	s.Allow("/etc/resolv.conf")
	s.ResetAll()
	if s.Len() != 0 {
		t.Fatalf("expected 0 tracked paths after ResetAll, got %d", s.Len())
	}
}

func TestLen_TracksCorrectly(t *testing.T) {
	s := New(time.Minute)
	if s.Len() != 0 {
		t.Fatal("expected 0 initially")
	}
	s.Allow("/a")
	s.Allow("/b")
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
}
