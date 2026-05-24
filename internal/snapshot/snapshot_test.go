package snapshot_test

import (
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/snapshot"
)

func makeEntry(path string) snapshot.Entry {
	return snapshot.Entry{
		Path:    path,
		Size:    1024,
		Mode:    0o644,
		ModTime: time.Now(),
		Missing: false,
	}
}

func TestSet_And_Get_RoundTrip(t *testing.T) {
	s := snapshot.New()
	e := makeEntry("/etc/hosts")
	if err := s.Set(e); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, ok := s.Get("/etc/hosts")
	if !ok {
		t.Fatal("expected entry to be present")
	}
	if got.Size != e.Size {
		t.Errorf("size: got %d, want %d", got.Size, e.Size)
	}
	if got.Updated.IsZero() {
		t.Error("Updated timestamp should be set")
	}
}

func TestGet_ReturnsFalseWhenUnset(t *testing.T) {
	s := snapshot.New()
	_, ok := s.Get("/nonexistent")
	if ok {
		t.Fatal("expected false for unknown path")
	}
}

func TestSet_EmptyPathReturnsError(t *testing.T) {
	s := snapshot.New()
	err := s.Set(snapshot.Entry{})
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestDelete_RemovesEntry(t *testing.T) {
	s := snapshot.New()
	_ = s.Set(makeEntry("/etc/resolv.conf"))
	s.Delete("/etc/resolv.conf")
	_, ok := s.Get("/etc/resolv.conf")
	if ok {
		t.Fatal("expected entry to be removed")
	}
}

func TestAll_ReturnsAllEntries(t *testing.T) {
	s := snapshot.New()
	paths := []string{"/a", "/b", "/c"}
	for _, p := range paths {
		_ = s.Set(makeEntry(p))
	}
	all := s.All()
	if len(all) != len(paths) {
		t.Errorf("All: got %d entries, want %d", len(all), len(paths))
	}
}

func TestLen_ReflectsCurrentCount(t *testing.T) {
	s := snapshot.New()
	if s.Len() != 0 {
		t.Fatal("expected empty store")
	}
	_ = s.Set(makeEntry("/x"))
	_ = s.Set(makeEntry("/y"))
	if s.Len() != 2 {
		t.Errorf("Len: got %d, want 2", s.Len())
	}
	s.Delete("/x")
	if s.Len() != 1 {
		t.Errorf("Len after delete: got %d, want 1", s.Len())
	}
}
