package escalation

import (
	"testing"
	"time"
)

const (
	warn = 5 * time.Minute
	crit = 30 * time.Minute
)

func newFixed(t time.Time) *Escalator {
	e := New(warn, crit)
	e.now = func() time.Time { return t }
	return e
}

func TestNew_PanicsOnZeroWarning(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero warning")
		}
	}()
	New(0, crit)
}

func TestNew_PanicsOnZeroCritical(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero critical")
		}
	}()
	New(warn, 0)
}

func TestNew_PanicsWhenWarningExceedsCritical(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when warning >= critical")
		}
	}()
	New(crit, warn)
}

func TestObserve_NoneOnFirstSeen(t *testing.T) {
	now := time.Now()
	e := newFixed(now)
	if got := e.Observe("/etc/hosts"); got != LevelNone {
		t.Fatalf("expected LevelNone, got %s", got)
	}
}

func TestObserve_WarningAfterThreshold(t *testing.T) {
	base := time.Now()
	e := New(warn, crit)
	e.now = func() time.Time { return base }
	e.Observe("/etc/hosts")

	e.now = func() time.Time { return base.Add(warn + time.Second) }
	if got := e.Observe("/etc/hosts"); got != LevelWarning {
		t.Fatalf("expected LevelWarning, got %s", got)
	}
}

func TestObserve_CriticalAfterThreshold(t *testing.T) {
	base := time.Now()
	e := New(warn, crit)
	e.now = func() time.Time { return base }
	e.Observe("/etc/hosts")

	e.now = func() time.Time { return base.Add(crit + time.Second) }
	if got := e.Observe("/etc/hosts"); got != LevelCritical {
		t.Fatalf("expected LevelCritical, got %s", got)
	}
}

func TestResolve_ClearsPath(t *testing.T) {
	base := time.Now()
	e := New(warn, crit)
	e.now = func() time.Time { return base }
	e.Observe("/etc/hosts")

	e.now = func() time.Time { return base.Add(crit + time.Second) }
	e.Resolve("/etc/hosts")

	// After resolve, first-seen resets so level should be None again.
	if got := e.Observe("/etc/hosts"); got != LevelNone {
		t.Fatalf("expected LevelNone after resolve, got %s", got)
	}
}

func TestActive_ReturnsPaths(t *testing.T) {
	now := time.Now()
	e := newFixed(now)
	e.Observe("/etc/hosts")
	e.Observe("/etc/resolv.conf")

	active := e.Active()
	if len(active) != 2 {
		t.Fatalf("expected 2 active paths, got %d", len(active))
	}
}

func TestLevel_String(t *testing.T) {
	for _, tc := range []struct{ l Level; want string }{
		{LevelNone, "none"},
		{LevelWarning, "warning"},
		{LevelCritical, "critical"},
	} {
		if got := tc.l.String(); got != tc.want {
			t.Errorf("Level(%d).String() = %q, want %q", tc.l, got, tc.want)
		}
	}
}
