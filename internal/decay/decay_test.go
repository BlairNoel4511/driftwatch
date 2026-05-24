package decay

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroRate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0)
}

func TestNew_PanicsOnNegativeRate(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(-1)
}

func TestScore_ZeroForUnknownPath(t *testing.T) {
	d := New(1.0)
	if got := d.Score("/etc/hosts"); got != 0 {
		t.Fatalf("expected 0, got %f", got)
	}
}

func TestAdd_IncreasesScore(t *testing.T) {
	now := time.Now()
	d := New(1.0)
	d.now = fixedClock(now)
	d.Add("/etc/hosts", 5.0)
	if got := d.Score("/etc/hosts"); got != 5.0 {
		t.Fatalf("expected 5.0, got %f", got)
	}
}

func TestScore_DecaysOverTime(t *testing.T) {
	base := time.Now()
	d := New(2.0) // 2 points/sec
	d.now = fixedClock(base)
	d.Add("/etc/hosts", 10.0)

	// Advance 3 seconds — should lose 6 points.
	d.now = fixedClock(base.Add(3 * time.Second))
	got := d.Score("/etc/hosts")
	if got != 4.0 {
		t.Fatalf("expected 4.0 after decay, got %f", got)
	}
}

func TestScore_FloorIsZero(t *testing.T) {
	base := time.Now()
	d := New(5.0)
	d.now = fixedClock(base)
	d.Add("/etc/passwd", 2.0)

	// Advance 10 seconds — score would go negative without floor.
	d.now = fixedClock(base.Add(10 * time.Second))
	got := d.Score("/etc/passwd")
	if got != 0 {
		t.Fatalf("expected 0 (floor), got %f", got)
	}
}

func TestReset_ClearsScore(t *testing.T) {
	now := time.Now()
	d := New(1.0)
	d.now = fixedClock(now)
	d.Add("/etc/hosts", 8.0)
	d.Reset("/etc/hosts")
	if got := d.Score("/etc/hosts"); got != 0 {
		t.Fatalf("expected 0 after reset, got %f", got)
	}
}

func TestAdd_IndependentPaths(t *testing.T) {
	now := time.Now()
	d := New(1.0)
	d.now = fixedClock(now)
	d.Add("/etc/hosts", 3.0)
	d.Add("/etc/passwd", 7.0)

	if got := d.Score("/etc/hosts"); got != 3.0 {
		t.Fatalf("hosts: expected 3.0, got %f", got)
	}
	if got := d.Score("/etc/passwd"); got != 7.0 {
		t.Fatalf("passwd: expected 7.0, got %f", got)
	}
}
