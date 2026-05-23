package silencer

import (
	"testing"
	"time"
)

func fixedClock(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsSilenced_ReturnsFalseWhenUnset(t *testing.T) {
	s := New()
	if s.IsSilenced("/etc/hosts") {
		t.Fatal("expected false for unregistered path")
	}
}

func TestSilence_And_IsSilenced(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Silence("/etc/hosts", 10*time.Minute)

	if !s.IsSilenced("/etc/hosts") {
		t.Fatal("expected path to be silenced")
	}
}

func TestIsSilenced_ReturnsFalseAfterExpiry(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)
	s.Silence("/etc/hosts", 5*time.Minute)

	// Advance clock past expiry.
	s.now = fixedClock(now.Add(10 * time.Minute))

	if s.IsSilenced("/etc/hosts") {
		t.Fatal("expected silence to have expired")
	}
}

func TestLift_RemovesSilence(t *testing.T) {
	s := New()
	s.Silence("/etc/nginx/nginx.conf", time.Hour)
	s.Lift("/etc/nginx/nginx.conf")

	if s.IsSilenced("/etc/nginx/nginx.conf") {
		t.Fatal("expected silence to be lifted")
	}
}

func TestSilence_ZeroDurationIsNoOp(t *testing.T) {
	s := New()
	s.Silence("/etc/hosts", 0)

	if s.IsSilenced("/etc/hosts") {
		t.Fatal("zero-duration silence should not register")
	}
}

func TestActive_ReturnsOnlyLiveSilences(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Silence("/etc/hosts", 10*time.Minute)
	s.Silence("/etc/passwd", 1*time.Minute)

	// Advance past /etc/passwd expiry only.
	s.now = fixedClock(now.Add(2 * time.Minute))

	active := s.Active()
	if _, ok := active["/etc/hosts"]; !ok {
		t.Error("expected /etc/hosts to be in active silences")
	}
	if _, ok := active["/etc/passwd"]; ok {
		t.Error("expected /etc/passwd to have been evicted from active silences")
	}
}

func TestSilence_ExtendsExistingWindow(t *testing.T) {
	now := time.Now()
	s := New()
	s.now = fixedClock(now)

	s.Silence("/etc/hosts", 5*time.Minute)
	s.Silence("/etc/hosts", 30*time.Minute)

	// Advance past original 5-minute window.
	s.now = fixedClock(now.Add(10 * time.Minute))

	if !s.IsSilenced("/etc/hosts") {
		t.Fatal("expected extended silence to still be active")
	}
}
