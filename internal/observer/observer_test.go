package observer

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestNew_PanicsOnZeroWarning(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(0, time.Minute)
}

func TestNew_PanicsOnZeroCritical(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(time.Minute, 0)
}

func TestNew_PanicsWhenWarningExceedsCritical(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	New(time.Hour, time.Minute)
}

func TestObserve_NoneOnFirstSeen(t *testing.T) {
	base := time.Now()
	obs := New(5*time.Minute, 30*time.Minute)
	obs.now = fixedClock(base)

	sev := obs.Observe("/etc/hosts")
	if sev != SeverityNone {
		t.Fatalf("expected none, got %s", sev)
	}
}

func TestObserve_WarningSeverityAfterThreshold(t *testing.T) {
	base := time.Now()
	obs := New(5*time.Minute, 30*time.Minute)
	obs.now = fixedClock(base)
	obs.Observe("/etc/hosts")

	obs.now = fixedClock(base.Add(10 * time.Minute))
	sev := obs.Observe("/etc/hosts")
	if sev != SeverityWarning {
		t.Fatalf("expected warning, got %s", sev)
	}
}

func TestObserve_CriticalSeverityAfterThreshold(t *testing.T) {
	base := time.Now()
	obs := New(5*time.Minute, 30*time.Minute)
	obs.now = fixedClock(base)
	obs.Observe("/etc/hosts")

	obs.now = fixedClock(base.Add(45 * time.Minute))
	sev := obs.Observe("/etc/hosts")
	if sev != SeverityCritical {
		t.Fatalf("expected critical, got %s", sev)
	}
}

func TestResolve_ClearsTracking(t *testing.T) {
	base := time.Now()
	obs := New(5*time.Minute, 30*time.Minute)
	obs.now = fixedClock(base)
	obs.Observe("/etc/hosts")

	obs.Resolve("/etc/hosts")

	obs.now = fixedClock(base.Add(time.Hour))
	sev := obs.Observe("/etc/hosts")
	if sev != SeverityNone {
		t.Fatalf("expected none after resolve, got %s", sev)
	}
}

func TestAge_ZeroForUnknownPath(t *testing.T) {
	obs := New(time.Minute, time.Hour)
	if age := obs.Age("/unknown"); age != 0 {
		t.Fatalf("expected 0, got %v", age)
	}
}

func TestAge_ReturnsElapsedTime(t *testing.T) {
	base := time.Now()
	obs := New(time.Minute, time.Hour)
	obs.now = fixedClock(base)
	obs.Observe("/etc/hosts")

	obs.now = fixedClock(base.Add(7 * time.Minute))
	age := obs.Age("/etc/hosts")
	if age != 7*time.Minute {
		t.Fatalf("expected 7m, got %v", age)
	}
}

func TestSeverity_String(t *testing.T) {
	cases := []struct {
		sev  Severity
		want string
	}{
		{SeverityNone, "none"},
		{SeverityWarning, "warning"},
		{SeverityCritical, "critical"},
	}
	for _, tc := range cases {
		if got := tc.sev.String(); got != tc.want {
			t.Errorf("Severity(%d).String() = %q, want %q", tc.sev, got, tc.want)
		}
	}
}
