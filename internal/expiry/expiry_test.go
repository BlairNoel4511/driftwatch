package expiry

import (
	"testing"
	"time"
)

// fixedClock returns a Clock whose current time can be advanced manually.
func fixedClock(initial time.Time) (Clock, func(d time.Duration)) {
	current := initial
	clock := func() time.Time { return current }
	advance := func(d time.Duration) { current = current.Add(d) }
	return clock, advance
}

func TestIsExpired_ReturnsTrueWhenUnset(t *testing.T) {
	e := New(nil)
	if !e.IsExpired("/etc/hosts") {
		t.Fatal("expected unset path to be expired")
	}
}

func TestSet_And_IsExpired_NotYetExpired(t *testing.T) {
	clock, _ := fixedClock(time.Unix(1000, 0))
	e := New(clock)

	if err := e.Set("/etc/hosts", 5*time.Second); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if e.IsExpired("/etc/hosts") {
		t.Fatal("expected path to not be expired yet")
	}
}

func TestSet_And_IsExpired_AfterTTL(t *testing.T) {
	clock, advance := fixedClock(time.Unix(1000, 0))
	e := New(clock)

	_ = e.Set("/etc/hosts", 5*time.Second)
	advance(6 * time.Second)

	if !e.IsExpired("/etc/hosts") {
		t.Fatal("expected path to be expired after TTL")
	}
}

func TestSet_EmptyPathReturnsError(t *testing.T) {
	e := New(nil)
	if err := e.Set("", time.Second); err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestSet_NonPositiveTTLReturnsError(t *testing.T) {
	e := New(nil)
	if err := e.Set("/etc/hosts", 0); err == nil {
		t.Fatal("expected error for zero TTL")
	}
	if err := e.Set("/etc/hosts", -time.Second); err == nil {
		t.Fatal("expected error for negative TTL")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	e := New(nil)
	_ = e.Set("/etc/hosts", time.Hour)
	e.Delete("/etc/hosts")
	if !e.IsExpired("/etc/hosts") {
		t.Fatal("expected deleted path to report as expired")
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	clock, advance := fixedClock(time.Unix(1000, 0))
	e := New(clock)

	_ = e.Set("/etc/hosts", 2*time.Second)
	_ = e.Set("/etc/passwd", time.Hour)

	advance(3 * time.Second)
	e.Purge()

	if e.Len() != 1 {
		t.Fatalf("expected 1 entry after purge, got %d", e.Len())
	}
	if e.IsExpired("/etc/passwd") {
		t.Fatal("expected /etc/passwd to still be valid after purge")
	}
}

func TestLen_ReflectsTrackedEntries(t *testing.T) {
	e := New(nil)
	if e.Len() != 0 {
		t.Fatalf("expected 0, got %d", e.Len())
	}
	_ = e.Set("/a", time.Second)
	_ = e.Set("/b", time.Second)
	if e.Len() != 2 {
		t.Fatalf("expected 2, got %d", e.Len())
	}
}
