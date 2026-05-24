package circuitbreaker

import (
	"testing"
	"time"
)

func TestNew_PanicsOnZeroThreshold(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0, time.Second)
}

func TestNew_PanicsOnZeroCooldown(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(1, 0)
}

func TestAllow_ClosedByDefault(t *testing.T) {
	b := New(3, time.Second)
	if err := b.Allow("/etc/hosts"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestAllow_OpensAfterThreshold(t *testing.T) {
	b := New(2, time.Second)
	b.RecordFailure("/etc/hosts")
	b.RecordFailure("/etc/hosts")

	if err := b.Allow("/etc/hosts"); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}
}

func TestAllow_RemainsClosedBelowThreshold(t *testing.T) {
	b := New(3, time.Second)
	b.RecordFailure("/etc/hosts")
	b.RecordFailure("/etc/hosts")

	if err := b.Allow("/etc/hosts"); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
}

func TestRecordSuccess_ClosesCircuit(t *testing.T) {
	b := New(1, time.Second)
	b.RecordFailure("/etc/hosts")
	if b.StateOf("/etc/hosts") != StateOpen {
		t.Fatal("expected open")
	}

	// Simulate cooldown elapsed so half-open probe is allowed.
	now := time.Now().Add(2 * time.Second)
	b.now = func() time.Time { return now }
	_ = b.Allow("/etc/hosts") // transitions to half-open
	b.RecordSuccess("/etc/hosts")

	if b.StateOf("/etc/hosts") != StateClosed {
		t.Fatalf("expected closed, got %s", b.StateOf("/etc/hosts"))
	}
}

func TestAllow_HalfOpenAfterCooldown(t *testing.T) {
	now := time.Now()
	b := New(1, time.Second)
	b.now = func() time.Time { return now }

	b.RecordFailure("/etc/hosts")
	if err := b.Allow("/etc/hosts"); err != ErrOpen {
		t.Fatalf("expected ErrOpen, got %v", err)
	}

	b.now = func() time.Time { return now.Add(2 * time.Second) }
	if err := b.Allow("/etc/hosts"); err != nil {
		t.Fatalf("expected nil in half-open, got %v", err)
	}
	if b.StateOf("/etc/hosts") != StateHalfOpen {
		t.Fatalf("expected half-open, got %s", b.StateOf("/etc/hosts"))
	}
}

func TestStateOf_IndependentPaths(t *testing.T) {
	b := New(2, time.Second)
	b.RecordFailure("/etc/hosts")
	b.RecordFailure("/etc/hosts")

	if b.StateOf("/etc/resolv.conf") != StateClosed {
		t.Fatal("/etc/resolv.conf should be closed")
	}
}

func TestStateString(t *testing.T) {
	cases := []struct {
		s    State
		want string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(99), "unknown"},
	}
	for _, tc := range cases {
		if got := tc.s.String(); got != tc.want {
			t.Errorf("State(%d).String() = %q, want %q", tc.s, got, tc.want)
		}
	}
}
