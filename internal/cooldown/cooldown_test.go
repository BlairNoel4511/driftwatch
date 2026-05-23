package cooldown_test

import (
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/cooldown"
)

var t0 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestNew_PanicsOnZeroBase(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero base")
		}
	}()
	cooldown.New(0, time.Minute)
}

func TestNew_PanicsOnZeroMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero max")
		}
	}()
	cooldown.New(time.Second, 0)
}

func TestNew_PanicsWhenBaseExceedsMax(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when base > max")
		}
	}()
	cooldown.New(time.Minute, time.Second)
}

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	c := cooldown.New(time.Second, time.Minute)
	if !c.Allow("/etc/hosts", t0) {
		t.Fatal("expected first call to be allowed")
	}
}

func TestAllow_SecondCallWithinCooldownBlocked(t *testing.T) {
	c := cooldown.New(time.Second, time.Minute)
	c.Allow("/etc/hosts", t0)
	if c.Allow("/etc/hosts", t0.Add(500*time.Millisecond)) {
		t.Fatal("expected second call within cooldown to be blocked")
	}
}

func TestAllow_PassesAfterCooldownExpires(t *testing.T) {
	c := cooldown.New(time.Second, time.Minute)
	c.Allow("/etc/hosts", t0)
	if !c.Allow("/etc/hosts", t0.Add(2*time.Second)) {
		t.Fatal("expected call after cooldown to be allowed")
	}
}

func TestAllow_BackoffDoubles(t *testing.T) {
	c := cooldown.New(time.Second, time.Minute)
	c.Allow("/etc/hosts", t0)                      // window = 1s, ready at t0+1s
	c.Allow("/etc/hosts", t0.Add(2*time.Second))   // window = 2s, ready at t0+4s
	if c.Allow("/etc/hosts", t0.Add(3*time.Second)) {
		t.Fatal("expected call within doubled window to be blocked")
	}
	if !c.Allow("/etc/hosts", t0.Add(5*time.Second)) {
		t.Fatal("expected call after doubled window to pass")
	}
}

func TestAllow_BackoffCapsAtMax(t *testing.T) {
	base := time.Second
	max := 4 * time.Second
	c := cooldown.New(base, max)
	now := t0
	for i := 0; i < 10; i++ {
		c.Allow("/etc/hosts", now)
		now = now.Add(max + time.Second)
	}
	// after many doublings the window is capped; a call just under max should be blocked
	c.Allow("/etc/hosts", now)
	if c.Allow("/etc/hosts", now.Add(max-time.Millisecond)) {
		t.Fatal("expected call just before max cap to be blocked")
	}
}

func TestAllow_IndependentPaths(t *testing.T) {
	c := cooldown.New(time.Second, time.Minute)
	c.Allow("/etc/hosts", t0)
	if !c.Allow("/etc/passwd", t0) {
		t.Fatal("expected independent path to be allowed")
	}
}

func TestReset_ClearsBackoff(t *testing.T) {
	c := cooldown.New(time.Second, time.Minute)
	c.Allow("/etc/hosts", t0)
	c.Reset("/etc/hosts")
	if !c.Allow("/etc/hosts", t0) {
		t.Fatal("expected allow after reset")
	}
}

func TestResetAll_ClearsAllPaths(t *testing.T) {
	c := cooldown.New(time.Second, time.Minute)
	c.Allow("/etc/hosts", t0)
	c.Allow("/etc/passwd", t0)
	c.ResetAll()
	if !c.Allow("/etc/hosts", t0) || !c.Allow("/etc/passwd", t0) {
		t.Fatal("expected all paths allowed after ResetAll")
	}
}
