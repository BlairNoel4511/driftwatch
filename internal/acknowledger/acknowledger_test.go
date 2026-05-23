package acknowledger

import (
	"testing"
	"time"
)

func fixedNow(t time.Time) func() time.Time {
	return func() time.Time { return t }
}

func TestIsAcknowledged_ReturnsFalseWhenUnset(t *testing.T) {
	a := New()
	if a.IsAcknowledged("/etc/hosts") {
		t.Fatal("expected false for unacknowledged path")
	}
}

func TestAcknowledge_And_IsAcknowledged(t *testing.T) {
	a := New()
	base := time.Now()
	a.now = fixedNow(base)
	a.Acknowledge("/etc/hosts", "alice", time.Hour)
	if !a.IsAcknowledged("/etc/hosts") {
		t.Fatal("expected path to be acknowledged")
	}
}

func TestIsAcknowledged_ReturnsFalseAfterExpiry(t *testing.T) {
	a := New()
	base := time.Now()
	a.now = fixedNow(base)
	a.Acknowledge("/etc/hosts", "alice", time.Second)
	// advance time past expiry
	a.now = fixedNow(base.Add(2 * time.Second))
	if a.IsAcknowledged("/etc/hosts") {
		t.Fatal("expected acknowledgement to be expired")
	}
}

func TestRevoke_RemovesAcknowledgement(t *testing.T) {
	a := New()
	a.Acknowledge("/etc/hosts", "bob", time.Hour)
	a.Revoke("/etc/hosts")
	if a.IsAcknowledged("/etc/hosts") {
		t.Fatal("expected acknowledgement to be revoked")
	}
}

func TestGet_ReturnsEntry(t *testing.T) {
	a := New()
	base := time.Now()
	a.now = fixedNow(base)
	a.Acknowledge("/etc/passwd", "carol", time.Minute)
	e, ok := a.Get("/etc/passwd")
	if !ok {
		t.Fatal("expected entry to exist")
	}
	if e.AckedBy != "carol" {
		t.Fatalf("expected AckedBy=carol, got %s", e.AckedBy)
	}
	if e.Path != "/etc/passwd" {
		t.Fatalf("expected Path=/etc/passwd, got %s", e.Path)
	}
}

func TestPurge_RemovesExpiredEntries(t *testing.T) {
	a := New()
	base := time.Now()
	a.now = fixedNow(base)
	a.Acknowledge("/etc/hosts", "alice", time.Second)
	a.Acknowledge("/etc/passwd", "bob", time.Hour)
	// advance past first expiry only
	a.now = fixedNow(base.Add(2 * time.Second))
	a.Purge()
	if a.IsAcknowledged("/etc/hosts") {
		t.Fatal("expected expired entry to be purged")
	}
	if !a.IsAcknowledged("/etc/passwd") {
		t.Fatal("expected non-expired entry to survive purge")
	}
}
