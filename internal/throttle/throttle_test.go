package throttle_test

import (
	"testing"
	"time"

	"github.com/yourusername/driftwatch/internal/throttle"
)

func TestNew_PanicsOnZeroWindow(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero window")
		}
	}()
	throttle.New(0, 3)
}

func TestNew_PanicsOnZeroBurst(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on zero burst")
		}
	}()
	throttle.New(time.Second, 0)
}

func TestAllow_FirstCallAlwaysPasses(t *testing.T) {
	th := throttle.New(time.Second, 2)
	if !th.Allow("file.conf") {
		t.Fatal("expected first call to pass")
	}
}

func TestAllow_BurstLimitEnforced(t *testing.T) {
	th := throttle.New(time.Second, 3)
	key := "nginx.conf"

	for i := 0; i < 3; i++ {
		if !th.Allow(key) {
			t.Fatalf("expected call %d to pass", i+1)
		}
	}
	if th.Allow(key) {
		t.Fatal("expected 4th call to be blocked")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	th := throttle.New(time.Second, 1)

	if !th.Allow("a.conf") {
		t.Fatal("expected a.conf to pass")
	}
	if !th.Allow("b.conf") {
		t.Fatal("expected b.conf to pass independently")
	}
	if th.Allow("a.conf") {
		t.Fatal("expected second a.conf to be blocked")
	}
}

func TestAllow_ResetsAfterWindow(t *testing.T) {
	th := throttle.New(50*time.Millisecond, 1)
	key := "hosts"

	if !th.Allow(key) {
		t.Fatal("expected first call to pass")
	}
	if th.Allow(key) {
		t.Fatal("expected second call within window to be blocked")
	}
	time.Sleep(60 * time.Millisecond)
	if !th.Allow(key) {
		t.Fatal("expected call after window expiry to pass")
	}
}

func TestRemaining_DecreasesWithAllows(t *testing.T) {
	th := throttle.New(time.Second, 3)
	key := "sshd.conf"

	if got := th.Remaining(key); got != 3 {
		t.Fatalf("expected 3 remaining, got %d", got)
	}
	th.Allow(key)
	if got := th.Remaining(key); got != 2 {
		t.Fatalf("expected 2 remaining, got %d", got)
	}
}

func TestReset_ClearsState(t *testing.T) {
	th := throttle.New(time.Second, 1)
	key := "resolv.conf"

	th.Allow(key)
	if th.Allow(key) {
		t.Fatal("expected second call to be blocked before reset")
	}
	th.Reset(key)
	if !th.Allow(key) {
		t.Fatal("expected call to pass after reset")
	}
}
