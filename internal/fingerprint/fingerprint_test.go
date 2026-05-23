package fingerprint_test

import (
	"os"
	"testing"
	"time"

	"github.com/driftwatch/internal/fingerprint"
	"github.com/driftwatch/internal/watcher"
)

func baseSnap(path string) watcher.Snapshot {
	return watcher.Snapshot{
		Path:     path,
		Size:     1024,
		Mode:     0o644,
		ModTime:  time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		Checksum: "abc123",
	}
}

func TestCompute_IsStable(t *testing.T) {
	fp := fingerprint.New()
	snap := baseSnap("/etc/hosts")
	a := fp.Compute(snap)
	b := fp.Compute(snap)
	if a != b {
		t.Fatalf("expected stable fingerprint, got %q then %q", a, b)
	}
}

func TestCompute_DiffersOnSizeChange(t *testing.T) {
	fp := fingerprint.New()
	s1 := baseSnap("/etc/hosts")
	s2 := baseSnap("/etc/hosts")
	s2.Size = 2048
	if fp.Compute(s1) == fp.Compute(s2) {
		t.Fatal("expected different fingerprints for different sizes")
	}
}

func TestCompute_DiffersOnModeChange(t *testing.T) {
	fp := fingerprint.New()
	s1 := baseSnap("/etc/passwd")
	s2 := baseSnap("/etc/passwd")
	s2.Mode = os.FileMode(0o600)
	if fp.Compute(s1) == fp.Compute(s2) {
		t.Fatal("expected different fingerprints for different modes")
	}
}

func TestChanged_FirstObservationIsNotChanged(t *testing.T) {
	fp := fingerprint.New()
	snap := baseSnap("/etc/hosts")
	if fp.Changed(snap) {
		t.Fatal("first observation should never report changed")
	}
}

func TestChanged_DetectsDrift(t *testing.T) {
	fp := fingerprint.New()
	snap := baseSnap("/etc/hosts")
	fp.Changed(snap) // seed cache

	modified := snap
	modified.Checksum = "deadbeef"
	if !fp.Changed(modified) {
		t.Fatal("expected Changed to return true after checksum change")
	}
}

func TestEvict_ForcesFirstObservation(t *testing.T) {
	fp := fingerprint.New()
	snap := baseSnap("/etc/hosts")
	fp.Changed(snap) // seed

	fp.Evict(snap.Path)

	// After eviction the same snap should look like a first observation.
	if fp.Changed(snap) {
		t.Fatal("after eviction, same snap should not report changed")
	}
}
